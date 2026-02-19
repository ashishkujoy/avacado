package memory

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestListMemoryStore_RPush(t *testing.T) {
	store := NewListMemoryStore(1024)

	size, err := store.RPush(context.Background(), "Foo", []byte("Hello"), []byte("World"))
	assert.NoError(t, err)
	assert.Equal(t, 2, size)

	size, err = store.RPush(context.Background(), "Foo", []byte("1"))
	assert.NoError(t, err)
	assert.Equal(t, 3, size)

	size, err = store.RPush(context.Background(), "bar", []byte("2"))
	assert.NoError(t, err)
	assert.Equal(t, 1, size)
}

func TestListMemoryStore_RPop(t *testing.T) {
	t.Run("Pop from existing list", func(t *testing.T) {
		store := NewListMemoryStore(1024)

		_, _ = store.RPush(context.Background(), "Foo", []byte("Hello"), []byte("World"))
		elements, err := store.RPop(context.Background(), "Foo", 3)
		assert.NoError(t, err)
		assert.Equal(t, 2, len(elements))
		assert.Equal(t, []byte("World"), elements[0])
		assert.Equal(t, []byte("Hello"), elements[1])
	})

	t.Run("Pop from a non existing list", func(t *testing.T) {
		store := NewListMemoryStore(1024)

		elements, err := store.RPop(context.Background(), "non-existing-key", 12)
		assert.NoError(t, err)
		assert.Nil(t, elements)
	})
}

func TestListMemoryStore_Len(t *testing.T) {
	t.Run("Len of existing list", func(t *testing.T) {
		store := NewListMemoryStore(1024)

		_, _ = store.RPush(context.Background(), "Foo", []byte("Hello"), []byte("World"))
		l, err := store.Len(context.Background(), "Foo")
		assert.NoError(t, err)
		assert.Equal(t, 2, l)
	})

	t.Run("Len of non existing list", func(t *testing.T) {
		NewListMemoryStore(1024)
	})
}
