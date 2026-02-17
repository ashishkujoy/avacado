package memory

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEncoding_16BitNumber(t *testing.T) {
	buf := make([]byte, 100)
	offset1, err := encode(buf, 0, []byte("4096"))
	assert.NoError(t, err)

	offset2, err := encode(buf, offset1, []byte("32767"))
	assert.NoError(t, err)

	offset3, err := encode(buf, offset2, []byte("-32768"))
	assert.NoError(t, err)

	value1, newOffset1, err := decode(buf, 0)
	assert.NoError(t, err)
	assert.Equal(t, 4096, value1)
	assert.Equal(t, offset1, newOffset1)

	value2, newOffset2, err := decode(buf, offset1)
	assert.NoError(t, err)
	assert.Equal(t, 32767, value2)
	assert.Equal(t, offset2, newOffset2)

	value3, newOffset3, err := decode(buf, offset2)
	assert.NoError(t, err)
	assert.Equal(t, -32768, value3)
	assert.Equal(t, offset3, newOffset3)
}

func TestEncoding_13BitNumber(t *testing.T) {
	buf := make([]byte, 100)
	offset1, err := encode(buf, 0, []byte("128"))
	assert.NoError(t, err)

	offset2, err := encode(buf, offset1, []byte("4095"))
	assert.NoError(t, err)

	offset3, err := encode(buf, offset2, []byte("-4096"))
	assert.NoError(t, err)

	value1, newOffset1, err := decode(buf, 0)
	assert.NoError(t, err)
	assert.Equal(t, 128, value1)
	assert.Equal(t, offset1, newOffset1)

	value2, newOffset2, err := decode(buf, offset1)
	assert.NoError(t, err)
	assert.Equal(t, 4095, value2)
	assert.Equal(t, offset2, newOffset2)

	value3, newOffset3, err := decode(buf, offset2)
	assert.NoError(t, err)
	assert.Equal(t, -4096, value3)
	assert.Equal(t, offset3, newOffset3)
}

func TestEncoding_7BitNumber(t *testing.T) {
	buf := make([]byte, 100)
	offset1, err := encode(buf, 0, []byte("127"))
	assert.NoError(t, err)

	offset2, err := encode(buf, offset1, []byte("500"))
	assert.NoError(t, err)

	offset3, err := encode(buf, offset2, []byte("-17"))
	assert.NoError(t, err)

	value1, newOffset1, err := decode(buf, 0)
	assert.NoError(t, err)
	assert.Equal(t, 127, value1)
	assert.Equal(t, offset1, newOffset1)

	value2, newOffset2, err := decode(buf, offset1)
	assert.NoError(t, err)
	assert.Equal(t, 500, value2)
	assert.Equal(t, offset2, newOffset2)

	value3, newOffset3, err := decode(buf, offset2)
	assert.NoError(t, err)
	assert.Equal(t, -17, value3)
	assert.Equal(t, offset3, newOffset3)
}

func TestEncoding_24BitNumber(t *testing.T) {
	buf := make([]byte, 100)
	offset1, err := encode(buf, 0, []byte("32768"))
	assert.NoError(t, err)

	offset2, err := encode(buf, offset1, []byte("8388607"))
	assert.NoError(t, err)

	offset3, err := encode(buf, offset2, []byte("-8388608"))
	assert.NoError(t, err)

	value1, newOffset1, err := decode(buf, 0)
	assert.NoError(t, err)
	assert.Equal(t, 32768, value1)
	assert.Equal(t, offset1, newOffset1)

	value2, newOffset2, err := decode(buf, offset1)
	assert.NoError(t, err)
	assert.Equal(t, 8388607, value2)
	assert.Equal(t, offset2, newOffset2)

	value3, newOffset3, err := decode(buf, offset2)
	assert.NoError(t, err)
	assert.Equal(t, -8388608, value3)
	assert.Equal(t, offset3, newOffset3)
}

func TestEncoding_32BitNumber(t *testing.T) {
	buf := make([]byte, 100)
	offset1, err := encode(buf, 0, []byte("8388608"))
	assert.NoError(t, err)

	offset2, err := encode(buf, offset1, []byte("16777216"))
	assert.NoError(t, err)

	offset3, err := encode(buf, offset2, []byte("-16777216"))
	assert.NoError(t, err)

	value1, newOffset1, err := decode(buf, 0)
	assert.NoError(t, err)
	assert.Equal(t, 8388608, value1)
	assert.Equal(t, offset1, newOffset1)

	value2, newOffset2, err := decode(buf, offset1)
	assert.NoError(t, err)
	assert.Equal(t, 16777216, value2)
	assert.Equal(t, offset2, newOffset2)

	value3, newOffset3, err := decode(buf, offset2)
	assert.NoError(t, err)
	assert.Equal(t, -16777216, value3)
	assert.Equal(t, offset3, newOffset3)
}

