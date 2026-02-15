package memory

import (
	"encoding/binary"
	"errors"
	"fmt"
	"strconv"
)

var unknownEncodingError = errors.New("unknown encoding")
var unknownTypeError = errors.New("unknown type")

// Encoding type constants
const (
	Encoding13bitInt = 0xC0 // 110xxxxx + 1 byte
	Encoding16bitInt = 0xF1 // 11110001 + 2 bytes
	Encoding24bitInt = 0xF2 // 11110010 + 3 bytes
	Encoding32bitInt = 0xF3 // 11110011 + 4 bytes
	Encoding64bitInt = 0xF4 // 11110100 + 8 bytes
)

// lpEncodeBackLen encodes the backLen field which stores the total entry length
// (encoding + data) NOT including the backLen itself.
// BackLen is encoded from right to left with MSB as continuation bit.
// Returns the number of bytes used for backLen encoding.
func lpEncodeBackLen(buf []byte, offset int, entryLen uint64) int {
	if entryLen <= 127 {
		// 1 byte: 0xxxxxxx
		buf[offset] = byte(entryLen)
		return 1
	} else if entryLen <= 16383 {
		// 2 bytes: high byte, then low byte with MSB=1
		buf[offset] = byte(entryLen >> 7)
		buf[offset+1] = byte(entryLen&0x7F) | 0x80
		return 2
	} else if entryLen <= 2097151 {
		// 3 bytes
		buf[offset] = byte(entryLen >> 14)
		buf[offset+1] = byte((entryLen>>7)&0x7F) | 0x80
		buf[offset+2] = byte(entryLen&0x7F) | 0x80
		return 3
	} else if entryLen <= 268435455 {
		// 4 bytes
		buf[offset] = byte(entryLen >> 21)
		buf[offset+1] = byte((entryLen>>14)&0x7F) | 0x80
		buf[offset+2] = byte((entryLen>>7)&0x7F) | 0x80
		buf[offset+3] = byte(entryLen&0x7F) | 0x80
		return 4
	}

	// 5 bytes
	buf[offset] = byte(entryLen >> 28)
	buf[offset+1] = byte((entryLen>>21)&0x7F) | 0x80
	buf[offset+2] = byte((entryLen>>14)&0x7F) | 0x80
	buf[offset+3] = byte((entryLen>>7)&0x7F) | 0x80
	buf[offset+4] = byte(entryLen&0x7F) | 0x80
	return 5
}

// lpDecodeBackLen decodes the backLen by reading from right to left
// offset should point to the last byte of the backLen
// Returns: entryLen (not including backLen), backLenSize, error
func lpDecodeBackLen(buf []byte, offset int) (uint64, int, error) {
	if offset < 0 || offset >= len(buf) {
		return 0, 0, fmt.Errorf("invalid offset %d", offset)
	}

	var entryLen uint64
	bytesRead := 0
	shift := uint(0)

	// Read from right to left
	for {
		if offset-bytesRead < 0 {
			return 0, 0, fmt.Errorf("buffer underflow reading backLen")
		}

		b := buf[offset-bytesRead]
		bytesRead++

		if b&0x80 != 0 {
			// Continuation bit set, more bytes to read
			entryLen |= uint64(b&0x7F) << shift
			shift += 7
		} else {
			// No continuation bit, this is the last (leftmost) byte
			entryLen |= uint64(b) << shift
			break
		}

		if bytesRead >= 5 {
			break // Maximum 5 bytes for backLen
		}
	}

	return entryLen, bytesRead, nil
}

// encode determines the type of the element and encodes it into the buffer starting at the given offset.
func encode(buf []byte, offset int, element []byte) (int, error) {
	if v, err := strconv.Atoi(string(element)); err == nil {
		if v >= 0 && v <= 127 {
			return encode7BitInt(buf, offset, v), nil
		} else if v >= -4096 && v <= 4095 {
			return encode13BitInt(buf, offset, v), nil
		} else if v >= -32768 && v <= 32767 {
			return encode16BitInt(buf, offset, v), nil
		} else if v >= -8388608 && v <= 8388607 {
			return encode24BitInt(buf, offset, v), nil
		} else if v >= -16777216 && v <= 16777216 {
			return encode32BitInt(buf, offset, v), nil
		} else if v >= -2147483648 && v <= 2147483647 {
			return encode64BitInt(buf, offset, v), nil
		}
		return offset, unknownTypeError
	}
	return offset, unknownTypeError
}

