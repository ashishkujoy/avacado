package memory

import (
	"avacado/internal/storage/lists"
	"context"
	"sync"
)

// ListMemoryStore represent an in-memory implementation of Lists interface.
// It's a key value store where each value is a quicklist.
type ListMemoryStore struct {
	mu              sync.RWMutex
	lists           map[string]*quickList
	maxListPackSize int
	notifier        ListAvailabilityNotifier
}

func (l *ListMemoryStore) BlPop(ctx context.Context, keys []string) <-chan lists.ListNameToItem {
	resultCh := make(chan lists.ListNameToItem, 1)
	notifyCh := l.notifier.AwaitFor(keys)

	go func() {
		defer close(resultCh)
		defer l.notifier.DeregisterClients(notifyCh, keys)

		select {
		case <-ctx.Done():
			return
		case key := <-notifyCh:
			l.mu.RLock()
			list, ok := l.lists[key]
			l.mu.RUnlock()
			if !ok {
				return
			}
			elements, _ := list.lPop(1)
			if len(elements) == 0 {
				return
			}
			resultCh <- lists.ListNameToItem{Key: key, Value: elements[0]}
		}
	}()

	return resultCh
}

// LIndex finds an element in listPact from left side.
// return nil byte slice if there is no element at given index
func (l *ListMemoryStore) LIndex(ctx context.Context, key string, index int) ([]byte, error) {
	l.mu.RLock()
	list, found := l.lists[key]
	l.mu.RUnlock()
	if !found {
		return nil, nil
	}
	element, _ := list.atIndex(index)
	return element, nil
}

// NewListMemoryStore creates a ListMemoryStore with given maxListPackSize
func NewListMemoryStore(maxListPackSize int) *ListMemoryStore {
	return &ListMemoryStore{
		mu:              sync.RWMutex{},
		lists:           make(map[string]*quickList),
		maxListPackSize: maxListPackSize,
		notifier:        *NewListAvailabilityNotifier(),
	}
}

// LPush add the given values at the head of quicklist specified by the given key
// If key is not present a new quicklist entry is created first, followed by
// pushing elements.
func (l *ListMemoryStore) LPush(ctx context.Context, key string, values ...[]byte) (int, error) {
	l.mu.RLock()

	list, ok := l.lists[key]
	if !ok {
		list = newQuickList(l.maxListPackSize)
		l.mu.RUnlock()
		l.mu.Lock()
		l.lists[key] = list
		l.mu.Unlock()
	} else {
		l.mu.RUnlock()
	}
	length := list.lPush(values)
	l.notifier.NotifyAvailable(key)
	return length, nil
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
	l.notifier.NotifyAvailable(key)
	return length, nil
}

// LPop remove the given number of values from the head of quicklist specified by the given key
// If key is not present a nil slice is returned
func (l *ListMemoryStore) LPop(ctx context.Context, key string, count int) ([][]byte, error) {
	l.mu.RLock()
	list, ok := l.lists[key]
	l.mu.RUnlock()
	if !ok {
		return nil, nil
	}
	elements, _ := list.lPop(count)
	return elements, nil
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
