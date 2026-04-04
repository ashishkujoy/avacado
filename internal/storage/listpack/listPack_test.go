package listpack

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestListPack_NewEmptyListPack(t *testing.T) {
	lp := NewEmptyListPack(256)
	assert.Equal(t, 0, lp.Length())
}

func TestListPack_NewListPack(t *testing.T) {
	elements := [][]byte{[]byte("hello"), []byte("world")}
	lp := NewListPack(256, elements...)
	assert.Equal(t, 2, lp.Length())
}

func TestListPack_PushElements(t *testing.T) {
	initialElements := [][]byte{[]byte("hello"), []byte("world")}
	lp := NewListPack(256, initialElements...)

	_, _ = lp.Push([]byte("avacado"))
	size, err := lp.Push([]byte("ListPack"))
	assert.NoError(t, err)

	assert.Equal(t, 4, size)
	assert.Equal(t, 4, lp.Length())
}

func TestListsMemoryStore_PopElements(t *testing.T) {
	initialElements := [][]byte{[]byte("hello"), []byte("world")}
	lp := NewListPack(256, initialElements...)
	_, _ = lp.Push([]byte("avacado"))
	_, _ = lp.Push([]byte("ListPack"))

	assert.Equal(t, []byte("ListPack"), lp.Pop())
	assert.Equal(t, []byte("avacado"), lp.Pop())
	assert.Equal(t, []byte("world"), lp.Pop())
	assert.Equal(t, []byte("hello"), lp.Pop())
}

func TestListsMemoryStore_LPush(t *testing.T) {
	lp := NewEmptyListPack(1024)
	_, _ = lp.Push([]byte("world"))
	_, _ = lp.Push([]byte("-124"))
	_, _ = lp.LPush([]byte("hello"))
	_, _ = lp.LPush([]byte("1231313"))

	assert.Equal(t, 4, lp.Length())
	assert.Equal(t, []byte("-124"), lp.Pop())
	assert.Equal(t, []byte("world"), lp.Pop())
	assert.Equal(t, []byte("hello"), lp.Pop())
	assert.Equal(t, []byte("1231313"), lp.Pop())
}

func TestListPack_PushOverflow(t *testing.T) {
	t.Run("single value too large for empty listpack", func(t *testing.T) {
		// maxSize=10: freeBytes = 10-7 = 3; "hello" needs 7 bytes
		lp := NewEmptyListPack(10)
		count, err := lp.Push([]byte("hello"))
		assert.Error(t, err)
		assert.Equal(t, 0, count)       // original elemCount returned
		assert.Equal(t, 0, lp.Length()) // listpack not modified
	})

	t.Run("second value overflows, first value committed", func(t *testing.T) {
		// maxSize=19: freeBytes=12; "hello"(7) fits → totalSize=14, freeBytes=5; second "hello"(7) doesn't
		lp := NewEmptyListPack(19)
		count, err := lp.Push([]byte("hello"))
		assert.NoError(t, err)
		assert.Equal(t, 1, count)

		count, err = lp.Push([]byte("hello"))
		assert.Error(t, err)
		assert.Equal(t, 1, count)       // original elemCount before failed Push
		assert.Equal(t, 1, lp.Length()) // first value still intact
	})

	t.Run("listpack not corrupted after overflow error", func(t *testing.T) {
		lp := NewEmptyListPack(19)
		_, _ = lp.Push([]byte("hello"))
		_, _ = lp.Push([]byte("hello")) // overflows, ignored

		assert.Equal(t, []byte("hello"), lp.Pop())
	})
}

func TestListPack_LPushOverflow(t *testing.T) {
	t.Run("single value too large for empty listpack", func(t *testing.T) {
		// maxSize=10: freeBytes=3; "hello" needs 7 bytes
		lp := NewEmptyListPack(10)
		count, err := lp.LPush([]byte("hello"))
		assert.Error(t, err)
		assert.Equal(t, 0, count)
		assert.Equal(t, 0, lp.Length())
	})

	t.Run("second value overflows, first value committed", func(t *testing.T) {
		lp := NewEmptyListPack(19)
		count, err := lp.LPush([]byte("hello"))
		assert.NoError(t, err)
		assert.Equal(t, 1, count)

		count, err = lp.LPush([]byte("hello"))
		assert.Error(t, err)
		assert.Equal(t, 1, count)
		assert.Equal(t, 1, lp.Length())
	})
}

