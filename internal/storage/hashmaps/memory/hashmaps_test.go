package memory

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHashMaps_HSet(t *testing.T) {
	maps := NewHashMaps()
	maps.HSet(context.Background(), "map1", []string{"key1", "V1", "key2", "V2"})

	assert.Equal(t, 1, len(maps.maps))
	assert.Equal(t, 2, maps.maps["map1"].Size())
}

func TestHashMaps_HGet(t *testing.T) {
	maps := NewHashMaps()
	ctx := context.Background()
	maps.HSet(ctx, "map1", []string{"key1", "V1", "key2", "V2"})

	_, err := maps.HGet(ctx, "non-existing-map", "key1")
	assert.Error(t, err)

	_, err = maps.HGet(ctx, "map1", "non-existing-field")
	assert.Error(t, err)

	value, err := maps.HGet(ctx, "map1", "key1")
	assert.NoError(t, err)
	assert.Equal(t, "V1", string(value))
}

func TestHashMaps_HGetAll(t *testing.T) {
	ctx := context.Background()

	t.Run("returns empty map for non-existing map name", func(t *testing.T) {
		maps := NewHashMaps()
		result, err := maps.HGetAll(ctx, "non-existing-map")
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Empty(t, result)
	})

	t.Run("returns all key-value pairs - listpack encoding", func(t *testing.T) {
		maps := NewHashMaps()
		maps.HSet(ctx, "map1", []string{"key1", "V1", "key2", "V2", "key3", "V3"})

		// confirm still listpack-encoded
		assert.Nil(t, maps.maps["map1"].hash)
		assert.NotNil(t, maps.maps["map1"].lp)

		result, err := maps.HGetAll(ctx, "map1")
		assert.NoError(t, err)
		assert.Equal(t, map[string]string{
			"key1": "V1",
			"key2": "V2",
			"key3": "V3",
		}, result)
	})

	t.Run("returns all key-value pairs - hash encoding", func(t *testing.T) {
		maps := NewHashMaps()
		kvs := make([]string, 0, (maxEntryCount+1)*2)
		for i := 0; i <= maxEntryCount; i++ {
			kvs = append(kvs, fmt.Sprintf("key%d", i), fmt.Sprintf("val%d", i))
		}
		maps.HSet(ctx, "map1", kvs)

		// confirm migrated to hash encoding
		assert.NotNil(t, maps.maps["map1"].hash)
		assert.Nil(t, maps.maps["map1"].lp)

		result, err := maps.HGetAll(ctx, "map1")
		assert.NoError(t, err)
		assert.Len(t, result, maxEntryCount+1)
		for i := 0; i <= maxEntryCount; i++ {
			assert.Equal(t, fmt.Sprintf("val%d", i), result[fmt.Sprintf("key%d", i)])
		}
	})
}

func TestHashMaps_HDel(t *testing.T) {
	ctx := context.Background()

	t.Run("returns 0 for non-existing map", func(t *testing.T) {
		maps := NewHashMaps()
		deleted, err := maps.HDel(ctx, "non-existing-map", []string{"key1"})
		assert.NoError(t, err)
		assert.Equal(t, 0, deleted)
	})

	t.Run("deletes existing fields - listpack encoding", func(t *testing.T) {
		maps := NewHashMaps()
		maps.HSet(ctx, "map1", []string{"key1", "V1", "key2", "V2", "key3", "V3"})

		assert.Nil(t, maps.maps["map1"].hash)
		assert.NotNil(t, maps.maps["map1"].lp)

		deleted, err := maps.HDel(ctx, "map1", []string{"key1", "key3"})
		assert.NoError(t, err)
		assert.Equal(t, 2, deleted)

		_, err = maps.HGet(ctx, "map1", "key1")
		assert.Error(t, err)
		_, err = maps.HGet(ctx, "map1", "key3")
		assert.Error(t, err)
		value, err := maps.HGet(ctx, "map1", "key2")
		assert.NoError(t, err)
		assert.Equal(t, "V2", string(value))
	})

	t.Run("returns 0 for non-existing fields - listpack encoding", func(t *testing.T) {
		maps := NewHashMaps()
		maps.HSet(ctx, "map1", []string{"key1", "V1"})

		deleted, err := maps.HDel(ctx, "map1", []string{"missing"})
		assert.NoError(t, err)
		assert.Equal(t, 0, deleted)
	})

	t.Run("deletes existing fields - hash encoding", func(t *testing.T) {
		maps := NewHashMaps()
		kvs := make([]string, 0, (maxEntryCount+1)*2)
		for i := 0; i <= maxEntryCount; i++ {
			kvs = append(kvs, fmt.Sprintf("key%d", i), fmt.Sprintf("val%d", i))
		}
		maps.HSet(ctx, "map1", kvs)

		assert.NotNil(t, maps.maps["map1"].hash)
		assert.Nil(t, maps.maps["map1"].lp)

		deleted, err := maps.HDel(ctx, "map1", []string{"key0", "key1"})
		assert.NoError(t, err)
		assert.Equal(t, 2, deleted)

		_, err = maps.HGet(ctx, "map1", "key0")
		assert.Error(t, err)
		_, err = maps.HGet(ctx, "map1", "key1")
		assert.Error(t, err)
		value, err := maps.HGet(ctx, "map1", "key2")
		assert.NoError(t, err)
		assert.Equal(t, "val2", string(value))
	})

	t.Run("returns 0 for non-existing fields - hash encoding", func(t *testing.T) {
		maps := NewHashMaps()
		kvs := make([]string, 0, (maxEntryCount+1)*2)
		for i := 0; i <= maxEntryCount; i++ {
			kvs = append(kvs, fmt.Sprintf("key%d", i), fmt.Sprintf("val%d", i))
		}
		maps.HSet(ctx, "map1", kvs)

		assert.NotNil(t, maps.maps["map1"].hash)

		deleted, err := maps.HDel(ctx, "map1", []string{"missing"})
		assert.NoError(t, err)
		assert.Equal(t, 0, deleted)
	})

	t.Run("deletes mix of existing and non-existing fields", func(t *testing.T) {
		maps := NewHashMaps()
		maps.HSet(ctx, "map1", []string{"key1", "V1", "key2", "V2"})

		deleted, err := maps.HDel(ctx, "map1", []string{"key1", "missing"})
		assert.NoError(t, err)
		assert.Equal(t, 1, deleted)

		_, err = maps.HGet(ctx, "map1", "key1")
		assert.Error(t, err)
		value, err := maps.HGet(ctx, "map1", "key2")
		assert.NoError(t, err)
		assert.Equal(t, "V2", string(value))
	})
}

