package memory

import (
	"context"
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
