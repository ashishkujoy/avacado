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
