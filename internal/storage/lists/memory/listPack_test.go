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
	lp.push([][]byte{[]byte("world"), []byte("-124")}...)
	lp.lPush([]byte("hello"), []byte("1231313"))

	assert.Equal(t, 4, lp.length())
	elements := lp.pop(4)
	assert.Equal(t, []byte("1231313"), elements[0])
	assert.Equal(t, []byte("hello"), elements[1])
	assert.Equal(t, []byte("world"), elements[2])
	assert.Equal(t, []byte("-124"), elements[3])
}
