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
		err = traverse(buf[:offset7], 0, func(el interface{}) error {
			got = append(got, el)
			return nil
		})
		assert.NoError(t, err)
		expected := []interface{}{42, 190, -32768, 8388608, []byte("Hi"), str12Bit, str32Bit}
		assert.Equal(t, expected, got)
	})

	t.Run("empty buffer", func(t *testing.T) {
		var got []interface{}
		err := traverse([]byte{}, 0, func(el interface{}) error {
			got = append(got, el)
			return nil
		})
		assert.NoError(t, err)
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
		err = traverse(buf[:offset2], 0, func(el interface{}) error {
			got = append(got, el)
			return sentinel
		})
		assert.ErrorIs(t, err, sentinel)
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
		err = traverseReverse(buf[:offset7], offset7-1, func(el interface{}) error {
			got = append(got, el)
			return nil
		})
		assert.NoError(t, err)
		expected := []interface{}{str32Bit, str12Bit, []byte("Hi"), 8388608, -32768, 190, 42}
		assert.Equal(t, expected, got)
	})

	t.Run("empty buffer", func(t *testing.T) {
		var got []interface{}
		err := traverseReverse([]byte{}, -1, func(el interface{}) error {
			got = append(got, el)
			return nil
		})
		assert.NoError(t, err)
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
		err = traverseReverse(buf[:offset3], offset3-1, func(el interface{}) error {
			got = append(got, el)
			return sentinel
		})
		assert.ErrorIs(t, err, sentinel)
		assert.Len(t, got, 1)
	})
}

func newStringOfLength(n int) []byte {
	buf := make([]byte, n)
	for i := range buf {
		buf[i] = 'x'
	}
	return buf
}