// decode decodes an element from the buffer starting at the given offset.
func decode(buf []byte, offset int) (interface{}, int, error) {
	prefix := buf[offset]
	if prefix&0x80 == 0 {
		// 7-bit number
		return decode7BitInt(buf, offset)
	} else if prefix&0xE0 == 0xC0 {
		// 13-bit number
		return decode13BitInt(buf, offset)
	} else if prefix == 0xF1 {
		// 16-bit number
		return decode16BitInt(buf, offset)
	} else if prefix == 0xF2 {
		// 24-bit number
		return decode24BitInt(buf, offset)
	} else if prefix == 0xF3 {
		// 32-bit number
		return decode32BitInt(buf, offset)
	} else if prefix == 0xF4 {
		// 64-bit number
		return decode64BitInt(buf, offset)
	}
	return nil, offset, unknownEncodingError
}

// getBackLenSize calculates how many bytes are needed to encode a length
func getBackLenSize(length uint64) int {
	if length <= 127 {
		return 1
	} else if length <= 16383 {
		return 2
	} else if length <= 2097151 {
		return 3
	} else if length <= 268435455 {
		return 4
	}
	return 5
}

// encode7BitInt encodes a 7-bit unsigned integer (0-127) with backLen
// Format: 0xxxxxxx <backLen>
// Entry structure: encoding(1) + backLen
func encode7BitInt(buf []byte, offset int, n int) int {
	// Encode the value
	buf[offset] = byte(n) // High bit is 0 automatically
	offset++

	// Calculate entry length (just the encoding byte, not including backLen)
	entryLen := uint64(1)

	// Encode backLen
	backLenSize := lpEncodeBackLen(buf, offset, entryLen)
	offset += backLenSize

	return offset
}

// decode7BitInt decodes a 7-bit unsigned integer with backLen
// Returns: value, new offset, error
func decode7BitInt(buf []byte, offset int) (int, int, error) {
	if offset >= len(buf) {
		return 0, offset, fmt.Errorf("buffer overflow at offset %d", offset)
	}

	// Check if it's actually a 7-bit encoding (high bit should be 0)
	if buf[offset]&0x80 != 0 {
		return 0, offset, fmt.Errorf("not a 7-bit encoding at offset %d", offset)
	}

	value := int(buf[offset] & 0x7F)
	return value, offset + 2, nil
}

// encode13BitInt encodes a 13-bit signed integer (-4096 to 4095) with backLen
// Format: 110xxxxx xxxxxxxx <backLen> (big-endian value)
// Entry structure: encoding(2) + backLen
func encode13BitInt(buf []byte, offset int, n int) int {
	if n < -4096 || n > 4095 {
		panic(fmt.Sprintf("value %d out of range for 13-bit encoding (-4096 to 4095)", n))
	}

	// Convert to unsigned representation for bit manipulation
	var val uint16
	if n < 0 {
		val = uint16(8192 + n) // 2^13 + n for negative numbers
	} else {
		val = uint16(n)
	}

	// Big-endian: high 5 bits in first byte, low 8 bits in second byte
	buf[offset] = Encoding13bitInt | byte((val>>8)&0x1F) // 110xxxxx
	buf[offset+1] = byte(val & 0xFF)
	offset += 2

	// Calculate entry length (2 bytes for encoding)
	entryLen := uint64(2)

	// Encode backLen
	backLenSize := lpEncodeBackLen(buf, offset, entryLen)
	offset += backLenSize

	return offset
}

// decode13BitInt decodes a 13-bit signed integer with backLen
func decode13BitInt(buf []byte, offset int) (int, int, error) {
	if offset+1 >= len(buf) {
		return 0, offset, fmt.Errorf("buffer overflow at offset %d", offset)
	}

	// Check encoding type
	if buf[offset]&0xE0 != Encoding13bitInt {
		return 0, offset, fmt.Errorf("not a 13-bit encoding at offset %d", offset)
	}

	// Big-endian reconstruction
	high := uint16(buf[offset]&0x1F) << 8
	low := uint16(buf[offset+1])
	val := high | low

	// Convert from unsigned to signed
	var result int
	if val >= 4096 { // Negative number
		result = int(val) - 8192
	} else {
		result = int(val)
	}

	offset += 3
	return result, offset, nil
}

// encode16BitInt encodes a 16-bit signed integer (-32768 to 32767) with backLen
// Format: 0xF1 + 2 bytes (little-endian) + backLen
// Entry structure: encoding(1) + data(2) + backLen
func encode16BitInt(buf []byte, offset int, n int) int {
	if n < -32768 || n > 32767 {
		panic(fmt.Sprintf("value %d out of range for 16-bit encoding (-32768 to 32767)", n))
	}

	buf[offset] = Encoding16bitInt
	binary.LittleEndian.PutUint16(buf[offset+1:], uint16(n))
	offset += 3

	// Calculate entry length (1 + 2 = 3 bytes)
	entryLen := uint64(3)

	// Encode backLen
	backLenSize := lpEncodeBackLen(buf, offset, entryLen)
	offset += backLenSize

	return offset
}

