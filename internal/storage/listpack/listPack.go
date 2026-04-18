package listpack

import (
	"encoding/binary"
	"errors"
	"fmt"
	"sync"
)

var listPackNotEnoughSizeError = errors.New("not enough space in lp")

// ListPack represents a list pack data structure used for storing small lists in memory.
type ListPack struct {
	mu      sync.RWMutex
	data    []byte
	maxSize int
}

func NewEmptyListPack(maxSize int) *ListPack {
	data := make([]byte, maxSize)
	binary.BigEndian.PutUint32(data[:4], 7)
	binary.BigEndian.PutUint16(data[4:6], 0)
	data[6] = 0xFF
	return &ListPack{data: data, mu: sync.RWMutex{}, maxSize: maxSize}
}

func NewListPack(maxSize int, elements ...[]byte) *ListPack {
	lp := NewEmptyListPack(maxSize)
	for _, e := range elements {
		_, _ = lp.Push(e)
	}
	return lp
}

func (lp *ListPack) Length() int {
	lp.mu.RLock()
	defer lp.mu.RUnlock()
	return int(binary.BigEndian.Uint16(lp.data[4:6]))
}

func (lp *ListPack) IsFull() bool {
	lp.mu.RLock()
	defer lp.mu.RUnlock()
	size := int(binary.BigEndian.Uint32(lp.data[0:4]))
	return (size*100)/lp.maxSize >= 95
}

func (lp *ListPack) ByteSize() int {
	lp.mu.RLock()
	defer lp.mu.RUnlock()
	return int(binary.BigEndian.Uint32(lp.data[:4]))
}

func (lp *ListPack) Push(value []byte) (int, error) {
	lp.mu.Lock()
	defer lp.mu.Unlock()
	return lp._push(value)
}

