package memory

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKVMemoryStore_GetAndSet(t *testing.T) {
	store := NewKVMemoryStore()
	value, err := store.Get(context.Background(), "key1")

	assert.NoError(t, err)
	assert.Nil(t, value)

	err = store.Set(context.Background(), "key1", []byte("value1"))

	assert.NoError(t, err)

	value, err = store.Get(context.Background(), "key1")

	assert.NoError(t, err)
	assert.Equal(t, "value1", string(value))
}
