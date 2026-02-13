package memory

import (
	"encoding/binary"
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

func (lp *listPack) push(values ...[]byte) int {
	lp.mu.Lock()
	defer lp.mu.Unlock()
	elemCount := int(binary.BigEndian.Uint16(lp.data[4:6]))
	newElemCount := elemCount + len(values)

	offset := int(binary.BigEndian.Uint32(lp.data[:4])) - 1
	for _, value := range values {
		offset = lp.encodeElement(lp.data, offset, value)
	}
	lp.data[offset] = 0xFF
	offset += 1
	binary.BigEndian.PutUint32(lp.data[:4], uint32(offset))
	binary.BigEndian.PutUint16(lp.data[4:6], uint16(newElemCount))
	return newElemCount
}

func (lp *listPack) encodeElement(buf []byte, offset int, element []byte) int {
	binary.BigEndian.PutUint16(buf[offset:offset+2], uint16(len(element)))
	offset += 2
	copy(buf[offset:offset+len(element)], element)
	offset += len(element)
	binary.BigEndian.PutUint16(buf[offset:offset+2], uint16(4+len(element)))
	return offset + 2
}

func (lp *listPack) pop(count int) [][]byte {
	count = min(count, lp.length())
	offset := int(binary.BigEndian.Uint32(lp.data[:4])) - 1
	elements := make([][]byte, count)
	for i := count - 1; i >= 0; i-- {
		// Start of total element size field
		offset -= 2
		elemSize := int(binary.BigEndian.Uint16(lp.data[offset : offset+2]))
		// Copy element data
		dataSize := elemSize - 4
		element := make([]byte, dataSize)
		copy(element, lp.data[offset-dataSize:offset])
		elements[i] = element
		// Start of element data
		offset -= dataSize + 2
	}
	lp.data[offset] = 0xFF
	offset += 1
	length := lp.length() - count
	binary.BigEndian.PutUint32(lp.data[0:4], uint32(offset))
	binary.BigEndian.PutUint16(lp.data[4:6], uint16(length))
	return elements
}
