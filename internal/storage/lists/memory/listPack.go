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
	for _, e := range elements {
		_, _ = lp.push(e)
	}
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

func (lp *listPack) push(value []byte) (int, error) {
	lp.mu.Lock()
	defer lp.mu.Unlock()
	elemCount := int(binary.BigEndian.Uint16(lp.data[4:6]))

	offset := int(binary.BigEndian.Uint32(lp.data[:4])) - 1
	newOffset, err := encode(lp.data, offset, value)
	if err != nil {
		return elemCount, err
	}
	lp.data[newOffset] = 0xFF
	newOffset++
	binary.BigEndian.PutUint32(lp.data[:4], uint32(newOffset))
	binary.BigEndian.PutUint16(lp.data[4:6], uint16(elemCount+1))
	return elemCount + 1, nil
}

func (lp *listPack) pop() []byte {
	storedSize := int(binary.BigEndian.Uint32(lp.data[:4]))
	lastBackLenOffset := storedSize - 2 - 6

	var element []byte
	cursor, _ := traverseReverse(lp.data[6:], lastBackLenOffset, func(elem interface{}) (bool, error) {
		switch elem.(type) {
		case []byte:
			element = elem.([]byte)
		case int:
			element = []byte(fmt.Sprintf("%d", elem.(int)))
		}
		return false, nil
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
	binary.BigEndian.PutUint16(lp.data[4:6], uint16(lp.length()-1))
	return element
}

func (lp *listPack) lPush(value []byte) (int, error) {
	lp.mu.Lock()
	defer lp.mu.Unlock()
	elemCount := int(binary.BigEndian.Uint16(lp.data[4:6]))

	bytesSize := int(binary.BigEndian.Uint32(lp.data[:4]))
	tempBuf := make([]byte, lp.maxSize-bytesSize)
	newOffset, err := encode(tempBuf, 0, value)
	if err != nil {
		return elemCount, err
	}

	// Shift existing data to the right to make room for the new entry at the beginning.
	copy(lp.data[6+newOffset:], lp.data[6:bytesSize])
	// Copy new entry from tempBuf to the beginning of the list pack data area.
	copy(lp.data[6:], tempBuf[:newOffset])
	binary.BigEndian.PutUint32(lp.data[:4], uint32(bytesSize+newOffset))
	binary.BigEndian.PutUint16(lp.data[4:6], uint16(elemCount+1))
	return elemCount + 1, nil
}

func (lp *listPack) lPop(count int) [][]byte {
	lp.mu.Lock()
	defer lp.mu.Unlock()
	byteSize := int(binary.BigEndian.Uint32(lp.data[:4]))
	length := int(binary.BigEndian.Uint16(lp.data[4:6]))
	count = min(count, length)
	elements := make([][]byte, count)
	index := 0
	cursor, _ := traverse(lp.data[6:], 0, func(elem interface{}) (bool, error) {
		switch elem.(type) {
		case []byte:
			elements[index] = elem.([]byte)
		case int:
			elements[index] = []byte(fmt.Sprintf("%d", elem.(int)))
		}
		index++
		if index >= count {
			return false, nil
		}
		return true, nil
	})

	copy(lp.data[6:], lp.data[6+cursor:byteSize])
	newSize := byteSize - cursor
	newLength := length - count
	binary.BigEndian.PutUint32(lp.data[:4], uint32(newSize))
	binary.BigEndian.PutUint16(lp.data[4:6], uint16(newLength))
	return elements
}

func (lp *listPack) isEmpty() bool {
	lp.mu.RLock()
	defer lp.mu.RUnlock()

	size := binary.BigEndian.Uint32(lp.data[:4])
	return size == 7
}

func isLargerThanListPackSize(value []byte, maxListPackSize int) bool {
	return encodedSize(value)+7 > maxListPackSize
}

func newPlainListPack(element []byte) *listPack {
	size := 6 + encodedSize(element) + 1
	return newListPack(size, element)
}
