package memory

import (
	"encoding/binary"
	"fmt"
	"sync"
)

// listPack represents a list pack data structure used for storing small lists in memory.
type listPack struct {
	mu      sync.RWMutex
	data    []byte
	maxSize int
}

func newEmptyListPack(maxSize int) *listPack {
	data := make([]byte, maxSize)
	binary.BigEndian.PutUint32(data[:4], 7)
	binary.BigEndian.PutUint16(data[4:6], 0)
	data[6] = 0xFF
	return &listPack{data: data, mu: sync.RWMutex{}, maxSize: maxSize}
}

func newListPack(maxSize int, elements ...[]byte) *listPack {
	lp := newEmptyListPack(maxSize)
	lp.byteSize()
	lp.push(elements...)
	return lp
}

func (lp *listPack) length() int {
	lp.mu.RLock()
	defer lp.mu.RUnlock()
	return int(binary.BigEndian.Uint16(lp.data[4:6]))
}

func (lp *listPack) isFull() bool {
	lp.mu.RLock()
	defer lp.mu.RUnlock()
	size := int(binary.BigEndian.Uint32(lp.data[0:4]))
	return (size*100)/lp.maxSize >= 95
}

func (lp *listPack) byteSize() int {
	lp.mu.RLock()
	defer lp.mu.RUnlock()
	return int(binary.BigEndian.Uint32(lp.data[:4]))
}

func (lp *listPack) push(values ...[]byte) (int, error) {
	lp.mu.Lock()
	defer lp.mu.Unlock()
	elemCount := int(binary.BigEndian.Uint16(lp.data[4:6]))
	newElemCount := elemCount + len(values)

	offset := int(binary.BigEndian.Uint32(lp.data[:4])) - 1
	for _, value := range values {
		newOffset, err := encode(lp.data, offset, value)
		if err != nil {
			return elemCount, err
		}
		offset = newOffset
	}
	lp.data[offset] = 0xFF
	offset += 1
	binary.BigEndian.PutUint32(lp.data[:4], uint32(offset))
	binary.BigEndian.PutUint16(lp.data[4:6], uint16(newElemCount))
	return newElemCount, nil
}

func (lp *listPack) pop(count int) [][]byte {
	count = min(count, lp.length())
	// offset points to the 0xFF terminator in lp.data (absolute).
	// lp.data[6:] holds the encoded entries; the last backLen byte is one
	// position before 0xFF, adjusted for the 6-byte header: (stored_size-1) - 1 - 6.
	storedSize := int(binary.BigEndian.Uint32(lp.data[:4]))
	lastBackLenOffset := storedSize - 2 - 6 // relative to lp.data[6:]

	elements := make([][]byte, count)
	remaining := count
	cursor, _ := traverseReverse(lp.data[6:], lastBackLenOffset, func(elem interface{}) (bool, error) {
		remaining--
		switch elem.(type) {
		case []byte:
			elements[remaining] = elem.([]byte)
		case int:
			elements[remaining] = []byte(fmt.Sprintf("%d", elem.(int)))
		}
		if remaining == 0 {
			return false, nil
		}
		return true, nil
	})

	// cursor is in lp.data[6:] coords and points to the last backLen byte of
	// the last kept entry (or -1 if all entries were consumed).
	// The new 0xFF goes at the absolute position right after that.
	var newTerminator int
	if cursor < 0 {
		newTerminator = 6 // no entries remain; 0xFF goes right after header
	} else {
		newTerminator = cursor + 6 + 1 // absolute: skip header (6) + one past cursor
	}
	lp.data[newTerminator] = 0xFF
	binary.BigEndian.PutUint32(lp.data[0:4], uint32(newTerminator+1))
	binary.BigEndian.PutUint16(lp.data[4:6], uint16(lp.length()-count))
	return elements
}
