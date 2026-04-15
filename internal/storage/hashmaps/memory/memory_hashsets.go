package memory

import (
	"context"
	"fmt"
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
func (h *HashMaps) HSet(_ context.Context, name string, keyValues []string) int {
	h.mu.Lock()
	defer h.mu.Unlock()

	hMap, found := h.maps[name]
	if !found {
		hMap = NewHashMap()
		h.maps[name] = hMap
	}
	addedCount := 0
	for i := 0; i < len(keyValues); i += 2 {
		addedCount += hMap.Set(keyValues[i], keyValues[i+1])
	}
	return addedCount
}

// HGet return the specified field of the map
func (h *HashMaps) HGet(_ context.Context, name, field string) ([]byte, error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	hMap, found := h.maps[name]
	if !found {
		return nil, fmt.Errorf("%s does not exists", name)
	}
	value, valueFound := hMap.Get(field)
	if !valueFound {
		return nil, fmt.Errorf("%s field does not exists in %s map", field, name)
	}
	return value, nil
}

func (h *HashMaps) HGetAll(_ context.Context, name string) (map[string]string, error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	hMap, found := h.maps[name]
	if !found {
		return make(map[string]string), nil
	}
	return hMap.GetAll(), nil
}