func TestEncoding_64BitNumber(t *testing.T) {
	buf := make([]byte, 100)
	offset1, err := encode(buf, 0, []byte("16777217"))
	assert.NoError(t, err)

	offset2, err := encode(buf, offset1, []byte("2147483647"))
	assert.NoError(t, err)

	offset3, err := encode(buf, offset2, []byte("-2147483648"))
	assert.NoError(t, err)

	value1, newOffset1, err := decode(buf, 0)
	assert.NoError(t, err)
	assert.Equal(t, 16777217, value1)
	assert.Equal(t, offset1, newOffset1)

	value2, newOffset2, err := decode(buf, offset1)
	assert.NoError(t, err)
	assert.Equal(t, 2147483647, value2)
	assert.Equal(t, offset2, newOffset2)

	value3, newOffset3, err := decode(buf, offset2)
	assert.NoError(t, err)
	assert.Equal(t, -2147483648, value3)
	assert.Equal(t, offset3, newOffset3)
}

func TestEncoding_6BitString(t *testing.T) {
	buf := make([]byte, 256)

	// short string, empty string, exactly 63 bytes
	offset1, err := encode(buf, 0, []byte("hello"))
	assert.NoError(t, err)

	offset2, err := encode(buf, offset1, []byte(""))
	assert.NoError(t, err)

	pad63 := []byte("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa") // 63 'a's
	offset3, err := encode(buf, offset2, pad63)
	assert.NoError(t, err)

	v1, newOffset1, err := decode(buf, 0)
	assert.NoError(t, err)
	assert.Equal(t, []byte("hello"), v1)
	assert.Equal(t, offset1, newOffset1)

	v2, newOffset2, err := decode(buf, offset1)
	assert.NoError(t, err)
	assert.Equal(t, []byte(""), v2)
	assert.Equal(t, offset2, newOffset2)

	v3, newOffset3, err := decode(buf, offset2)
	assert.NoError(t, err)
	assert.Equal(t, pad63, v3)
	assert.Equal(t, offset3, newOffset3)
}

func TestEncoding_12BitString(t *testing.T) {
	buf := make([]byte, 8192)

	// 64 bytes (just over 6-bit limit), 4095 bytes (12-bit max)
	s64 := make([]byte, 64)
	for i := range s64 {
		s64[i] = 'x'
	}
	s4095 := make([]byte, 4095)
	for i := range s4095 {
		s4095[i] = 'y'
	}

	offset1, err := encode(buf, 0, s64)
	assert.NoError(t, err)

	offset2, err := encode(buf, offset1, s4095)
	assert.NoError(t, err)

	v1, newOffset1, err := decode(buf, 0)
	assert.NoError(t, err)
	assert.Equal(t, s64, v1)
	assert.Equal(t, offset1, newOffset1)

	v2, newOffset2, err := decode(buf, offset1)
	assert.NoError(t, err)
	assert.Equal(t, s4095, v2)
	assert.Equal(t, offset2, newOffset2)
}

func TestEncoding_32BitString(t *testing.T) {
	buf := make([]byte, 8192)

	// 4096 bytes (just over 12-bit limit)
	s4096 := make([]byte, 4096)
	for i := range s4096 {
		s4096[i] = 'z'
	}

	offset1, err := encode(buf, 0, s4096)
	assert.NoError(t, err)

	v1, newOffset1, err := decode(buf, 0)
	assert.NoError(t, err)
	assert.Equal(t, s4096, v1)
	assert.Equal(t, offset1, newOffset1)
}

