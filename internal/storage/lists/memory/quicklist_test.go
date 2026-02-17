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
