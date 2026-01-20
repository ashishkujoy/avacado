package memory

import (
	"avacado/internal/storage/kv"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKVMemoryStore_GetAndSet(t *testing.T) {
	store := NewKVMemoryStore()
	value, err := store.Get(context.Background(), "key1")

	assert.NoError(t, err)
	assert.Nil(t, value)
	options := kv.NewSetOptions()

	err = store.Set(context.Background(), "key1", []byte("value1"), options)

	assert.NoError(t, err)

	value, err = store.Get(context.Background(), "key1")

	assert.NoError(t, err)
	assert.Equal(t, "value1", string(value))
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
