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

// rPush adds an element to the end of the quick list and returns the new length of the list.
func (ql *quickList) rPush(element []byte) int {
	ql.mu.Lock()
	defer ql.mu.Unlock()
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
	return ql.size
}
