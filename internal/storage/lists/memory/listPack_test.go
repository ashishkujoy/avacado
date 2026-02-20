package memory

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestListPack_NewEmptyListPack(t *testing.T) {
	lp := newEmptyListPack(256)
	assert.Equal(t, 0, lp.length())
}

func TestListPack_NewListPack(t *testing.T) {
	elements := [][]byte{[]byte("hello"), []byte("world")}
	lp := newListPack(256, elements...)
	assert.Equal(t, 2, lp.length())
}

func TestListPack_PushElements(t *testing.T) {
	initialElements := [][]byte{[]byte("hello"), []byte("world")}
	lp := newListPack(256, initialElements...)

	_, _ = lp.push([]byte("avacado"))
	size, err := lp.push([]byte("listPack"))
	assert.NoError(t, err)

	assert.Equal(t, 4, size)
	assert.Equal(t, 4, lp.length())
}

func TestListsMemoryStore_PopElements(t *testing.T) {
	initialElements := [][]byte{[]byte("hello"), []byte("world")}
	lp := newListPack(256, initialElements...)
	_, _ = lp.push([]byte("avacado"))
	_, _ = lp.push([]byte("listPack"))

	assert.Equal(t, []byte("listPack"), lp.pop())
	assert.Equal(t, []byte("avacado"), lp.pop())
	assert.Equal(t, []byte("world"), lp.pop())
	assert.Equal(t, []byte("hello"), lp.pop())
}

func TestListsMemoryStore_LPush(t *testing.T) {
	lp := newEmptyListPack(1024)
	_, _ = lp.push([]byte("world"))
	_, _ = lp.push([]byte("-124"))
	_, _ = lp.lPush([]byte("hello"))
	_, _ = lp.lPush([]byte("1231313"))

	assert.Equal(t, 4, lp.length())
	assert.Equal(t, []byte("-124"), lp.pop())
	assert.Equal(t, []byte("world"), lp.pop())
	assert.Equal(t, []byte("hello"), lp.pop())
	assert.Equal(t, []byte("1231313"), lp.pop())
}

func TestListPack_PushOverflow(t *testing.T) {
	t.Run("single value too large for empty listpack", func(t *testing.T) {
		// maxSize=10: freeBytes = 10-7 = 3; "hello" needs 7 bytes
		lp := newEmptyListPack(10)
		count, err := lp.push([]byte("hello"))
		assert.Error(t, err)
		assert.Equal(t, 0, count)       // original elemCount returned
		assert.Equal(t, 0, lp.length()) // listpack not modified
	})

	t.Run("second value overflows, first value committed", func(t *testing.T) {
		// maxSize=19: freeBytes=12; "hello"(7) fits → totalSize=14, freeBytes=5; second "hello"(7) doesn't
		lp := newEmptyListPack(19)
		count, err := lp.push([]byte("hello"))
		assert.NoError(t, err)
		assert.Equal(t, 1, count)

		count, err = lp.push([]byte("hello"))
		assert.Error(t, err)
		assert.Equal(t, 1, count)       // original elemCount before failed push
		assert.Equal(t, 1, lp.length()) // first value still intact
	})

	t.Run("listpack not corrupted after overflow error", func(t *testing.T) {
		lp := newEmptyListPack(19)
		_, _ = lp.push([]byte("hello"))
		_, _ = lp.push([]byte("hello")) // overflows, ignored

		assert.Equal(t, []byte("hello"), lp.pop())
	})
}

func TestListPack_LPushOverflow(t *testing.T) {
	t.Run("single value too large for empty listpack", func(t *testing.T) {
		// maxSize=10: freeBytes=3; "hello" needs 7 bytes
		lp := newEmptyListPack(10)
		count, err := lp.lPush([]byte("hello"))
		assert.Error(t, err)
		assert.Equal(t, 0, count)
		assert.Equal(t, 0, lp.length())
	})

	t.Run("second value overflows, first value committed", func(t *testing.T) {
		lp := newEmptyListPack(19)
		count, err := lp.lPush([]byte("hello"))
		assert.NoError(t, err)
		assert.Equal(t, 1, count)

		count, err = lp.lPush([]byte("hello"))
		assert.Error(t, err)
		assert.Equal(t, 1, count)
		assert.Equal(t, 1, lp.length())
	})
}

func TestListsMemoryStore_LPop(t *testing.T) {
	elements := [][]byte{[]byte("hello"), []byte("world"), []byte("124"), []byte("JamesBond")}
	lp := newListPack(1024, elements...)

	popped := lp.lPop(2)
	assert.Equal(t, 2, len(popped))
	assert.Equal(t, elements[0], popped[0])
	assert.Equal(t, elements[1], popped[1])

	popped = lp.lPop(3)
	assert.Equal(t, 2, len(popped))
	assert.Equal(t, string(elements[2]), string(popped[0]))
	assert.Equal(t, elements[3], popped[1])

	popped = lp.lPop(4)
	assert.Equal(t, 0, len(popped))
}

func TestListsMemoryStore_IsEmpty(t *testing.T) {
	lp := newEmptyListPack(24)
	assert.True(t, lp.isEmpty())

	_, _ = lp.push([]byte("12"))
	assert.False(t, lp.isEmpty())

	lp.pop()
	assert.True(t, lp.isEmpty())
}

func TestListMemoryStore_LIndex(t *testing.T) {
	lp := newEmptyListPack(24)
	_, _ = lp.push([]byte("Hi"))
	_, _ = lp.push([]byte("120"))
	_, _ = lp.push([]byte("300"))
	_, _ = lp.push([]byte("hi bye"))

	t.Run("Positive index only", func(t *testing.T) {
		element, found := lp.atIndex(3)
		assert.True(t, found)
		assert.Equal(t, []byte("hi bye"), element)

		element, found = lp.atIndex(1)
		assert.True(t, found)
		assert.Equal(t, []byte("120"), element)

		element, found = lp.atIndex(0)
		assert.True(t, found)
		assert.Equal(t, []byte("Hi"), element)

		element, found = lp.atIndex(4)
		assert.False(t, found)
		assert.Nil(t, element)
	})

	t.Run("Negative index", func(t *testing.T) {
		element, found := lp.atIndex(-1)
		assert.True(t, found)
		assert.Equal(t, []byte("hi bye"), element)

		element, found = lp.atIndex(-2)
		assert.True(t, found)
		assert.Equal(t, []byte("300"), element)

		element, found = lp.atIndex(-3)
		assert.True(t, found)
		assert.Equal(t, []byte("120"), element)

		element, found = lp.atIndex(-4)
		assert.True(t, found)
		assert.Equal(t, []byte("Hi"), element)

		element, found = lp.atIndex(-5)
		assert.False(t, found)
		assert.Nil(t, element)
	})
}
