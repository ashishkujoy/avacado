package memory

import (
	"avacado/internal/storage/kv"
	"context"
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

// TestKVMemoryStore_LazyExpirationImmediateCleanup verifies that lazy expiration
// immediately removes expired keys on GET
func TestKVMemoryStore_LazyExpirationImmediateCleanup(t *testing.T) {
	store := NewKVMemoryStore()
	options := kv.NewSetOptions().WithEX(1)

	_, err := store.Set(context.Background(), "key1", []byte("value1"), options)
	assert.NoError(t, err)

	time.Sleep(1100 * time.Millisecond)

	v, err := store.Get(context.Background(), "key1")
	assert.NoError(t, err)
	assert.Nil(t, v)

	v, err = store.Get(context.Background(), "key1")
	assert.NoError(t, err)
	assert.Nil(t, v)
}

func TestKVMemoryStore_Incr(t *testing.T) {
	store := NewKVMemoryStore()
	ctx := context.Background()

	val, err := store.Incr(ctx, "counter")
	assert.NoError(t, err)
	assert.Equal(t, int64(1), val)

	_, err = store.Set(ctx, "counter1", []byte("10"), kv.NewSetOptions())
	val, err = store.Incr(ctx, "counter1")
	assert.NoError(t, err)
	assert.Equal(t, int64(11), val)

	_, err = store.Set(ctx, "counter2", []byte("20"), kv.NewSetOptions().WithEX(1))
	pastTime := time.Now().Add(-2 * time.Second)
	store.store["counter2"].expiry = &pastTime

	v, err := store.Incr(ctx, "counter2")
	assert.NoError(t, err)
	assert.Equal(t, int64(1), v)

	_, err = store.Set(context.Background(), "counter3", []byte("not-an-integer"), kv.NewSetOptions())
	assert.NoError(t, err)

	_, err = store.Incr(ctx, "counter3")
	assert.Error(t, err)
}

func TestKVMemoryStore_Decr(t *testing.T) {
	store := NewKVMemoryStore()
	ctx := context.Background()

	val, err := store.Decr(ctx, "counter")
	assert.NoError(t, err)
	assert.Equal(t, int64(-1), val)

	_, err = store.Set(ctx, "counter1", []byte("10"), kv.NewSetOptions())
	val, err = store.Decr(ctx, "counter1")
	assert.NoError(t, err)
	assert.Equal(t, int64(9), val)

	_, err = store.Set(ctx, "counter2", []byte("20"), kv.NewSetOptions().WithEX(1))
	pastTime := time.Now().Add(-2 * time.Second)
	store.store["counter2"].expiry = &pastTime

	v, err := store.Decr(ctx, "counter2")
	assert.NoError(t, err)
	assert.Equal(t, int64(-1), v)

	_, err = store.Set(context.Background(), "counter3", []byte("not-an-integer"), kv.NewSetOptions())
	assert.NoError(t, err)

	_, err = store.Decr(ctx, "counter3")
	assert.Error(t, err)
}

func TestKVMemoryStore_DecrBy(t *testing.T) {
	store := NewKVMemoryStore()
	ctx := context.Background()

	val, err := store.DecrBy(ctx, "counter", 5)
	assert.NoError(t, err)
	assert.Equal(t, int64(-5), val)

	val, err = store.DecrBy(ctx, "counter", 3)
	assert.NoError(t, err)
	assert.Equal(t, int64(-8), val)

	_, err = store.Set(ctx, "counter1", []byte("100"), kv.NewSetOptions())
	val, err = store.DecrBy(ctx, "counter1", 20)
	assert.NoError(t, err)
	assert.Equal(t, int64(80), val)

	val, err = store.DecrBy(ctx, "counter1", 30)
	assert.NoError(t, err)
	assert.Equal(t, int64(50), val)

	_, err = store.Set(ctx, "counter2", []byte("50"), kv.NewSetOptions().WithEX(1))
	pastTime := time.Now().Add(-2 * time.Second)
	store.store["counter2"].expiry = &pastTime

	v, err := store.DecrBy(ctx, "counter2", 15)
	assert.NoError(t, err)
	assert.Equal(t, int64(-15), v)

	_, err = store.Set(context.Background(), "counter3", []byte("not-an-integer"), kv.NewSetOptions())
	assert.NoError(t, err)

	_, err = store.DecrBy(ctx, "counter3", 10)
	assert.Error(t, err)
}

func TestKVMemoryStore_Del(t *testing.T) {
	store := NewKVMemoryStore()
	ctx := context.Background()

	count, err := store.Del(ctx, "nonexistent")
	assert.NoError(t, err)
	assert.Equal(t, int64(0), count)

	_, err = store.Set(ctx, "key1", []byte("value1"), kv.NewSetOptions())
	assert.NoError(t, err)
	_, err = store.Set(ctx, "key2", []byte("value2"), kv.NewSetOptions())
	assert.NoError(t, err)
	_, err = store.Set(ctx, "key3", []byte("value3"), kv.NewSetOptions())
	assert.NoError(t, err)

	count, err = store.Del(ctx, "key1")
	assert.NoError(t, err)
	assert.Equal(t, int64(1), count)

	val, err := store.Get(ctx, "key1")
	assert.NoError(t, err)
	assert.Nil(t, val)

	count, err = store.Del(ctx, "key2", "key3")
	assert.NoError(t, err)
	assert.Equal(t, int64(2), count)

	val, err = store.Get(ctx, "key2")
	assert.NoError(t, err)
	assert.Nil(t, val)
	val, err = store.Get(ctx, "key3")
	assert.NoError(t, err)
	assert.Nil(t, val)

	_, err = store.Set(ctx, "key4", []byte("value4"), kv.NewSetOptions())
	assert.NoError(t, err)
	count, err = store.Del(ctx, "key4", "nonexistent", "alsonothere")
	assert.NoError(t, err)
	assert.Equal(t, int64(1), count)

	_, err = store.Set(ctx, "expiring", []byte("value"), kv.NewSetOptions().WithEX(1))
	assert.NoError(t, err)
	pastTime := time.Now().Add(-2 * time.Second)
	store.store["expiring"].expiry = &pastTime

	count, err = store.Del(ctx, "expiring")
	assert.NoError(t, err)
	assert.Equal(t, int64(0), count)
}

func TestKVMemoryStore_Exists(t *testing.T) {
	store := NewKVMemoryStore()
	ctx := context.Background()

	count, err := store.Exists(ctx, "nonexistent")
	assert.NoError(t, err)
	assert.Equal(t, int64(0), count)

	_, err = store.Set(ctx, "key1", []byte("value1"), kv.NewSetOptions())
	assert.NoError(t, err)
	count, err = store.Exists(ctx, "key1")
	assert.NoError(t, err)
	assert.Equal(t, int64(1), count)

	_, err = store.Set(ctx, "key2", []byte("value2"), kv.NewSetOptions())
	assert.NoError(t, err)
	_, err = store.Set(ctx, "key3", []byte("value3"), kv.NewSetOptions())
	assert.NoError(t, err)

	count, err = store.Exists(ctx, "key1", "key2", "key3")
	assert.NoError(t, err)
	assert.Equal(t, int64(3), count)

	count, err = store.Exists(ctx, "key1", "nonexistent", "key2")
	assert.NoError(t, err)
	assert.Equal(t, int64(2), count)

	_, err = store.Set(ctx, "expiring", []byte("value"), kv.NewSetOptions().WithEX(1))
	assert.NoError(t, err)
	pastTime := time.Now().Add(-2 * time.Second)
	store.store["expiring"].expiry = &pastTime

	count, err = store.Exists(ctx, "expiring")
	assert.NoError(t, err)
	assert.Equal(t, int64(0), count)

	count, err = store.Exists(ctx, "key1", "key1", "key1")
	assert.NoError(t, err)
	assert.Equal(t, int64(3), count)

	count, err = store.Exists(ctx, "key1", "nonexistent", "key1")
	assert.NoError(t, err)
	assert.Equal(t, int64(2), count)
}

func TestKVMemoryStore_NumberEncoding(t *testing.T) {
	v := newValue([]byte("123"), nil)
	n, err := v.AsInt64()
	assert.NoError(t, err)
	assert.Equal(t, n, int64(123))

	v.data = encodeNumber(80)
	n, err = v.AsInt64()
	assert.NoError(t, err)
	assert.Equal(t, n, int64(80))
}

func TestKVMemoryStore_SetWithIFEQMatchingValue(t *testing.T) {
	store := NewKVMemoryStore()
	ctx := context.Background()

	_, err := store.Set(ctx, "key1", []byte("oldvalue"), kv.NewSetOptions())
	assert.NoError(t, err)

	options := kv.NewSetOptions().WithIFEQ([]byte("oldvalue"))
	_, err = store.Set(ctx, "key1", []byte("newvalue"), options)
	assert.NoError(t, err)

	val, err := store.Get(ctx, "key1")
	assert.NoError(t, err)
	assert.Equal(t, []byte("newvalue"), val)
}

func TestKVMemoryStore_SetWithIFEQNonMatchingValue(t *testing.T) {
	store := NewKVMemoryStore()
	ctx := context.Background()

	_, err := store.Set(ctx, "key1", []byte("oldvalue"), kv.NewSetOptions())
	assert.NoError(t, err)

	options := kv.NewSetOptions().WithIFEQ([]byte("differentvalue"))
	_, err = store.Set(ctx, "key1", []byte("newvalue"), options)
	assert.Error(t, err)

	val, err := store.Get(ctx, "key1")
	assert.NoError(t, err)
	assert.Equal(t, []byte("oldvalue"), val)
}

func TestKVMemoryStore_SetWithIFEQNonExistentKey(t *testing.T) {
	store := NewKVMemoryStore()
	ctx := context.Background()

	options := kv.NewSetOptions().WithIFEQ([]byte("somevalue"))
	_, err := store.Set(ctx, "nonexistent", []byte("newvalue"), options)
	assert.Error(t, err)

	val, err := store.Get(ctx, "nonexistent")
	assert.NoError(t, err)
	assert.Nil(t, val)
}

func TestKVMemoryStore_SetWithIFEQExpiredKey(t *testing.T) {
	store := NewKVMemoryStore()
	ctx := context.Background()

	_, err := store.Set(ctx, "key1", []byte("oldvalue"), kv.NewSetOptions().WithEX(1))
	assert.NoError(t, err)

	pastTime := time.Now().Add(-2 * time.Second)
	store.store["key1"].expiry = &pastTime

	options := kv.NewSetOptions().WithIFEQ([]byte("oldvalue"))
	_, err = store.Set(ctx, "key1", []byte("newvalue"), options)
	assert.Error(t, err)

	val, err := store.Get(ctx, "key1")
	assert.NoError(t, err)
	assert.Nil(t, val)
}

func TestKVMemoryStore_SetWithIFEQAndGet(t *testing.T) {
	store := NewKVMemoryStore()
	ctx := context.Background()

	_, err := store.Set(ctx, "key1", []byte("oldvalue"), kv.NewSetOptions())
	assert.NoError(t, err)

	options := kv.NewSetOptions().WithIFEQ([]byte("oldvalue")).WithGet()
	oldVal, err := store.Set(ctx, "key1", []byte("newvalue"), options)
	assert.NoError(t, err)
	assert.Equal(t, []byte("oldvalue"), oldVal)

	val, err := store.Get(ctx, "key1")
	assert.NoError(t, err)
	assert.Equal(t, []byte("newvalue"), val)
}