func TestListsMemoryStore_LPop(t *testing.T) {
	elements := [][]byte{[]byte("hello"), []byte("world"), []byte("124"), []byte("JamesBond")}
	lp := NewListPack(1024, elements...)

	popped := lp.LPop(2)
	assert.Equal(t, 2, len(popped))
	assert.Equal(t, elements[0], popped[0])
	assert.Equal(t, elements[1], popped[1])

	popped = lp.LPop(3)
	assert.Equal(t, 2, len(popped))
	assert.Equal(t, string(elements[2]), string(popped[0]))
	assert.Equal(t, elements[3], popped[1])

	popped = lp.LPop(4)
	assert.Equal(t, 0, len(popped))
}

func TestListsMemoryStore_IsEmpty(t *testing.T) {
	lp := NewEmptyListPack(24)
	assert.True(t, lp.IsEmpty())

	_, _ = lp.Push([]byte("12"))
	assert.False(t, lp.IsEmpty())

	lp.Pop()
	assert.True(t, lp.IsEmpty())
}

func TestListMemoryStore_LIndex(t *testing.T) {
	lp := NewEmptyListPack(24)
	_, _ = lp.Push([]byte("Hi"))
	_, _ = lp.Push([]byte("120"))
	_, _ = lp.Push([]byte("300"))
	_, _ = lp.Push([]byte("hi bye"))

	t.Run("Positive index only", func(t *testing.T) {
		element, found := lp.AtIndex(3)
		assert.True(t, found)
		assert.Equal(t, []byte("hi bye"), element)

		element, found = lp.AtIndex(1)
		assert.True(t, found)
		assert.Equal(t, []byte("120"), element)

		element, found = lp.AtIndex(0)
		assert.True(t, found)
		assert.Equal(t, []byte("Hi"), element)

		element, found = lp.AtIndex(4)
		assert.False(t, found)
		assert.Nil(t, element)
	})

	t.Run("Negative index", func(t *testing.T) {
		element, found := lp.AtIndex(-1)
		assert.True(t, found)
		assert.Equal(t, []byte("hi bye"), element)

		element, found = lp.AtIndex(-2)
		assert.True(t, found)
		assert.Equal(t, []byte("300"), element)

		element, found = lp.AtIndex(-3)
		assert.True(t, found)
		assert.Equal(t, []byte("120"), element)

		element, found = lp.AtIndex(-4)
		assert.True(t, found)
		assert.Equal(t, []byte("Hi"), element)

		element, found = lp.AtIndex(-5)
		assert.False(t, found)
		assert.Nil(t, element)
	})
}

func TestListPack_LRange(t *testing.T) {
	t.Run("from zero index to last", func(t *testing.T) {
		lp := NewListPack(60, []byte("Hi"), []byte("hello World"), []byte("12"))
		elements, err := lp.LRange(0, 3)
		assert.NoError(t, err)
		assert.Equal(t, 3, len(elements))
		assert.Equal(t, []byte("Hi"), elements[0])
		assert.Equal(t, []byte("hello World"), elements[1])
		assert.Equal(t, []byte("12"), elements[2])
	})

	t.Run("from non zero index to last", func(t *testing.T) {
		lp := NewListPack(60, []byte("Hi"), []byte("hello World"), []byte("12"), []byte("43"))
		elements, err := lp.LRange(1, 3)
		assert.NoError(t, err)
		assert.Equal(t, 3, len(elements))
		assert.Equal(t, []byte("hello World"), elements[0])
		assert.Equal(t, []byte("12"), elements[1])
		assert.Equal(t, []byte("43"), elements[2])
	})

	t.Run("from non zero index to non last", func(t *testing.T) {
		lp := NewListPack(60, []byte("Hi"), []byte("hello World"), []byte("12"), []byte("43"))
		elements, err := lp.LRange(1, 2)
		assert.NoError(t, err)
		assert.Equal(t, 2, len(elements))
		assert.Equal(t, []byte("hello World"), elements[0])
		assert.Equal(t, []byte("12"), elements[1])
	})
}
