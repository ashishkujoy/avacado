package memory

import (
	"avacado/internal/storage/lists"
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

func setupListStoreForLMove() *ListMemoryStore {
	store := NewListMemoryStore(20)

	_, _ = store.RPush(
		context.Background(),
		"l1",
		[]byte("a"), []byte("b"), []byte("c"),
	)
	_, _ = store.RPush(
		context.Background(),
		"l2",
		[]byte("1"), []byte("2"), []byte("3"),
	)

	return store
}

func verifyListContainsExactly(
	t *testing.T,
	store *ListMemoryStore,
	key string,
	elements [][]byte,
) {
	actualElements, err := store.LRange(context.Background(), key, 0, -1)
	assert.NoError(t, err)
	assert.Equal(t, len(elements), len(actualElements), "Length not equal")

	for i, actualElement := range actualElements {
		assert.Equal(t, elements[i], actualElement)
	}

}

func asByteSlices(elements ...string) [][]byte {
	bytes := make([][]byte, len(elements))
	for i, element := range elements {
		bytes[i] = []byte(element)
	}
	return bytes
}

func TestListMemoryStore_LMove(t *testing.T) {
	t.Run("move from source left to destination left", func(t *testing.T) {
		store := setupListStoreForLMove()
		element, err := store.LMove(context.Background(), "l1", "l2", lists.Left, lists.Left)

		assert.NoError(t, err)
		assert.Equal(t, []byte("a"), element)

		verifyListContainsExactly(t, store, "l1", asByteSlices("b", "c"))
		verifyListContainsExactly(t, store, "l2", asByteSlices("a", "1", "2", "3"))
	})

	t.Run("move from source left to destination right", func(t *testing.T) {
		store := setupListStoreForLMove()
		element, err := store.LMove(context.Background(), "l1", "l2", lists.Left, lists.Right)

		assert.NoError(t, err)
		assert.Equal(t, []byte("a"), element)

		verifyListContainsExactly(t, store, "l1", asByteSlices("b", "c"))
		verifyListContainsExactly(t, store, "l2", asByteSlices("1", "2", "3", "a"))
	})

	t.Run("move from source right to destination right", func(t *testing.T) {
		store := setupListStoreForLMove()
		element, err := store.LMove(context.Background(), "l1", "l2", lists.Right, lists.Right)

		assert.NoError(t, err)
		assert.Equal(t, []byte("c"), element)

		verifyListContainsExactly(t, store, "l1", asByteSlices("a", "b"))
		verifyListContainsExactly(t, store, "l2", asByteSlices("1", "2", "3", "c"))
	})

	t.Run("move from source right to destination left", func(t *testing.T) {
		store := setupListStoreForLMove()
		element, err := store.LMove(context.Background(), "l1", "l2", lists.Right, lists.Left)

		assert.NoError(t, err)
		assert.Equal(t, []byte("c"), element)

		verifyListContainsExactly(t, store, "l1", asByteSlices("a", "b"))
		verifyListContainsExactly(t, store, "l2", asByteSlices("c", "1", "2", "3"))
	})

	t.Run("move from source to non existing destination", func(t *testing.T) {
		store := setupListStoreForLMove()
		element, err := store.LMove(context.Background(), "l1", "l3", lists.Right, lists.Left)

		assert.NoError(t, err)
		assert.Equal(t, []byte("c"), element)

		verifyListContainsExactly(t, store, "l1", asByteSlices("a", "b"))
		verifyListContainsExactly(t, store, "l3", asByteSlices("c"))
	})

	t.Run("move from non existing source", func(t *testing.T) {
		store := setupListStoreForLMove()
		element, err := store.LMove(context.Background(), "l3", "l1", lists.Right, lists.Left)

		assert.NoError(t, err)
		assert.Nil(t, element)
	})

	t.Run("move from empty source", func(t *testing.T) {
		store := setupListStoreForLMove()
		_, _ = store.LPush(context.Background(), "l3", asByteSlices("1")...)
		_, _ = store.LPop(context.Background(), "l3", 1)
		element, err := store.LMove(context.Background(), "l3", "l1", lists.Right, lists.Left)

		assert.NoError(t, err)
		assert.Nil(t, element)
	})
}
