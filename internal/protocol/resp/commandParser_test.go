package resp

import (
	"avacado/internal/protocol"
	"bufio"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToProtocolValue_Map(t *testing.T) {
	respValue := NewMap([]MapEntry{
		{Key: "name", Val: NewBulkString([]byte("John"))},
		{Key: "age", Val: NewInteger(30)},
	})
	protocolValue, err := toProtocolValue(respValue)

	assert.NoError(t, err)
	assert.Equal(t, byte('%'), byte(protocolValue.Type))
	assert.Equal(t, 2, len(protocolValue.Map))

	assert.Equal(t, "name", protocolValue.Map[0].Key)
	assert.Equal(t, []byte("John"), protocolValue.Map[0].Val.Bytes)

	assert.Equal(t, "age", protocolValue.Map[1].Key)
	assert.Equal(t, int64(30), protocolValue.Map[1].Val.Number)
}

func TestToProtocolValue_EmptyMap(t *testing.T) {
	respValue := NewMap([]MapEntry{})
	protocolValue, err := toProtocolValue(respValue)

	assert.NoError(t, err)
	assert.Equal(t, byte('%'), byte(protocolValue.Type))
	assert.Equal(t, 0, len(protocolValue.Map))
}

func TestToProtocolValue_NestedMap(t *testing.T) {
	innerMap := NewMap([]MapEntry{
		{Key: "inner", Val: NewInteger(42)},
	})
	respValue := NewMap([]MapEntry{
		{Key: "outer", Val: innerMap},
	})
	protocolValue, err := toProtocolValue(respValue)

	assert.NoError(t, err)
	assert.Equal(t, 1, len(protocolValue.Map))
	assert.Equal(t, "outer", protocolValue.Map[0].Key)

	innerProtocolMap := protocolValue.Map[0].Val
	assert.Equal(t, byte('%'), byte(innerProtocolMap.Type))
	assert.Equal(t, 1, len(innerProtocolMap.Map))
	assert.Equal(t, "inner", innerProtocolMap.Map[0].Key)
	assert.Equal(t, int64(42), innerProtocolMap.Map[0].Val.Number)
}

func TestToProtocolValue_MapPreservesOrder(t *testing.T) {
	respValue := NewMap([]MapEntry{
		{Key: "first", Val: NewInteger(1)},
		{Key: "second", Val: NewInteger(2)},
		{Key: "third", Val: NewInteger(3)},
		{Key: "fourth", Val: NewInteger(4)},
	})
	protocolValue, err := toProtocolValue(respValue)

	assert.NoError(t, err)
	assert.Equal(t, "first", protocolValue.Map[0].Key)
	assert.Equal(t, "second", protocolValue.Map[1].Key)
	assert.Equal(t, "third", protocolValue.Map[2].Key)
	assert.Equal(t, "fourth", protocolValue.Map[3].Key)
}

func TestToProtocolValue_MapInArray(t *testing.T) {
	mapValue := NewMap([]MapEntry{
		{Key: "key", Val: NewBulkString([]byte("value"))},
	})
	respValue := NewArray([]Value{
		NewInteger(1),
		mapValue,
	})
	protocolValue, err := toProtocolValue(respValue)

	assert.NoError(t, err)
	assert.Equal(t, byte('*'), byte(protocolValue.Type))
	assert.Equal(t, 2, len(protocolValue.Array))
	assert.Equal(t, byte('%'), byte(protocolValue.Array[1].Type))
	assert.Equal(t, 1, len(protocolValue.Array[1].Map))
}

func TestCommandParser_ParseRoundTripWithMap(t *testing.T) {
	// Create a RESP map, serialize it, parse it back, and convert to protocol value
	serializer := NewRESPSerializer()
	entries := []protocol.MapEntry{
		{Key: "name", Val: protocol.NewBulkStringProtocolValue([]byte("John"))},
		{Key: "age", Val: protocol.NewNumberProtocolValue(30)},
	}
	resp := protocol.NewMapResponse(entries)
	bytes, err := serializer.Serialize(resp)
	assert.NoError(t, err)

	// Parse the serialized RESP back
	parser := NewParser(bufio.NewReader(strings.NewReader(string(bytes))))
	respValue, err := parser.Parse()
	assert.NoError(t, err)

	// Convert to protocol value
	protocolValue, err := toProtocolValue(respValue)
	assert.NoError(t, err)

	assert.Equal(t, byte('%'), byte(protocolValue.Type))
	assert.Equal(t, 2, len(protocolValue.Map))
	assert.Equal(t, "name", protocolValue.Map[0].Key)
	assert.Equal(t, "age", protocolValue.Map[1].Key)
}