func TestTraverse(t *testing.T) {
	t.Run("Traverse mixed values", func(t *testing.T) {
		buf := make([]byte, 256000)
		offset1, err := encode(buf, 0, []byte("42"))
		assert.NoError(t, err)
		offset2, err := encode(buf, offset1, []byte("190"))
		assert.NoError(t, err)
		offset3, err := encode(buf, offset2, []byte("-32768"))
		assert.NoError(t, err)
		offset4, err := encode(buf, offset3, []byte("8388608"))
		assert.NoError(t, err)
		offset5, err := encode(buf, offset4, []byte("Hi"))
		assert.NoError(t, err)
		str12Bit := newStringOfLength(64)
		offset6, err := encode(buf, offset5, str12Bit)
		assert.NoError(t, err)
		str32Bit := newStringOfLength(4096)
		offset7, err := encode(buf, offset6, str32Bit)
		assert.NoError(t, err)

		var got []interface{}
		endOffset, err := traverse(buf[:offset7], 0, func(el interface{}) (bool, error) {
			got = append(got, el)
			return true, nil
		})
		assert.NoError(t, err)
		assert.Equal(t, offset7, endOffset)
		expected := []interface{}{42, 190, -32768, 8388608, []byte("Hi"), str12Bit, str32Bit}
		assert.Equal(t, expected, got)
	})

	t.Run("empty buffer", func(t *testing.T) {
		var got []interface{}
		endOffset, err := traverse([]byte{}, 0, func(el interface{}) (bool, error) {
			got = append(got, el)
			return true, nil
		})
		assert.NoError(t, err)
		assert.Equal(t, 0, endOffset)
		assert.Empty(t, got)
	})

	t.Run("callback error stops traversal", func(t *testing.T) {
		buf := make([]byte, 256)
		offset1, err := encode(buf, 0, []byte("1"))
		assert.NoError(t, err)
		offset2, err := encode(buf, offset1, []byte("2"))
		assert.NoError(t, err)
		_, err = encode(buf, offset2, []byte("3"))
		assert.NoError(t, err)

		var got []interface{}
		sentinel := errors.New("stop")
		_, err = traverse(buf[:offset2], 0, func(el interface{}) (bool, error) {
			got = append(got, el)
			return true, sentinel
		})
		assert.ErrorIs(t, err, sentinel)
		assert.Len(t, got, 1)
	})

	t.Run("callback false stops traversal", func(t *testing.T) {
		buf := make([]byte, 256)
		offset1, err := encode(buf, 0, []byte("1"))
		assert.NoError(t, err)
		offset2, err := encode(buf, offset1, []byte("2"))
		assert.NoError(t, err)
		_, err = encode(buf, offset2, []byte("3"))
		assert.NoError(t, err)

		var got []interface{}
		endOffset, err := traverse(buf[:offset2], 0, func(el interface{}) (bool, error) {
			got = append(got, el)
			return false, nil
		})
		assert.NoError(t, err)
		assert.Equal(t, offset1, endOffset)
		assert.Len(t, got, 1)
	})
}

func TestTraverseReverse(t *testing.T) {
	t.Run("Traverse reverse mixed values", func(t *testing.T) {
		buf := make([]byte, 256000)
		offset1, err := encode(buf, 0, []byte("42"))
		assert.NoError(t, err)
		offset2, err := encode(buf, offset1, []byte("190"))
		assert.NoError(t, err)
		offset3, err := encode(buf, offset2, []byte("-32768"))
		assert.NoError(t, err)
		offset4, err := encode(buf, offset3, []byte("8388608"))
		assert.NoError(t, err)
		offset5, err := encode(buf, offset4, []byte("Hi"))
		assert.NoError(t, err)
		str12Bit := newStringOfLength(64)
		offset6, err := encode(buf, offset5, str12Bit)
		assert.NoError(t, err)
		str32Bit := newStringOfLength(4096)
		offset7, err := encode(buf, offset6, str32Bit)
		assert.NoError(t, err)

		var got []interface{}
		endOffset, err := traverseReverse(buf[:offset7], offset7-1, func(el interface{}) (bool, error) {
			got = append(got, el)
			return true, nil
		})
		assert.NoError(t, err)
		assert.Equal(t, -1, endOffset)
		expected := []interface{}{str32Bit, str12Bit, []byte("Hi"), 8388608, -32768, 190, 42}
		assert.Equal(t, expected, got)
	})

	t.Run("empty buffer", func(t *testing.T) {
		var got []interface{}
		endOffset, err := traverseReverse([]byte{}, -1, func(el interface{}) (bool, error) {
			got = append(got, el)
			return true, nil
		})
		assert.NoError(t, err)
		assert.Equal(t, -1, endOffset)
		assert.Empty(t, got)
	})

	t.Run("callback error stops traversal", func(t *testing.T) {
		buf := make([]byte, 256)
		offset1, err := encode(buf, 0, []byte("1"))
		assert.NoError(t, err)
		offset2, err := encode(buf, offset1, []byte("2"))
		assert.NoError(t, err)
		offset3, err := encode(buf, offset2, []byte("3"))
		assert.NoError(t, err)

		var got []interface{}
		sentinel := errors.New("stop")
		_, err = traverseReverse(buf[:offset3], offset3-1, func(el interface{}) (bool, error) {
			got = append(got, el)
			return true, sentinel
		})
		assert.ErrorIs(t, err, sentinel)
		assert.Len(t, got, 1)
	})

	t.Run("callback false stops traversal", func(t *testing.T) {
		buf := make([]byte, 256)
		offset1, err := encode(buf, 0, []byte("1"))
		assert.NoError(t, err)
		offset2, err := encode(buf, offset1, []byte("2"))
		assert.NoError(t, err)
		offset3, err := encode(buf, offset2, []byte("3"))
		assert.NoError(t, err)

		var got []interface{}
		endOffset, err := traverseReverse(buf[:offset3], offset3-1, func(el interface{}) (bool, error) {
			got = append(got, el)
			return false, nil
		})
		assert.NoError(t, err)
		assert.Equal(t, offset2-1, endOffset)
		assert.Len(t, got, 1)
	})
}

