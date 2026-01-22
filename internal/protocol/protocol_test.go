package protocol

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValue_AsInt64_FromNumber(t *testing.T) {
	value := NewNumberProtocolValue(42)
	result, err := value.AsInt64()
	assert.NoError(t, err)
	assert.Equal(t, int64(42), result)
}

func TestValue_AsInt64_FromSimpleString(t *testing.T) {
	value := NewStringProtocolValue("123")
	result, err := value.AsInt64()
	assert.NoError(t, err)
	assert.Equal(t, int64(123), result)
}

func TestValue_AsInt64_FromBulkString(t *testing.T) {
	value := NewBulkStringProtocolValue([]byte("456"))
	result, err := value.AsInt64()
	assert.NoError(t, err)
	assert.Equal(t, int64(456), result)
}

func TestValue_AsInt64_FromNegativeNumber(t *testing.T) {
	value := NewStringProtocolValue("-789")
	result, err := value.AsInt64()
	assert.NoError(t, err)
	assert.Equal(t, int64(-789), result)
}

func TestValue_AsInt64_InvalidString(t *testing.T) {
	value := NewStringProtocolValue("not-a-number")
	_, err := value.AsInt64()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot be parsed as int64")
}

func TestValue_AsInt64_FromArray(t *testing.T) {
	value := NewArrayProtocolValue([]Value{
		NewStringProtocolValue("test"),
	})
	_, err := value.AsInt64()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not a number or string")
}
