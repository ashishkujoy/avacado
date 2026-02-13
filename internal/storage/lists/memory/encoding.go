package memory

import (
	"errors"
	"strconv"
)

// encode7BitInt encodes an integer value into a 7-bit number.
// Format: 0xxxxxxx for values 0-127
// The first bit is 0 representing that it's a 7-bit number, and the remaining 7 bits represent the value.
// Back length: 2 bytes (1 byte for the value and 1 byte for the back length)
func encode7BitInt(buf []byte, offset, value int) int {
	buf[offset] = byte(value)
	buf[offset+1] = byte(2)
	return offset + 2
}

// decode7BitInt decodes a 7-bit integer from the buffer starting at the given offset.
func decode7BitInt(buf []byte, offset int) (int, int) {
	value := int(buf[offset])
	return value, offset + 2
}

// encode13BitInt encodes an integer value into a 13-bit number.
// Format: 110xxxxx xxxxxxxx for values -4096 to 4095
// The first 3 bits are 110 representing that it's a 13-bit number, and the remaining 13 bits represent the value.
// Back length: 3 bytes (2 bytes for the value and 1 byte for the back length)
func encode13BitInt(buf []byte, offset, value int) int {
	var uval uint16
	if value < 0 {
		uval = uint16(value + 8192)
	} else {
		uval = uint16(value)
	}
	top5 := byte((uval >> 8) & 0x1F)
	bottom8 := byte(uval & 0xFF)

	// Byte 1: 110xxxxx (0xC0 is 11000000, OR with top 5 bits)
	buf[offset] = 0xC0 | top5
	// Byte 2: xxxxxxxx (bottom 8 bits)
	buf[offset+1] = bottom8
	// Byte 3: backlen (total entry length = 3 bytes)
	buf[offset+2] = 3
	return offset + 3
}

// decode13BitInt decodes a 13-bit integer from the buffer starting at the given offset.
func decode13BitInt(buf []byte, offset int) (int, int) {
	// Extract the 5 bits from first byte (mask out the 110 prefix)
	high5 := uint16(buf[offset] & 0x1F)
	// Get the 8 bits from second byte
	low8 := uint16(buf[offset+1])
	// Combine to get 13-bit unsigned value
	uval := (high5 << 8) | low8
	// Convert back to signed
	if uval >= 4096 {
		// Negative number
		return int(uval) - 8192, offset + 3
	}
	return int(uval), offset + 3
}

var unknownEncodingError = errors.New("unknown encoding")
var unknownTypeError = errors.New("unknown type")

// decode decodes an element from the buffer starting at the given offset.
func decode(buf []byte, offset int) (interface{}, int, error) {
	prefix := buf[offset]
	if prefix&0x80 == 0 {
		// 7-bit number
		v, o := decode7BitInt(buf, offset)
		return v, o, nil
	} else if prefix&0xE0 == 0xC0 {
		// 13-bit number
		v, o := decode13BitInt(buf, offset)
		return v, o, nil
	}
	return nil, offset, unknownEncodingError
}

// encode determines the type of the element and encodes it into the buffer starting at the given offset.
func encode(buf []byte, offset int, element []byte) (int, error) {
	if v, err := strconv.Atoi(string(element)); err == nil {
		if v >= 0 && v <= 127 {
			return encode7BitInt(buf, offset, v), nil
		} else if v >= -4096 && v <= 4095 {
			return encode13BitInt(buf, offset, v), nil
		}
		return offset, unknownTypeError
	}
	return offset, unknownTypeError
}
