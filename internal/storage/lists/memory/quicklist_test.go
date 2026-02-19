package memory

import (
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
		[]byte("listPack"),
	})

	assert.Equal(t, 4, ql.length())
}

func TestQuickList_RPop(t *testing.T) {
	ql := newQuickList(20)
	ql.rPush([][]byte{[]byte("12"), []byte("abcdefghi")})

	assert.Equal(t, 1, len(ql.lps))
	// should create a new listPack here
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
