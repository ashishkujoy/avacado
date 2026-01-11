package resp

import (
	"avacado/internal/protocol"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSerializer_SerializeError(t *testing.T) {
	serializer := NewRESPSerializer()
	bytes := serializer.SerializeError(fmt.Errorf("invalid command"))

	assert.Equal(t, string(bytes), "-invalid command\r\n")
}

func TestSerializer_SerializeSimpleString(t *testing.T) {
	serializer := NewRESPSerializer()
	resp := protocol.NewSuccessResponse(protocol.NewStringProtocolValue("OK"))
	bytes, err := serializer.Serialize(resp)

	assert.NoError(t, err)
	assert.Equal(t, string(bytes), "+OK\r\n")
}

func TestSerializer_SerializeBulkString(t *testing.T) {
	serializer := NewRESPSerializer()
	resp := protocol.NewSuccessResponse(protocol.NewBulkStringProtocolValue([]byte("OK")))
	bytes, err := serializer.Serialize(resp)
	assert.NoError(t, err)
	assert.Equal(t, "$2\r\nOK\r\n", string(bytes))
}

func TestSerializer_SerializeNumber(t *testing.T) {
	serializer := NewRESPSerializer()
	resp := protocol.NewSuccessResponse(protocol.NewNumberProtocolValue(123))
	bytes, err := serializer.Serialize(resp)
	assert.NoError(t, err)
	assert.Equal(t, ":123\r\n", string(bytes))
}

func TestSerializer_SerializeArray(t *testing.T) {
	serializer := NewRESPSerializer()
	resp := protocol.NewSuccessResponse(protocol.NewArrayProtocolValue([]protocol.Value{
		protocol.NewStringProtocolValue("OK"),
		protocol.NewNumberProtocolValue(123),
		protocol.NewBulkStringProtocolValue([]byte("OK")),
		protocol.NewArrayProtocolValue(
			[]protocol.Value{
				protocol.NewStringProtocolValue("Hello"),
				protocol.NewNumberProtocolValue(456),
			},
		),
	}))
	bytes, err := serializer.Serialize(resp)
	assert.NoError(t, err)
	assert.Equal(t, "*4\r\n+OK\r\n:123\r\n$2\r\nOK\r\n*2\r\n+Hello\r\n:456\r\n", string(bytes))
}
