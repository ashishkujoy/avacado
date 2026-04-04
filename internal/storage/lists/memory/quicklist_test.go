package memory

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestQuickList_RPush(t *testing.T) {
	ql := newQuickList(defaultMaxListPackSize)
	ql.rPush([][]byte{[]byte("hello")})
	assert.Equal(t, 1, ql.length())
	ql.rPush([][]byte{
		[]byte("world"),
		[]byte("avacado"),
		[]byte("ListPack"),
	})

	assert.Equal(t, 4, ql.length())
}

func TestQuickList_RPop(t *testing.T) {
	ql := newQuickList(20)
	ql.rPush([][]byte{[]byte("12"), []byte("abcdefghi")})

	assert.Equal(t, 1, len(ql.lps))
	// should create a new ListPack here
	ql.rPush([][]byte{[]byte("Hello")})
	assert.Equal(t, 2, len(ql.lps))
	ql.rPush([][]byte{[]byte("1")})
	assert.Equal(t, 2, len(ql.lps))

	elements, size := ql.rPop(2)
	assert.Equal(t, []byte("1"), elements[0])
	assert.Equal(t, []byte("Hello"), elements[1])
	assert.Equal(t, 2, size)

	elements, size = ql.rPop(4)
	assert.Equal(t, []byte("abcdefghi"), elements[0])
	assert.Equal(t, []byte("12"), elements[1])
	assert.Equal(t, 0, size)
}

func TestQuickList_AtIndex(t *testing.T) {
	ql := newQuickList(20)
	ql.rPush([][]byte{
		[]byte("12"),
		[]byte("abcdefghi"),
		[]byte("1"),
		[]byte("Hello"),
		[]byte("Hello World"),
	})
	// ensure we have three listPacks
	assert.Equal(t, 3, len(ql.lps))

	element, found := ql.atIndex(0)
	assert.True(t, found)
	assert.Equal(t, "12", string(element))

	element, found = ql.atIndex(1)
	assert.True(t, found)
	assert.Equal(t, "abcdefghi", string(element))

	element, found = ql.atIndex(2)
	assert.True(t, found)
	assert.Equal(t, "1", string(element))

	element, found = ql.atIndex(3)
	assert.True(t, found)
	assert.Equal(t, "Hello", string(element))

	element, found = ql.atIndex(4)
	assert.True(t, found)
	assert.Equal(t, "Hello World", string(element))

	element, found = ql.atIndex(-1)
	assert.True(t, found)
	assert.Equal(t, "Hello World", string(element))

	element, found = ql.atIndex(-2)
	assert.True(t, found)
	assert.Equal(t, "Hello", string(element))

	element, found = ql.atIndex(5)
	assert.False(t, found)
	assert.Nil(t, element)

	element, found = ql.atIndex(-6)
	assert.False(t, found)
	assert.Nil(t, element)
}

func TestQuickList_LRange(t *testing.T) {
	ql := newQuickList(20)
	elements := [][]byte{
		[]byte("12"),
		[]byte("abcdefghi"),
		[]byte("1"),
		[]byte("Hello"),
		[]byte("Hello World"),
	}
	ql.rPush(elements)
	// ensure we have three listPacks
	assert.Equal(t, 3, len(ql.lps))

	t.Run("positive start and end", func(t *testing.T) {
		assert.Equal(t, elements, ql.lRange(0, 10))
		assert.Equal(t, elements, ql.lRange(0, 5))
		assert.Equal(t, elements, ql.lRange(0, 4))
		assert.Equal(t, elements[0:4], ql.lRange(0, 3))
	})

	t.Run("Negative start", func(t *testing.T) {
		assert.Equal(t, elements[3:], ql.lRange(-2, 10))
		assert.Equal(t, elements, ql.lRange(-20, 10))
	})

	t.Run("Negative end", func(t *testing.T) {
		assert.Equal(t, elements[2:], ql.lRange(2, -1))
	})
}

func equalSlices(actual, expected [][]byte, t *testing.T) {
	assert.Equal(t, len(expected), len(actual), "Actual and expected have different Length")
	for i, actualElem := range actual {
		assert.Equal(t, expected[i], actualElem, fmt.Sprintf("Elements differ at %d", i))
	}
}
