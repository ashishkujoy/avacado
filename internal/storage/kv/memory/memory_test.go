package memory

import (
	"avacado/internal/storage/kv"
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestKVMemoryStore_GetAndSet(t *testing.T) {
	store := NewKVMemoryStore()
	v, err := store.Get(context.Background(), "key1")

	assert.NoError(t, err)
	assert.Nil(t, v)
	options := kv.NewSetOptions()

	err = store.Set(context.Background(), "key1", []byte("value1"), options)

	assert.NoError(t, err)

	v, err = store.Get(context.Background(), "key1")

	assert.NoError(t, err)
	assert.Equal(t, "value1", string(v))
}

func TestKVMemoryStore_SetExistingKeyWithNXOptionEnabled(t *testing.T) {
	store := NewKVMemoryStore()
	options := kv.NewSetOptions()
	options.WithNX()

	err := store.Set(context.Background(), "key1", []byte("value1"), options)
	assert.NoError(t, err)

	err = store.Set(context.Background(), "key1", []byte("value2"), options)
	assert.Error(t, err, NewKeyAlreadyExistsError("key1"))
}

func TestKVMemoryStore_SetExistingKeyWithNXOptionDisabled(t *testing.T) {
	store := NewKVMemoryStore()
	option := kv.NewSetOptions()

	err := store.Set(context.Background(), "key1", []byte("value1"), option)
	assert.NoError(t, err)

	err = store.Set(context.Background(), "key1", []byte("value2"), option)
	assert.NoError(t, err)

	value, _ := store.Get(context.Background(), "key1")
	assert.Equal(t, "value2", string(value))
}

func TestKVMemoryStore_SetWithXXEnabled(t *testing.T) {
	store := NewKVMemoryStore()
	optionWithXX := kv.NewSetOptions()
	optionWithXX.WithXX()

	err := store.Set(context.Background(), "key1", []byte("value1"), optionWithXX)
	assert.Error(t, err)

	err = store.Set(context.Background(), "key1", []byte("value2"), kv.NewSetOptions())
	assert.NoError(t, err)

	err = store.Set(context.Background(), "key1", []byte("value3"), optionWithXX)
	assert.NoError(t, err)

	value, _ := store.Get(context.Background(), "key1")
	assert.Equal(t, "value3", string(value))
}

func TestKVMemoryStore_Expiry(t *testing.T) {
	store := NewKVMemoryStore()
	defer store.Close()
	options := kv.NewSetOptions()
	options.WithEX(1)

	err := store.Set(context.Background(), "key1", []byte("value1"), options)
	assert.NoError(t, err)

	v, _ := store.Get(context.Background(), "key1")
	assert.NotNil(t, v)

	<-time.Tick(1200 * time.Millisecond)
	v, _ = store.Get(context.Background(), "key1")
	assert.Nil(t, v)
}

// TestKVMemoryStore_ConcurrentGetExpiredKey verifies no race condition when multiple
// goroutines try to GET an expired key simultaneously
func TestKVMemoryStore_ConcurrentGetExpiredKey(t *testing.T) {
	store := NewKVMemoryStore()
	defer store.Close()
	options := kv.NewSetOptions().WithEX(1)

	// Set a key that will expire
	err := store.Set(context.Background(), "key1", []byte("value1"), options)
	assert.NoError(t, err)

	// Wait for expiry
	time.Sleep(1100 * time.Millisecond)

	// Concurrent GET operations on expired key
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			v, err := store.Get(context.Background(), "key1")
			assert.NoError(t, err)
			assert.Nil(t, v)
		}()
	}

	wg.Wait()

	// Verify key was deleted
	v, _ := store.Get(context.Background(), "key1")
	assert.Nil(t, v)
}

// TestKVMemoryStore_ConcurrentReadWrite verifies concurrent GET and SET operations
// work correctly without race conditions
func TestKVMemoryStore_ConcurrentReadWrite(t *testing.T) {
	store := NewKVMemoryStore()
	defer store.Close()

	var wg sync.WaitGroup

	// Multiple writers
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			key := "key1"
			value := []byte("value")
			options := kv.NewSetOptions()
			err := store.Set(context.Background(), key, value, options)
			assert.NoError(t, err)
		}(i)
	}

	// Multiple readers
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := store.Get(context.Background(), "key1")
			assert.NoError(t, err)
		}()
	}

	wg.Wait()
}

// TestKVMemoryStore_BackgroundCleanup verifies that background cleanup removes expired keys
func TestKVMemoryStore_BackgroundCleanup(t *testing.T) {
	store := NewKVMemoryStore()
	defer store.Close()

	// Set multiple keys with short expiry
	for i := 0; i < 10; i++ {
		options := kv.NewSetOptions().WithEX(1)
		err := store.Set(context.Background(), "key"+string(rune('0'+i)), []byte("value"), options)
		assert.NoError(t, err)
	}

	// Wait for expiry + background cleanup
	time.Sleep(2500 * time.Millisecond)

	// Try to get all keys - they should all be nil (cleaned up)
	for i := 0; i < 10; i++ {
		v, err := store.Get(context.Background(), "key"+string(rune('0'+i)))
		assert.NoError(t, err)
		assert.Nil(t, v, "Key should have been cleaned up by background process")
	}
}

// TestKVMemoryStore_Close verifies graceful shutdown
func TestKVMemoryStore_Close(t *testing.T) {
	store := NewKVMemoryStore()

	// Set some keys
	options := kv.NewSetOptions()
	err := store.Set(context.Background(), "key1", []byte("value1"), options)
	assert.NoError(t, err)

	// Close the store
	err = store.Close()
	assert.NoError(t, err)

	// Background cleanup should have stopped
	// Note: We can't easily verify the goroutine stopped, but no panic is good
}

// TestKVMemoryStore_LazyExpirationImmediateCleanup verifies that lazy expiration
// immediately removes expired keys on GET
func TestKVMemoryStore_LazyExpirationImmediateCleanup(t *testing.T) {
	store := NewKVMemoryStore()
	defer store.Close()
	options := kv.NewSetOptions().WithEX(1)

	// Set a key
	err := store.Set(context.Background(), "key1", []byte("value1"), options)
	assert.NoError(t, err)

	// Wait for expiry
	time.Sleep(1100 * time.Millisecond)

	// First GET should trigger deletion
	v, err := store.Get(context.Background(), "key1")
	assert.NoError(t, err)
	assert.Nil(t, v)

	// Second GET should also return nil (key already deleted)
	v, err = store.Get(context.Background(), "key1")
	assert.NoError(t, err)
	assert.Nil(t, v)
}

// TestKVMemoryStore_NonExpiredKeysDuringConcurrentAccess verifies that non-expired
// keys are correctly returned during concurrent access
func TestKVMemoryStore_NonExpiredKeysDuringConcurrentAccess(t *testing.T) {
	store := NewKVMemoryStore()
	defer store.Close()
	options := kv.NewSetOptions().WithEX(10) // Long expiry

	err := store.Set(context.Background(), "key1", []byte("value1"), options)
	assert.NoError(t, err)

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			v, err := store.Get(context.Background(), "key1")
			assert.NoError(t, err)
			assert.Equal(t, []byte("value1"), v)
		}()
	}

	wg.Wait()
}
