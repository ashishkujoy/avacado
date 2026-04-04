package memory

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_SetValues(t *testing.T) {
	hashSet := NewListPackBasedHashSet()
	assert.Equal(t, 0, hashSet.lp.Length())

	hashSet.Set("Key1", "Value1")
	assert.Equal(t, 2, hashSet.lp.Length())

	hashSet.Set("Key2", "Value2")
	assert.Equal(t, 4, hashSet.lp.Length())
}

func Test_GetValue(t *testing.T) {
	hashSet := NewListPackBasedHashSet()
	hashSet.Set("Key1", "Value1")
	hashSet.Set("Key2", "Value2")

	v2, ok := hashSet.Get("Key2")
	assert.True(t, ok)
	assert.Equal(t, []byte("Value2"), v2)

	v1, ok := hashSet.Get("Key1")
	assert.True(t, ok)
	assert.Equal(t, []byte("Value1"), v1)

	_, ok = hashSet.Get("Key3")
	assert.False(t, ok)
}
