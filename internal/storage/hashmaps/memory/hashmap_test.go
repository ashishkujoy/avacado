package memory

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_SetValues(t *testing.T) {
	t.Run("Set new value", func(t *testing.T) {
		hashSet := NewHashMap()
		assert.Equal(t, 0, hashSet.lp.Length())

		hashSet.Set("Key1", "Value1")
		assert.Equal(t, 2, hashSet.lp.Length())

		hashSet.Set("Key2", "Value2")
		assert.Equal(t, 4, hashSet.lp.Length())
	})

	t.Run("Update existing value", func(t *testing.T) {
		hashSet := NewHashMap()

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

	t.Run("migrate to hashmap when key count goes beyond threshold", func(t *testing.T) {
		hs := NewHashMap()
		for i := 0; i < maxEntryCount; i++ {
			hs.Set(fmt.Sprintf("%d", i), "hi")
		}
		assert.Nil(t, hs.hash)
		assert.NotNil(t, hs.lp)

		hs.Set(fmt.Sprintf("%d", maxEntryCount+1), "Value")
		assert.NotNil(t, hs.hash)
		assert.Nil(t, hs.lp)
	})

	t.Run("migrate to hashmap when key size exceed threshold", func(t *testing.T) {
		hs := NewHashMap()
		hs.Set("AA", "Value1")
		assert.Nil(t, hs.hash)
		assert.NotNil(t, hs.lp)

		bigKey := strings.Repeat("A", maxEntrySize)
		hs.Set(bigKey, "Value2")
		assert.NotNil(t, hs.hash)
		assert.Nil(t, hs.lp)
	})

	t.Run("migrate to hashmap when value size exceed threshold", func(t *testing.T) {
		hs := NewHashMap()
		hs.Set("AA", "Value1")
		assert.Nil(t, hs.hash)
		assert.NotNil(t, hs.lp)

		bigValue := strings.Repeat("A", maxEntrySize)
		hs.Set("BB", bigValue)
		assert.NotNil(t, hs.hash)
		assert.Nil(t, hs.lp)
	})

	t.Run("migrate to hashmap when value size exceed threshold for existing key", func(t *testing.T) {
		hs := NewHashMap()
		hs.Set("AA", "Value1")
		assert.Nil(t, hs.hash)
		assert.NotNil(t, hs.lp)

		bigValue := strings.Repeat("A", maxEntrySize)
		hs.Set("AA", bigValue)
		assert.NotNil(t, hs.hash)
		assert.Nil(t, hs.lp)
	})

}

func Test_GetAll(t *testing.T) {
	t.Run("returns empty map when no entries - listpack encoding", func(t *testing.T) {
		hs := NewHashMap()
		result := hs.GetAll()
		assert.NotNil(t, result)
		assert.Empty(t, result)
	})

	t.Run("returns all key-value pairs - listpack encoding", func(t *testing.T) {
		hs := NewHashMap()
		hs.Set("Key1", "Value1")
		hs.Set("Key2", "Value2")
		hs.Set("Key3", "Value3")

		assert.Nil(t, hs.hash)
		assert.NotNil(t, hs.lp)

		result := hs.GetAll()
		assert.Equal(t, map[string]string{
			"Key1": "Value1",
			"Key2": "Value2",
			"Key3": "Value3",
		}, result)
	})

	t.Run("returns a copy, not the underlying storage - listpack encoding", func(t *testing.T) {
		hs := NewHashMap()
		hs.Set("Key1", "Value1")

		result := hs.GetAll()
		result["Key1"] = "Modified"
		result["Key2"] = "New"

		original := hs.GetAll()
		assert.Equal(t, "Value1", original["Key1"])
		_, exists := original["Key2"]
		assert.False(t, exists)
	})

	t.Run("returns all key-value pairs after migration - hash encoding", func(t *testing.T) {
		hs := NewHashMap()
		for i := 0; i <= maxEntryCount; i++ {
			hs.Set(fmt.Sprintf("%d", i), fmt.Sprintf("val%d", i))
		}
		assert.NotNil(t, hs.hash)
		assert.Nil(t, hs.lp)

		result := hs.GetAll()
		assert.Len(t, result, maxEntryCount+1)
		for i := 0; i <= maxEntryCount; i++ {
			assert.Equal(t, fmt.Sprintf("val%d", i), result[fmt.Sprintf("%d", i)])
		}
	})

	t.Run("returns a copy, not the underlying storage - hash encoding", func(t *testing.T) {
		hs := NewHashMap()
		for i := 0; i <= maxEntryCount; i++ {
			hs.Set(fmt.Sprintf("%d", i), "value")
		}
		assert.NotNil(t, hs.hash)
		assert.Nil(t, hs.lp)

		result := hs.GetAll()
		result["newKey"] = "newValue"

		original := hs.GetAll()
		_, exists := original["newKey"]
		assert.False(t, exists)
	})
}

func Test_GetValue(t *testing.T) {
	hashSet := NewHashMap()
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