func TestEncode_Overflow(t *testing.T) {
	t.Run("integer too large for buffer", func(t *testing.T) {
		buf := make([]byte, 1) // "42" needs 2 bytes (7-bit int + backlen)
		offset, err := encode(buf, 0, []byte("42"))
		assert.Error(t, err)
		assert.Equal(t, 0, offset) // offset unchanged on error
	})

	t.Run("string too large for buffer", func(t *testing.T) {
		buf := make([]byte, 6) // "hello" needs 7 bytes (1 header + 5 data + 1 backlen)
		offset, err := encode(buf, 0, []byte("hello"))
		assert.Error(t, err)
		assert.Equal(t, 0, offset)
	})

	t.Run("integer exact fit", func(t *testing.T) {
		buf := make([]byte, 2) // exactly enough for "42"
		offset, err := encode(buf, 0, []byte("42"))
		assert.NoError(t, err)
		assert.Equal(t, 2, offset)
	})

	t.Run("string exact fit", func(t *testing.T) {
		buf := make([]byte, 7) // exactly enough for "hello"
		offset, err := encode(buf, 0, []byte("hello"))
		assert.NoError(t, err)
		assert.Equal(t, 7, offset)
	})

	t.Run("second encode overflows after first succeeds", func(t *testing.T) {
		// "42" (2 bytes) fits at offset 0; "hello" (7 bytes) at offset 2 needs 9 total — only 8 available
		buf := make([]byte, 8)
		offset, err := encode(buf, 0, []byte("42"))
		assert.NoError(t, err)
		_, err = encode(buf, offset, []byte("hello"))
		assert.Error(t, err)
		// buffer must not be corrupted: first entry still decodable
		v, _, err := decode(buf, 0)
		assert.NoError(t, err)
		assert.Equal(t, 42, v)
	})

	t.Run("zero-length buffer", func(t *testing.T) {
		buf := make([]byte, 0)
		_, err := encode(buf, 0, []byte("1"))
		assert.Error(t, err)
	})
}

func TestEncodedSize(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected int
	}{
		{"7-bit int", []byte("0"), 2},
		{"7-bit int max", []byte("127"), 2},
		{"13-bit int", []byte("128"), 3},
		{"13-bit int max", []byte("4095"), 3},
		{"16-bit int", []byte("4096"), 4},
		{"16-bit int max", []byte("32767"), 4},
		{"24-bit int", []byte("32768"), 5},
		{"24-bit int max", []byte("8388607"), 5},
		{"32-bit int", []byte("8388608"), 6},
		{"64-bit int", []byte("2147483648"), 10},
		{"6-bit string empty", []byte(""), 2},   // 1 header + 0 data + 1 backlen
		{"6-bit string hello", []byte("hello"), 7}, // 1 + 5 + 1
		{"6-bit string max (63 bytes)", newStringOfLength(63), 65}, // 1 + 63 + 1
		{"12-bit string (64 bytes)", newStringOfLength(64), 67},    // 2 + 64 + 1
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, encodedSize(tt.input))
		})
	}
}

// decodeAt is a helper that resolves a backLen cursor to the decoded entry value.
// cursor must point to the last byte of an entry's backLen field.
func decodeAt(buf []byte, cursor int) (interface{}, error) {
	entryLen, backLenSize, err := lpDecodeBackLen(buf, cursor)
	if err != nil {
		return nil, err
	}
	entryStart := cursor - backLenSize - int(entryLen) + 1
	value, _, err := decode(buf, entryStart)
	return value, err
}


func newStringOfLength(n int) []byte {
	buf := make([]byte, n)
	for i := range buf {
		buf[i] = 'x'
	}
	return buf
}
