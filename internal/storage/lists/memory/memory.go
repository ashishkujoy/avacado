package memory

import (
	"avacado/internal/storage/lists"
	"context"
)

// ListMemoryStore represent an in-memory implementation of Lists interface.
// It's a key value store where each value is a quicklist.
// All methods are called exclusively by the executor goroutine — no locking needed.
type ListMemoryStore struct {
	lists           map[string]*quickList
	maxListPackSize int
}

// NewListMemoryStore creates a ListMemoryStore with given maxListPackSize
func NewListMemoryStore(maxListPackSize int) *ListMemoryStore {
	return &ListMemoryStore{
		lists:           make(map[string]*quickList),
		maxListPackSize: maxListPackSize,
	}
}

// LPush add the given values at the head of quicklist specified by the given key.
// If key is not present a new quicklist entry is created first.
func (l *ListMemoryStore) LPush(ctx context.Context, key string, values ...[]byte) (int, error) {
	list, ok := l.lists[key]
	if !ok {
		list = newQuickList(l.maxListPackSize)
		l.lists[key] = list
	}
	length := list.lPush(values)
	return length, nil
}

// RPush add the given values at the end of quicklist specified by the given key.
// If key is not present a new quicklist entry is created first.
func (l *ListMemoryStore) RPush(ctx context.Context, key string, values ...[]byte) (int, error) {
	list, ok := l.lists[key]
	if !ok {
		list = newQuickList(l.maxListPackSize)
		l.lists[key] = list
	}
	length := list.rPush(values)
	return length, nil
}

// LPop remove the given number of values from the head of quicklist specified by the given key.
// If key is not present a nil slice is returned.
func (l *ListMemoryStore) LPop(ctx context.Context, key string, count int) ([][]byte, error) {
	list, ok := l.lists[key]
	if !ok {
		return nil, nil
	}
	elements, _ := list.lPop(count)
	return elements, nil
}

// RPop remove the given number of values from the end of quicklist specified by the given key.
// If key is not present a nil slice is returned.
func (l *ListMemoryStore) RPop(ctx context.Context, key string, count int) ([][]byte, error) {
	list, ok := l.lists[key]
	if !ok {
		return nil, nil
	}
	elements, _ := list.rPop(count)
	return elements, nil
}

// LIndex finds an element in ListPack from left side.
// Returns nil if there is no element at the given index.
func (l *ListMemoryStore) LIndex(ctx context.Context, key string, index int) ([]byte, error) {
	list, found := l.lists[key]
	if !found {
		return nil, nil
	}
	element, _ := list.atIndex(index)
	return element, nil
}

// Len returns the count of elements in the list at the given key.
func (l *ListMemoryStore) Len(ctx context.Context, key string) (int, error) {
	list, ok := l.lists[key]
	if !ok {
		return 0, nil
	}
	return list.length(), nil
}

func (l *ListMemoryStore) LRange(ctx context.Context, key string, start, end int64) ([][]byte, error) {
	ql, ok := l.lists[key]
	if !ok {
		return [][]byte{}, nil
	}
	return ql.lRange(start, end), nil
}

func (l *ListMemoryStore) LMove(
	_ context.Context,
	source, destination string,
	sourceDirection, destinationDirection lists.Direction,
) ([]byte, error) {
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
