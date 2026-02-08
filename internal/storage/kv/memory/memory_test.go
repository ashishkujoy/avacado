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

	_, err = store.Set(context.Background(), "key1", []byte("value1"), options)

	assert.NoError(t, err)

	v, err = store.Get(context.Background(), "key1")

	assert.NoError(t, err)
	assert.Equal(t, "value1", string(v))
}

func TestKVMemoryStore_SetExistingKeyWithNXOptionEnabled(t *testing.T) {
	store := NewKVMemoryStore()
	options := kv.NewSetOptions()
	options.WithNX()

	_, err := store.Set(context.Background(), "key1", []byte("value1"), options)
	assert.NoError(t, err)

	_, err = store.Set(context.Background(), "key1", []byte("value2"), options)
	assert.Error(t, err, NewKeyAlreadyExistsError("key1"))
}

func TestKVMemoryStore_SetExistingKeyWithNXOptionDisabled(t *testing.T) {
	store := NewKVMemoryStore()
	option := kv.NewSetOptions()

	_, err := store.Set(context.Background(), "key1", []byte("value1"), option)
	assert.NoError(t, err)

	_, err = store.Set(context.Background(), "key1", []byte("value2"), option)
	assert.NoError(t, err)

	value, _ := store.Get(context.Background(), "key1")
	assert.Equal(t, "value2", string(value))
}

func TestKVMemoryStore_SetWithXXEnabled(t *testing.T) {
	store := NewKVMemoryStore()
	optionWithXX := kv.NewSetOptions()
	optionWithXX.WithXX()

	_, err := store.Set(context.Background(), "key1", []byte("value1"), optionWithXX)
	assert.Error(t, err)

	_, err = store.Set(context.Background(), "key1", []byte("value2"), kv.NewSetOptions())
	assert.NoError(t, err)

	_, err = store.Set(context.Background(), "key1", []byte("value3"), optionWithXX)
	assert.NoError(t, err)

	value, _ := store.Get(context.Background(), "key1")
	assert.Equal(t, "value3", string(value))
}