// decode16BitInt decodes a 16-bit signed integer with backLen
func decode16BitInt(buf []byte, offset int) (int, int, error) {
	if offset+2 >= len(buf) {
		return 0, offset, fmt.Errorf("buffer overflow at offset %d", offset)
	}

	if buf[offset] != Encoding16bitInt {
		return 0, offset, fmt.Errorf("not a 16-bit encoding at offset %d", offset)
	}
	val := binary.LittleEndian.Uint16(buf[offset+1:])
	result := int(int16(val))
	offset += 4

	return result, offset, nil
}

// encode24BitInt encodes a 24-bit signed integer with backLen
// Format: 0xF2 + 3 bytes (little-endian) + backLen
// Entry structure: encoding(1) + data(3) + backLen
func encode24BitInt(buf []byte, offset int, n int) int {
	if n < -(1<<23) || n >= (1<<23) {
		panic(fmt.Sprintf("value %d out of range for 24-bit encoding", n))
	}

	buf[offset] = Encoding24bitInt

	// Little-endian: least significant byte first
	val := uint32(n) & 0xFFFFFF
	buf[offset+1] = byte(val & 0xFF)
	buf[offset+2] = byte((val >> 8) & 0xFF)
	buf[offset+3] = byte((val >> 16) & 0xFF)
	offset += 4

	// Calculate entry length (1 + 3 = 4 bytes)
	entryLen := uint64(4)

	// Encode backLen
	backLenSize := lpEncodeBackLen(buf, offset, entryLen)
	offset += backLenSize

	return offset
}

// decode24BitInt decodes a 24-bit signed integer with backLen
func decode24BitInt(buf []byte, offset int) (int, int, error) {
	if offset+3 >= len(buf) {
		return 0, offset, fmt.Errorf("buffer overflow at offset %d", offset)
	}

	if buf[offset] != Encoding24bitInt {
		return 0, offset, fmt.Errorf("not a 24-bit encoding at offset %d", offset)
	}

	// Little-endian reconstruction
	val := uint32(buf[offset+1]) |
		uint32(buf[offset+2])<<8 |
		uint32(buf[offset+3])<<16

	// Sign extension for 24-bit to 32-bit
	if val&0x800000 != 0 {
		val |= 0xFF000000
	}

	result := int(int32(val))
	offset += 4

	// Decode backLen to skip past it
	for offset < len(buf) {
		if buf[offset]&0x80 == 0 {
			offset++
			break
		}
		offset++
	}

	return result, offset, nil
}

// encode32BitInt encodes a 32-bit signed integer with backLen
// Format: 0xF3 + 4 bytes (little-endian) + backLen
// Entry structure: encoding(1) + data(4) + backLen
func encode32BitInt(buf []byte, offset int, n int) int {
	buf[offset] = Encoding32bitInt
	binary.LittleEndian.PutUint32(buf[offset+1:], uint32(n))
	offset += 5

	// Calculate entry length (1 + 4 = 5 bytes)
	entryLen := uint64(5)

	// Encode backLen
	backLenSize := lpEncodeBackLen(buf, offset, entryLen)
	offset += backLenSize

	return offset
}

// decode32BitInt decodes a 32-bit signed integer with backLen
func decode32BitInt(buf []byte, offset int) (int, int, error) {
	if offset+4 >= len(buf) {
		return 0, offset, fmt.Errorf("buffer overflow at offset %d", offset)
	}

	if buf[offset] != Encoding32bitInt {
		return 0, offset, fmt.Errorf("not a 32-bit encoding at offset %d", offset)
	}

	val := binary.LittleEndian.Uint32(buf[offset+1:])
	result := int(int32(val))
	offset += 6

	return result, offset, nil
}

// encode64BitInt encodes a 64-bit signed integer with backLen
// Format: 0xF4 + 8 bytes (little-endian) + backLen
// Entry structure: encoding(1) + data(8) + backLen
func encode64BitInt(buf []byte, offset int, n int) int {
	buf[offset] = Encoding64bitInt
	binary.LittleEndian.PutUint64(buf[offset+1:], uint64(n))
	offset += 9

	// Calculate entry length (1 + 8 = 9 bytes)
	entryLen := uint64(9)

	// Encode backLen
	backLenSize := lpEncodeBackLen(buf, offset, entryLen)
	offset += backLenSize

	return offset
}

// decode64BitInt decodes a 64-bit signed integer with backLen
func decode64BitInt(buf []byte, offset int) (int, int, error) {
	if offset+8 >= len(buf) {
		return 0, offset, fmt.Errorf("buffer overflow at offset %d", offset)
	}

	if buf[offset] != Encoding64bitInt {
		return 0, offset, fmt.Errorf("not a 64-bit encoding at offset %d", offset)
	}

	val := binary.LittleEndian.Uint64(buf[offset+1:])
	result := int(int64(val))
	offset += 10

	return result, offset, nil
}
