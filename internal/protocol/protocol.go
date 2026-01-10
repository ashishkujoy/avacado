package protocol

import "io"

// Parser parses the protocol message
type Parser interface {
	Parse(r io.Reader) (*Message, error)
}

type ValueType string

const (
	String ValueType = "string"
	Bytes            = "bytes"
	Number           = "number"
	Array            = "array"
	Null             = "null"
)

// Value represent a protocol value
type Value struct {
	Type   ValueType
	Str    string
	Bytes  []byte
	Number int64
	Array  []Value
	Null   bool
}

func NewStringProtocolValue(s string) Value {
	return Value{Type: String, Str: s}
}

func NewBytesProtocolValue(b []byte) Value {
	return Value{Type: Bytes, Bytes: b}
}

func NewNumberProtocolValue(n int64) Value {
	return Value{Type: Number, Number: n}
}

func NewArrayProtocolValue(values []Value) Value {
	return Value{Type: Array, Array: values}
}

// Message represents a protocol message, containing command name and args
type Message struct {
	Command string
	Args    []Value
}

// Serializer serializes the protocol message
type Serializer interface {
	Serialize(value interface{}) ([]byte, error)
	SerializeError(e error) []byte
}

type Protocol interface {
	Serializer
	Parser
}
