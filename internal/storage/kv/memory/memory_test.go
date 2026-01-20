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

	err = store.Set(context.Background(), "key1", []byte("value1"), kv.NewSetOptions(false))

	assert.NoError(t, err)

	value, err = store.Get(context.Background(), "key1")

	assert.NoError(t, err)
	assert.Equal(t, "value1", string(value))
}

func TestKVMemoryStore_SetExistingKeyWithNXOptionEnabled(t *testing.T) {
	store := NewKVMemoryStore()

	err := store.Set(context.Background(), "key1", []byte("value1"), kv.NewSetOptions(true))
	assert.NoError(t, err)

	err = store.Set(context.Background(), "key1", []byte("value2"), kv.NewSetOptions(true))
	assert.Error(t, err, NewKeyAlreadyExistsError("key1"))
}

func TestKVMemoryStore_SetExistingKeyWithNXOptionDisabled(t *testing.T) {
	store := NewKVMemoryStore()

	err := store.Set(context.Background(), "key1", []byte("value1"), kv.NewSetOptions(false))
	assert.NoError(t, err)

	err = store.Set(context.Background(), "key1", []byte("value2"), kv.NewSetOptions(false))
	assert.NoError(t, err)

	value, _ := store.Get(context.Background(), "key1")
	assert.Equal(t, "value2", string(value))
}
