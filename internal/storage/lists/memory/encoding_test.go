package memory

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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