func TestKVMemoryStore_Expiry(t *testing.T) {
	store := NewKVMemoryStore()
	defer store.Close()
	options := kv.NewSetOptions()
	options.WithEX(1)

	_, err := store.Set(context.Background(), "key1", []byte("value1"), options)
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
	_, err := store.Set(context.Background(), "key1", []byte("value1"), options)
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
			_, err := store.Set(context.Background(), key, value, options)
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
		_, err := store.Set(context.Background(), "key"+string(rune('0'+i)), []byte("value"), options)
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
	_, err := store.Set(context.Background(), "key1", []byte("value1"), options)
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
	_, err := store.Set(context.Background(), "key1", []byte("value1"), options)
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

	_, err := store.Set(context.Background(), "key1", []byte("value1"), options)
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

func TestKVMemoryStore_Incr(t *testing.T) {
	store := NewKVMemoryStore()
	defer store.Close()
	ctx := context.Background()

	// Increment non-existing key should initialize it to 1
	val, err := store.Incr(ctx, "counter")
	assert.NoError(t, err)
	assert.Equal(t, int64(1), val)

	_, err = store.Set(ctx, "counter1", []byte("10"), kv.NewSetOptions())
	// Increment existing key should increment it
	val, err = store.Incr(ctx, "counter1")
	assert.NoError(t, err)
	assert.Equal(t, int64(11), val)

	_, err = store.Set(ctx, "counter2", []byte("20"), kv.NewSetOptions().WithEX(1))
	pastTime := time.Now().Add(-2 * time.Second)
	store.store["counter2"].expiry = &pastTime

	v, err := store.Incr(ctx, "counter2")
	assert.NoError(t, err)
	assert.Equal(t, int64(1), v)

	// Incrementing a key with non-integer value should return an error
	_, err = store.Set(context.Background(), "counter3", []byte("not-an-integer"), kv.NewSetOptions())
	assert.NoError(t, err)

	_, err = store.Incr(ctx, "counter3")
	assert.Error(t, err)
}

func TestKVMemoryStore_Decr(t *testing.T) {
	store := NewKVMemoryStore()
	defer store.Close()
	ctx := context.Background()

	// Decrement non-existing key should initialize it to -1
	val, err := store.Decr(ctx, "counter")
	assert.NoError(t, err)
	assert.Equal(t, int64(-1), val)

	_, err = store.Set(ctx, "counter1", []byte("10"), kv.NewSetOptions())
	// Decrement existing key should decrement it
	val, err = store.Decr(ctx, "counter1")
	assert.NoError(t, err)
	assert.Equal(t, int64(9), val)

	_, err = store.Set(ctx, "counter2", []byte("20"), kv.NewSetOptions().WithEX(1))
	pastTime := time.Now().Add(-2 * time.Second)
	store.store["counter2"].expiry = &pastTime

	v, err := store.Decr(ctx, "counter2")
	assert.NoError(t, err)
	assert.Equal(t, int64(-1), v)

	// Decrementing a key with non-integer value should return an error
	_, err = store.Set(context.Background(), "counter3", []byte("not-an-integer"), kv.NewSetOptions())
	assert.NoError(t, err)

	_, err = store.Decr(ctx, "counter3")
	assert.Error(t, err)
}

func TestKVMemoryStore_DecrBy(t *testing.T) {
	store := NewKVMemoryStore()
	defer store.Close()
	ctx := context.Background()

	// Decrement non-existing key should initialize it to 0-decrement
	val, err := store.DecrBy(ctx, "counter", 5)
	assert.NoError(t, err)
	assert.Equal(t, int64(-5), val)

	// Decrement it again
	val, err = store.DecrBy(ctx, "counter", 3)
	assert.NoError(t, err)
	assert.Equal(t, int64(-8), val)

	_, err = store.Set(ctx, "counter1", []byte("100"), kv.NewSetOptions())
	// Decrement existing key by specified amount
	val, err = store.DecrBy(ctx, "counter1", 20)
	assert.NoError(t, err)
	assert.Equal(t, int64(80), val)

	// Decrement again
	val, err = store.DecrBy(ctx, "counter1", 30)
	assert.NoError(t, err)
	assert.Equal(t, int64(50), val)

	_, err = store.Set(ctx, "counter2", []byte("50"), kv.NewSetOptions().WithEX(1))
	pastTime := time.Now().Add(-2 * time.Second)
	store.store["counter2"].expiry = &pastTime

	// Decrement expired key should treat it as non-existent
	v, err := store.DecrBy(ctx, "counter2", 15)
	assert.NoError(t, err)
	assert.Equal(t, int64(-15), v)

	// Decrementing a key with non-integer value should return an error
	_, err = store.Set(context.Background(), "counter3", []byte("not-an-integer"), kv.NewSetOptions())
	assert.NoError(t, err)

	_, err = store.DecrBy(ctx, "counter3", 10)
	assert.Error(t, err)
}

func TestKVMemoryStore_Del(t *testing.T) {
	store := NewKVMemoryStore()
	defer store.Close()
	ctx := context.Background()

	// Delete non-existing key should return 0
	count, err := store.Del(ctx, "nonexistent")
	assert.NoError(t, err)
	assert.Equal(t, int64(0), count)

	// Set some keys
	_, err = store.Set(ctx, "key1", []byte("value1"), kv.NewSetOptions())
	assert.NoError(t, err)
	_, err = store.Set(ctx, "key2", []byte("value2"), kv.NewSetOptions())
	assert.NoError(t, err)
	_, err = store.Set(ctx, "key3", []byte("value3"), kv.NewSetOptions())
	assert.NoError(t, err)

	// Delete single existing key
	count, err = store.Del(ctx, "key1")
	assert.NoError(t, err)
	assert.Equal(t, int64(1), count)

	// Verify key is deleted
	val, err := store.Get(ctx, "key1")
	assert.NoError(t, err)
	assert.Nil(t, val)

	// Delete multiple existing keys
	count, err = store.Del(ctx, "key2", "key3")
	assert.NoError(t, err)
	assert.Equal(t, int64(2), count)

	// Verify keys are deleted
	val, err = store.Get(ctx, "key2")
	assert.NoError(t, err)
	assert.Nil(t, val)
	val, err = store.Get(ctx, "key3")
	assert.NoError(t, err)
	assert.Nil(t, val)

	// Delete mix of existing and non-existing keys
	_, err = store.Set(ctx, "key4", []byte("value4"), kv.NewSetOptions())
	assert.NoError(t, err)
	count, err = store.Del(ctx, "key4", "nonexistent", "alsonothere")
	assert.NoError(t, err)
	assert.Equal(t, int64(1), count) // Only key4 was deleted

	// Delete expired key should return 0
	_, err = store.Set(ctx, "expiring", []byte("value"), kv.NewSetOptions().WithEX(1))
	assert.NoError(t, err)
	pastTime := time.Now().Add(-2 * time.Second)
	store.store["expiring"].expiry = &pastTime

	count, err = store.Del(ctx, "expiring")
	assert.NoError(t, err)
	assert.Equal(t, int64(0), count) // Expired key not counted as deleted
}
