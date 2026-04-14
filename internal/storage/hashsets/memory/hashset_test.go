package memory

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_SetValues(t *testing.T) {
	t.Run("Set new value", func(t *testing.T) {
		hashSet := NewListPackBasedHashSet()
		assert.Equal(t, 0, hashSet.lp.Length())

		hashSet.Set("Key1", "Value1")
		assert.Equal(t, 2, hashSet.lp.Length())

		hashSet.Set("Key2", "Value2")
		assert.Equal(t, 4, hashSet.lp.Length())
	})

	t.Run("Update existing value", func(t *testing.T) {
		hashSet := NewListPackBasedHashSet()

		hashSet.Set("Key1", "Value1")
		hashSet.Set("Key2", "Value2")
		hashSet.Set("Key3", "Value3")

		value, _ := hashSet.Get("Key2")
		assert.Equal(t, "Value2", string(value))

		hashSet.Set("Key2", "Foo")
		value, _ = hashSet.Get("Key2")
		assert.Equal(t, "Foo", string(value))

		hashSet.Set("Key1", "V1")
		value, _ = hashSet.Get("Key1")
		assert.Equal(t, "V1", string(value))

		hashSet.Set("Key1", "Key1_Value1")
		value, _ = hashSet.Get("Key1")
		assert.Equal(t, "Key1_Value1", string(value))
	})
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