func (lp *ListPack) _push(value []byte) (int, error) {
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

func (lp *ListPack) Pop() []byte {
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
	binary.BigEndian.PutUint16(lp.data[4:6], uint16(lp.Length()-1))
	return element
}

func (lp *ListPack) LPush(value []byte) (int, error) {
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

func (lp *ListPack) LPop(count int) [][]byte {
	lp.mu.Lock()
	defer lp.mu.Unlock()
	byteSize := int(binary.BigEndian.Uint32(lp.data[:4]))
	length := int(binary.BigEndian.Uint16(lp.data[4:6]))
	count = min(count, length)
	elements := make([][]byte, count)
	index := 0
	cursor, _ := traverse(lp.data, func(elem interface{}, _, _, _ int) (bool, error) {
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

	copy(lp.data[6:], lp.data[cursor:byteSize])
	newSize := byteSize - cursor + 6
	newLength := length - count
	binary.BigEndian.PutUint32(lp.data[:4], uint32(newSize))
	binary.BigEndian.PutUint16(lp.data[4:6], uint16(newLength))
	return elements
}

func (lp *ListPack) IsEmpty() bool {
	lp.mu.RLock()
	defer lp.mu.RUnlock()

	size := binary.BigEndian.Uint16(lp.data[4:6])
	return size == 0
}

func (lp *ListPack) AtIndex(i int) ([]byte, bool) {
	length := lp.Length()
	if i < 0 {
		i = i + length
	}
	if i >= length {
		return nil, false
	}
	r := 0
	var element []byte
	_, _ = traverse(lp.data, func(elem interface{}, _, _, _ int) (bool, error) {
		if r != i {
			r++
			return true, nil
		}
		switch elem.(type) {
		case []byte:
			element = elem.([]byte)
		case int:
			element = []byte(fmt.Sprintf("%d", elem.(int)))
		}
		return false, nil
	})
	return element, element != nil
}

// IndexOf finds the index of given candidate in the list pack
// Optionally skips odd element if skipOdds is set true(used in ListPackBasedHSet)
func (lp *ListPack) IndexOf(candidate string, skipOdds bool) (int, bool) {
	index := 0
	found := false
	_, _ = traverse(lp.data, func(elem interface{}, _, _, _ int) (bool, error) {
		defer func() { index++ }()
		if index%2 == 1 && skipOdds {
			return true, nil
		}
		var currentElement string
		switch elem.(type) {
		case []byte:
			currentElement = string(elem.([]byte))
		case int:
			currentElement = fmt.Sprintf("%d", elem.(int))
		}
		if currentElement == candidate {
			found = true
			return false, nil
		}
		return true, nil
	})
	if !found {
		return -1, false
	}
	return index - 1, found
}

func IsLargerThanListPackSize(value []byte, maxListPackSize int) bool {
	return encodedSize(value)+7 > maxListPackSize
}

func NewPlainListPack(element []byte) *ListPack {
	size := 6 + encodedSize(element) + 1
	return NewListPack(size, element)
}

func (lp *ListPack) LRange(start, end int64) ([][]byte, error) {
	lp.mu.Lock()
	defer lp.mu.Unlock()
	index := int64(0)
	var elements [][]byte
	_, err := traverse(lp.data, func(elem interface{}, _, _, _ int) (bool, error) {
		if index < start {
			index++
			return true, nil
		}
		if index > end {
			return false, nil
		}
		var element []byte
		switch elem.(type) {
		case []byte:
			element = elem.([]byte)
		case int:
			element = []byte(fmt.Sprintf("%d", elem.(int)))
		}
		elements = append(elements, element)
		index++
		return true, nil
	})
	return elements, err
}

func (lp *ListPack) Traverse(cb func(interface{}, int, int, int) (bool, error)) error {
	_, err := traverse(lp.data, cb)
	return err
}

// InsertAt insert the given element at the given index.
// negative index will result in error.
// Any index greater than or equal to length of listpack will be clamped to listpack length
func (lp *ListPack) InsertAt(i int, bytes []byte) error {
	if i < 0 {
		return errors.New("negative index not supported")
	}
	lp.mu.Lock()
	defer lp.mu.Unlock()
	oldSize := binary.BigEndian.Uint32(lp.data[:4])
	oldEntryCount := binary.BigEndian.Uint16(lp.data[4:6])
	if i > int(oldEntryCount) {
		i = int(oldEntryCount)
	}
	byteIndex := 6
	var err error
	if i == int(oldEntryCount) {
		byteIndex = int(oldSize) - 1
	} else if i != 0 {
		index := 0
		byteIndex, err = traverse(lp.data, func(element interface{}, _, _, _ int) (bool, error) {
			index++
			if index == i {
				return false, nil
			}
			return true, nil
		})
		if err != nil {
			return err
		}
	}
	size := encodedSize(bytes)
	encoded := make([]byte, size)
	_, err = encode(encoded, 0, bytes)
	if err != nil {
		return err
	}
	copy(lp.data[byteIndex+size:], lp.data[byteIndex:])
	copy(lp.data[byteIndex:], encoded)
	binary.BigEndian.PutUint32(lp.data[:4], oldSize+uint32(size))
	binary.BigEndian.PutUint16(lp.data[4:6], oldEntryCount+1)
	return nil
}

func (lp *ListPack) ReplaceAt(i int, element []byte) error {
	lp.mu.Lock()
	defer lp.mu.Unlock()
	start, end, err := getStartAndEndPositionOf(lp.data, i)
	if err != nil {
		return err
	}
	newSize := encodedSize(element)
	oldSize := end - start + 1
	switch {
	case newSize == oldSize:
		_, err = encode(lp.data, start, element)
		return err
	case newSize < oldSize:
		return lp.replaceWithShrink(start, end, element)
	default:
		return lp.replaceWithGrow(start, end, newSize, element)
	}
}

func (lp *ListPack) replaceWithShrink(start, end int, element []byte) error {
	total := int(binary.BigEndian.Uint32(lp.data[:4]))
	next, err := encode(lp.data, start, element)
	if err != nil {
		return err
	}
	// Shift tail left to close the gap left by the smaller replacement.
	copy(lp.data[next:], lp.data[end+1:total])
	binary.BigEndian.PutUint32(lp.data[:4], uint32(total-(end+1-next)))
	return nil
}

func (lp *ListPack) replaceWithGrow(start, end, newSize int, element []byte) error {
	total := int(binary.BigEndian.Uint32(lp.data[:4]))
	oldSize := end - start + 1
	diff := newSize - oldSize
	if total+diff > lp.maxSize {
		return errors.New("listpack overflow: not enough capacity to grow element")
	}
	// Shift tail right to make room, then encode the larger element in place.
	copy(lp.data[start+newSize:], lp.data[end+1:total])
	_, err := encode(lp.data, start, element)
	binary.BigEndian.PutUint32(lp.data[:4], uint32(total+diff))
	return err
}

func (lp *ListPack) PushAllOrNone(entries ...[]byte) (int, error) {
	lp.mu.Lock()
	defer lp.mu.Unlock()
	currentLpSize := int(binary.BigEndian.Uint32(lp.data[:4]))
	size := 0
	for _, entry := range entries {
		size += encodedSize(entry)
	}
	if currentLpSize+size > lp.maxSize {
		return -1, listPackNotEnoughSizeError
	}
	lastEntryCount := 0
	for _, entry := range entries {
		lastEntryCount, _ = lp._push(entry)
	}
	return lastEntryCount, nil
}

func (lp *ListPack) EncodedSize(entry []byte) int {
	return encodedSize(entry)
}

func (lp *ListPack) DeleteFromIndex(startingIndex int, count int) {
	lp.mu.Lock()
	defer lp.mu.Unlock()
	bytesSize := int(binary.BigEndian.Uint32(lp.data[:4]))
	size := int(binary.BigEndian.Uint16(lp.data[4:6]))
	endIndex := startingIndex + count - 1
	if endIndex >= size {
		endIndex = size - 1
	}
	start := 0
	end := 0

	_, _ = traverse(lp.data, func(_ interface{}, index, startPosition, endPosition int) (bool, error) {
		if index == startingIndex {
			start = startPosition
		}
		if index == endIndex {
			end = endPosition
			return false, nil
		}
		return true, nil
	})

	copy(lp.data[start:], lp.data[end+1:bytesSize])
	bytesSize = bytesSize - (end - start)
	removedCount := endIndex - startingIndex + 1
	binary.BigEndian.PutUint32(lp.data[:4], uint32(bytesSize))
	binary.BigEndian.PutUint16(lp.data[4:6], uint16(size-removedCount))
}
