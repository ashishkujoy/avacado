package protocol

import (
	"fmt"
	"io"
)

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

func (v *Value) AsString() (string, error) {
	if v.Type == TypeSimpleString {
		return v.Str, nil
	}
	if v.Type == TypeBulkString {
		return string(v.Bytes), nil
	}
	return "", fmt.Errorf("value is not a string")
}

func (v *Value) AsBytes() ([]byte, error) {
	if v.Type == TypeSimpleString {
		return []byte(v.Str), nil
	}
	if v.Type == TypeBulkString {
		return v.Bytes, nil
	}
	return nil, fmt.Errorf("value is not bytes")
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

// NewErrorResponse creates a new error response
func NewErrorResponse(err error) *Response {
	return &Response{Err: err}
}

// Serializer serializes the protocol message
type Serializer interface {
	Serialize(value *Response) ([]byte, error)
	SerializeError(e error) []byte
}

//go:generate sh -c "rm -f mock/protocol.go && mockgen -source=protocol.go -destination=mock/protocol.go -package=mockprotocol"
type Protocol interface {
	Serializer
	Parser
}

func NewSimpleStringResponse(s string) *Response {
	return NewSuccessResponse(NewStringProtocolValue(s))
}

func NewNullBulkStringResponse() *Response {
	return NewSuccessResponse(NewNullBulkStringProtocolValue())
}

func NewBulkStringResponse(b []byte) *Response {
	return NewSuccessResponse(NewBulkStringProtocolValue(b))
}

func NewNullBulkStringProtocolValue() Value {
	return Value{Null: true, Type: TypeBulkString}
}
