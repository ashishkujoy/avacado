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

	size, err := lp.push([][]byte{[]byte("avacado"), []byte("listPack")}...)
	assert.NoError(t, err)

	assert.Equal(t, 4, size)
	assert.Equal(t, 4, lp.length())
}

func TestListsMemoryStore_PopElements(t *testing.T) {
	initialElements := [][]byte{[]byte("hello"), []byte("world")}
	lp := newListPack(256, initialElements...)
	lp.push([][]byte{[]byte("avacado"), []byte("listPack")}...)

	elements := lp.pop(1)
	assert.Equal(t, []byte("listPack"), elements[0])
	assert.Equal(t, 3, lp.length())

	elements = lp.pop(2)
	assert.Equal(t, []byte("world"), elements[0])
	assert.Equal(t, []byte("avacado"), elements[1])
	assert.Equal(t, 1, lp.length())

	elements = lp.pop(3)
	assert.Equal(t, 1, len(elements))
	assert.Equal(t, []byte("hello"), elements[0])
	assert.Equal(t, 0, lp.length())
}

func TestListsMemoryStore_LPush(t *testing.T) {
	lp := newEmptyListPack(1024)
	_, _ = lp.push([][]byte{[]byte("world"), []byte("-124")}...)
	_, _ = lp.lPush([]byte("hello"), []byte("1231313"))

	assert.Equal(t, 4, lp.length())
	elements := lp.pop(4)
	assert.Equal(t, []byte("1231313"), elements[0])
	assert.Equal(t, []byte("hello"), elements[1])
	assert.Equal(t, []byte("world"), elements[2])
	assert.Equal(t, []byte("-124"), elements[3])
}

func TestListPack_PushOverflow(t *testing.T) {
	t.Run("single value too large for empty listpack", func(t *testing.T) {
		// maxSize=10: freeBytes = 10-7 = 3; "hello" needs 7 bytes
		lp := newEmptyListPack(10)
		count, err := lp.push([]byte("hello"))
		assert.Error(t, err)
		assert.Equal(t, 0, count)   // original elemCount returned
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
		assert.Equal(t, 1, count)   // original elemCount before failed push
		assert.Equal(t, 1, lp.length()) // first value still intact
	})

	t.Run("listpack not corrupted after overflow error", func(t *testing.T) {
		lp := newEmptyListPack(19)
		_, _ = lp.push([]byte("hello"))
		_, _ = lp.push([]byte("hello")) // overflows, ignored

		// original element still readable
		elems := lp.pop(1)
		assert.Equal(t, []byte("hello"), elems[0])
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
