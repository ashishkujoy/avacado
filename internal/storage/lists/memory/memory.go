package memory

import (
	"context"
	"sync"
)

// ListMemoryStore represent an in-memory implementation of Lists interface.
// It's a key value store where each value is a quicklist.
type ListMemoryStore struct {
	mu              sync.RWMutex
	lists           map[string]*quickList
	maxListPackSize int
}

// NewListMemoryStore creates a ListMemoryStore with given maxListPackSize
func NewListMemoryStore(maxListPackSize int) *ListMemoryStore {
	return &ListMemoryStore{
		mu:              sync.RWMutex{},
		lists:           make(map[string]*quickList),
		maxListPackSize: maxListPackSize,
	}
}

// RPush add the given values at the end of quicklist specified by the given key
// If key is not present a new quicklist entry is created first, followed by
// pushing elements.
// RPush try to manually release the read or write lock to allow maximum possible
// concurrency.
func (l *ListMemoryStore) RPush(ctx context.Context, key string, values ...[]byte) (int, error) {
	l.mu.RLock()

	list, ok := l.lists[key]
	if !ok {
		list = newQuickList(l.maxListPackSize)
		// Release read lock, need to get a write lock
		l.mu.RUnlock()
		l.mu.Lock()
		l.lists[key] = list
		// Release write lock
		l.mu.Unlock()
	} else {
		// Release read lock
		l.mu.RUnlock()
	}
	length := list.rPush(values)
	return length, nil
}

// RPop remove the given number of values from the end of quicklist specified by the given key
// If key is not present a nil slice is return
// RPop try to manually release the read or write lock to allow maximum possible
// concurrency.
func (l *ListMemoryStore) RPop(ctx context.Context, key string, count int) ([][]byte, error) {
	l.mu.RLock()
	list, ok := l.lists[key]
	l.mu.RUnlock()
	if !ok {
		return nil, nil
	}
	elements, _ := list.rPop(count)
	return elements, nil
}

// Len return the count of elements present in listpack of given key
func (l *ListMemoryStore) Len(ctx context.Context, key string) (int, error) {
	l.mu.RLock()
	list, ok := l.lists[key]
	l.mu.RUnlock()
	if !ok {
		return 0, nil
	}
	return list.length(), nil
}
