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
	waiter          *listWaiter
}

// NewListMemoryStore creates a ListMemoryStore with given maxListPackSize
func NewListMemoryStore(maxListPackSize int) *ListMemoryStore {
	return &ListMemoryStore{
		mu:              sync.RWMutex{},
		lists:           make(map[string]*quickList),
		maxListPackSize: maxListPackSize,
		waiter:          newListWaiter(),
	}
}

func (l *ListMemoryStore) BlPop(ctx context.Context, keys []string) <-chan lists.ListNameToItem {
	resultCh := make(chan lists.ListNameToItem, 1)
	entry := &waitEntry{
		keys:     keys,
		resultCh: resultCh,
		served:   make(chan struct{}),
	}

	// Register and do initial scan under write lock so that no concurrent
	// LPush/RPush can slip between registration and the check.
	l.mu.Lock()
	l.waiter.register(entry)
	for _, key := range keys {
		list, ok := l.lists[key]
		if !ok {
			continue
		}
		elements, _ := list.lPop(1)
		if len(elements) > 0 {
			l.waiter.deregister(entry)
			l.mu.Unlock()
			resultCh <- lists.ListNameToItem{Key: key, Value: elements[0]}
			close(entry.served)
			return resultCh
		}
	}
	l.mu.Unlock()

	// Goroutine handles only timeout/cancellation.
	// Serving is done directly by LPush/RPush when they find a waiting entry.
	go func() {
		defer close(resultCh)
		select {
		case <-ctx.Done():
			l.mu.Lock()
			l.waiter.deregister(entry)
			l.mu.Unlock()
		case <-entry.served:
			// Result already delivered to resultCh by LPush/RPush.
		}
	}()

	return resultCh
}

// tryServeWaiter checks if a BLPOP client is waiting for key; if so, pops one
// element and returns the entry and value so the caller can deliver outside the lock.
// Must be called under l.mu.Lock().
func (l *ListMemoryStore) tryServeWaiter(key string) (*waitEntry, []byte) {
	entry := l.waiter.firstWaiter(key)
	if entry == nil {
		return nil, nil
	}
	list, ok := l.lists[key]
	if !ok {
		return nil, nil
	}
	elements, _ := list.lPop(1)
	if len(elements) == 0 {
		return nil, nil
	}
	l.waiter.deregister(entry)
	return entry, elements[0]
}

// LPush add the given values at the head of quicklist specified by the given key.
// If key is not present a new quicklist entry is created first.
func (l *ListMemoryStore) LPush(ctx context.Context, key string, values ...[]byte) (int, error) {
	l.mu.Lock()
	list, ok := l.lists[key]
	if !ok {
		list = newQuickList(l.maxListPackSize)
		l.lists[key] = list
	}
	length := list.lPush(values)
	entry, value := l.tryServeWaiter(key)
	l.mu.Unlock()

	// Deliver outside the lock: resultCh has buffer=1 and entry is exclusively
	// ours after deregister, so this send never blocks.
	if entry != nil {
		entry.resultCh <- lists.ListNameToItem{Key: key, Value: value}
		close(entry.served)
	}
	return length, nil
}

// RPush add the given values at the end of quicklist specified by the given key.
// If key is not present a new quicklist entry is created first.
func (l *ListMemoryStore) RPush(ctx context.Context, key string, values ...[]byte) (int, error) {
	l.mu.Lock()
	list, ok := l.lists[key]
	if !ok {
		list = newQuickList(l.maxListPackSize)
		l.lists[key] = list
	}
	length := list.rPush(values)
	entry, value := l.tryServeWaiter(key)
	l.mu.Unlock()

	if entry != nil {
		entry.resultCh <- lists.ListNameToItem{Key: key, Value: value}
		close(entry.served)
	}
	return length, nil
}

// LPop remove the given number of values from the head of quicklist specified by the given key.
// If key is not present a nil slice is returned.
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

// RPop remove the given number of values from the end of quicklist specified by the given key.
// If key is not present a nil slice is returned.
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

// LIndex finds an element in listPack from left side.
// Returns nil if there is no element at the given index.
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

// Len returns the count of elements in the list at the given key.
func (l *ListMemoryStore) Len(ctx context.Context, key string) (int, error) {
	l.mu.RLock()
	list, ok := l.lists[key]
	l.mu.RUnlock()
	if !ok {
		return 0, nil
	}
	return list.length(), nil
}

func (l *ListMemoryStore) LRange(ctx context.Context, key string, start, end int64) ([][]byte, error) {
	l.mu.RLock()
	ql, ok := l.lists[key]
	l.mu.RUnlock()
	if !ok {
		return [][]byte{}, nil
	}
	return ql.lRange(start, end), nil
}

func (l *ListMemoryStore) LMove(
	ctx context.Context,
	source, destination string,
	sourceDirection, destinationDirection lists.Direction,
) ([]byte, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	sList, ok := l.lists[source]
	if !ok {
		return nil, nil
	}
	var poppedElements [][]byte
	if sourceDirection == lists.Left {
		poppedElements, _ = sList.lPop(1)
	} else {
		poppedElements, _ = sList.rPop(1)
	}

	if len(poppedElements) == 0 {
		return nil, nil
	}

	dList, ok := l.lists[destination]
	if !ok {
		dList = newQuickList(l.maxListPackSize)
		l.lists[destination] = dList
	}
	if destinationDirection == lists.Left {
		dList.lPush(poppedElements)
	} else {
		dList.rPush(poppedElements)
	}

	return poppedElements[0], nil
}
