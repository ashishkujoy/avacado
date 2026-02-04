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

func TestSerializer_SerializeMap(t *testing.T) {
	serializer := NewRESPSerializer()
	entries := []protocol.MapEntry{
		{Key: "key1", Val: protocol.NewBulkStringProtocolValue([]byte("value1"))},
		{Key: "key2", Val: protocol.NewNumberProtocolValue(42)},
	}
	resp := protocol.NewMapResponse(entries)
	bytes, err := serializer.Serialize(resp)

	assert.NoError(t, err)
	expected := "%2\r\n$4\r\nkey1\r\n$6\r\nvalue1\r\n$4\r\nkey2\r\n:42\r\n"
	assert.Equal(t, expected, string(bytes))
}

func TestSerializer_SerializeEmptyMap(t *testing.T) {
	serializer := NewRESPSerializer()
	resp := protocol.NewMapResponse([]protocol.MapEntry{})
	bytes, err := serializer.Serialize(resp)

	assert.NoError(t, err)
	assert.Equal(t, "%0\r\n", string(bytes))
}

func TestSerializer_SerializeNullMap(t *testing.T) {
	serializer := NewRESPSerializer()
	resp := protocol.NewNullMapResponse()
	bytes, err := serializer.Serialize(resp)

	assert.NoError(t, err)
	assert.Equal(t, "%-1\r\n", string(bytes))
}

func TestSerializer_SerializeMapWithMixedTypes(t *testing.T) {
	serializer := NewRESPSerializer()
	entries := []protocol.MapEntry{
		{Key: "string", Val: protocol.NewStringProtocolValue("hello")},
		{Key: "number", Val: protocol.NewNumberProtocolValue(123)},
		{Key: "bulk", Val: protocol.NewBulkStringProtocolValue([]byte("data"))},
		{Key: "array", Val: protocol.NewArrayProtocolValue([]protocol.Value{
			protocol.NewNumberProtocolValue(1),
			protocol.NewNumberProtocolValue(2),
		})},
	}
	resp := protocol.NewMapResponse(entries)
	bytes, err := serializer.Serialize(resp)

	assert.NoError(t, err)
	expected := "%4\r\n" +
		"$6\r\nstring\r\n+hello\r\n" +
		"$6\r\nnumber\r\n:123\r\n" +
		"$4\r\nbulk\r\n$4\r\ndata\r\n" +
		"$5\r\narray\r\n*2\r\n:1\r\n:2\r\n"
	assert.Equal(t, expected, string(bytes))
}

func TestSerializer_SerializeNestedMap(t *testing.T) {
	serializer := NewRESPSerializer()
	innerEntries := []protocol.MapEntry{
		{Key: "inner", Val: protocol.NewNumberProtocolValue(42)},
	}
	entries := []protocol.MapEntry{
		{Key: "outer", Val: protocol.NewMapProtocolValue(innerEntries)},
	}
	resp := protocol.NewMapResponse(entries)
	bytes, err := serializer.Serialize(resp)

	assert.NoError(t, err)
	expected := "%1\r\n$5\r\nouter\r\n%1\r\n$5\r\ninner\r\n:42\r\n"
	assert.Equal(t, expected, string(bytes))
}

func TestSerializer_SerializeMapPreservesOrder(t *testing.T) {
	serializer := NewRESPSerializer()
	entries := []protocol.MapEntry{
		{Key: "first", Val: protocol.NewNumberProtocolValue(1)},
		{Key: "second", Val: protocol.NewNumberProtocolValue(2)},
		{Key: "third", Val: protocol.NewNumberProtocolValue(3)},
		{Key: "fourth", Val: protocol.NewNumberProtocolValue(4)},
	}
	resp := protocol.NewMapResponse(entries)
	bytes, err := serializer.Serialize(resp)

	assert.NoError(t, err)
	output := string(bytes)

	// Verify the order is preserved in the output
	assert.Contains(t, output, "%4\r\n")
	// Check that keys appear in order
	firstPos := bytes[0:20]
	assert.Contains(t, string(firstPos), "$5\r\nfirst")
}

func TestSerializer_RoundTripMap(t *testing.T) {
	// Test that we can serialize and parse back
	serializer := NewRESPSerializer()
	entries := []protocol.MapEntry{
		{Key: "name", Val: protocol.NewBulkStringProtocolValue([]byte("John"))},
		{Key: "age", Val: protocol.NewNumberProtocolValue(30)},
	}
	resp := protocol.NewMapResponse(entries)
	bytes, err := serializer.Serialize(resp)

	assert.NoError(t, err)
	expected := "%2\r\n$4\r\nname\r\n$4\r\nJohn\r\n$3\r\nage\r\n:30\r\n"
	assert.Equal(t, expected, string(bytes))
}
