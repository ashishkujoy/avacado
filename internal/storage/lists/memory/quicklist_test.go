package memory

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestQuickList_RPush(t *testing.T) {
	ql := newQuickList(defaultMaxListPackSize)
	ql.rPush([]byte("hello"))
	ql.rPush([]byte("world"))
	ql.rPush([]byte("avacado"))
	ql.rPush([]byte("listPack"))

	assert.Equal(t, 4, ql.length())
}

func TestQuickList_RPop(t *testing.T) {
	ql := newQuickList(20)
	ql.rPush([]byte("12"))
	ql.rPush([]byte("abcdefghi"))

	assert.Equal(t, 1, len(ql.lps))
	// should create a new listPack here
	ql.rPush([]byte("Hello"))
	assert.Equal(t, 2, len(ql.lps))
	ql.rPush([]byte("1"))
	assert.Equal(t, 2, len(ql.lps))

	data, size := ql.rPop()
	assert.Equal(t, []byte("1"), data)
	assert.Equal(t, 3, size)

	data, size = ql.rPop()
	assert.Equal(t, []byte("Hello"), data)
	assert.Equal(t, 2, size)
	// should release empty listPack
	assert.Equal(t, 1, len(ql.lps))

	data, size = ql.rPop()
	assert.Equal(t, []byte("abcdefghi"), data)
	assert.Equal(t, 1, size)

	data, size = ql.rPop()
	assert.Equal(t, []byte("12"), data)
	assert.Equal(t, 0, size)
	// should not release empty head listPack
	assert.Equal(t, 1, len(ql.lps))
}
