package memory

import "sync"

const defaultMaxListPackSize = 1024 * 8

// quickList represents a quick list data structure used for storing lists in memory.
type quickList struct {
	lps             []*listPack
	mu              sync.RWMutex
	maxListPackSize int
	size            int
}

func newQuickList(maxListPackSize int) *quickList {
	return &quickList{
		lps:             []*listPack{newEmptyListPack(maxListPackSize)},
		mu:              sync.RWMutex{},
		maxListPackSize: maxListPackSize,
		size:            0,
	}
}

func (ql *quickList) length() int {
	ql.mu.RLock()
	defer ql.mu.RUnlock()
	return ql.size
}

// lPush adds elements to the head of the quick list and returns the new length of the list.
func (ql *quickList) lPush(elements [][]byte) int {
	ql.mu.Lock()
	defer ql.mu.Unlock()
	for _, element := range elements {
		if isLargerThanListPackSize(element, ql.maxListPackSize) {
			lp := newPlainListPack(element)
			if ql.lps[0].length() == 0 {
				ql.lps[0] = lp
			} else {
				ql.lps = append([]*listPack{lp}, ql.lps...)
			}
			ql.size++
			return ql.size
		}
		_, err := ql.lps[0].lPush(element)
		if err != nil {
			lp := newEmptyListPack(ql.maxListPackSize)
			_, _ = lp.lPush(element)
			ql.lps = append([]*listPack{lp}, ql.lps...)
		}
		ql.size++
	}
	return ql.size
}

// rPush adds an element to the end of the quick list and returns the new length of the list.
func (ql *quickList) rPush(elements [][]byte) int {
	ql.mu.Lock()
	defer ql.mu.Unlock()
	for _, element := range elements {
		// if element is larger than maxListPackSize, we need to create a plain listPack for it
		if isLargerThanListPackSize(element, ql.maxListPackSize) {
			lp := newPlainListPack(element)
			if ql.lps[len(ql.lps)-1].length() == 0 {
				// Tail is empty — replace it rather than leaving an orphaned node.
				ql.lps[len(ql.lps)-1] = lp
			} else {
				ql.lps = append(ql.lps, lp)
			}
			ql.size++
			return ql.size
		}
		_, err := ql.lps[len(ql.lps)-1].push(element)
		if err != nil {
			lp := newEmptyListPack(ql.maxListPackSize)
			_, _ = lp.push(element)
			ql.lps = append(ql.lps, lp)
		}
		ql.size++
	}
	return ql.size
}

// lPop removes elements from the head of the quick list and returns the elements and new length of the list.
func (ql *quickList) lPop(count int) ([][]byte, int) {
	ql.mu.Lock()
	defer ql.mu.Unlock()
	if ql.size == 0 {
		return nil, 0
	}
	elements := make([][]byte, count)
	length := 0
	for ; length < count; length++ {
		head := ql.lps[0]
		if head.isEmpty() {
			break
		}
		popped := head.lPop(1)
		if len(popped) == 0 {
			break
		}
		elements[length] = popped[0]
		ql.size -= 1
		if head.isEmpty() && len(ql.lps) > 1 {
			ql.lps = ql.lps[1:]
		}
	}
	return elements[:length], ql.size
}

// rPop removes an element from the end of the quick list and return the element and new length of the list.
func (ql *quickList) rPop(count int) ([][]byte, int) {
	ql.mu.Lock()
	defer ql.mu.Unlock()
	if ql.size == 0 {
		return nil, 0
	}
	elements := make([][]byte, count)
	length := 0
	for ; length < count; length++ {
		tail := ql.lps[len(ql.lps)-1]
		if tail.isEmpty() {
			break
		}
		element := tail.pop()
		elements[length] = element
		ql.size -= 1
		if tail.isEmpty() && len(ql.lps) > 1 {
			ql.lps = ql.lps[:len(ql.lps)-1]
		}
	}
	return elements[:length], ql.size
}
