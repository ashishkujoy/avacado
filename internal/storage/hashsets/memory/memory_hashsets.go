package memory

import (
	"context"
	"sync"
)

type HashMaps struct {
	mu   sync.Mutex
	maps map[string]*HashMap
}

func NewHashMaps() *HashMaps {
	return &HashMaps{
		mu:   sync.Mutex{},
		maps: make(map[string]*HashMap),
	}
}

// HSet sets given fields to the specified map
func (h *HashMaps) HSet(_ context.Context, name string, keyValues []string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	hMap, found := h.maps[name]
	if !found {
		hMap = NewHashMap()
		h.maps[name] = hMap
	}
	for i := 0; i < len(keyValues); i += 2 {
		hMap.Set(keyValues[i], keyValues[i+1])
	}
}
