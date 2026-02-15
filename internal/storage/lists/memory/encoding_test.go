package memory

import (
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
