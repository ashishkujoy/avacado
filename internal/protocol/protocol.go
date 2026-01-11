package protocol

import "io"

// Parser parses the protocol message
type Parser interface {
	Parse(r io.Reader) (*Message, error)
}

type ValueType byte

const (
	TypeSimpleString ValueType = '+'
	TypeBulkString             = '$'
	TypeNumber                 = ':'
	TypeArray                  = '*'
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
	return Value{Type: TypeSimpleString, Str: s}
}

func NewBulkStringProtocolValue(b []byte) Value {
	return Value{Type: TypeBulkString, Bytes: b}
}

func NewNumberProtocolValue(n int64) Value {
	return Value{Type: TypeNumber, Number: n}
}

func NewArrayProtocolValue(values []Value) Value {
	return Value{Type: TypeArray, Array: values}
}

// Message represents a protocol message, containing command name and args
type Message struct {
	Command string
	Args    []Value
}

// Response represents a protocol response
type Response struct {
	Value Value
	Err   error
}

// NewSuccessResponse creates a new success response
func NewSuccessResponse(value Value) *Response {
	return &Response{Value: value}
}

// Serializer serializes the protocol message
type Serializer interface {
	Serialize(value *Response) ([]byte, error)
	SerializeError(e error) []byte
}

type Protocol interface {
	Serializer
	Parser
}
