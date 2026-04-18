package resp

import "fmt"

type Type = byte

const (
	TypeSimpleString Type = '+'
	TypeError        Type = '-'
	TypeInteger      Type = ':'
	TypeBulkString   Type = '$'
	TypeArray        Type = '*'
	TypeMap          Type = '%'
)

// Value represents a parsed RESP value (Array or BulkString only for client commands)
type Value struct {
	Type  Type
	Bulk  []byte
	Array []Value
	Null  bool
}

func (v *Value) IsBulk() bool {
	return v.Type == TypeBulkString
}

// AsBulk returns the value as a byte slice
func (v *Value) AsBulk() ([]byte, error) {
	if v.Type != TypeBulkString {
		return nil, fmt.Errorf("value is not a bulk string, got type: %c", v.Type)
	}
	if v.Null {
		return nil, fmt.Errorf("null bulk string")
	}
	return v.Bulk, nil
}

func (v *Value) IsArray() bool {
	return v.Type == TypeArray
}

// AsArray returns the value as an array
func (v *Value) AsArray() ([]Value, error) {
	if v.Type != TypeArray {
		return nil, fmt.Errorf("value is not an array, got type: %c", v.Type)
	}
	if v.Null {
		return nil, fmt.Errorf("null array")
	}
	return v.Array, nil
}

// NewBulkString creates a bulk string value
func NewBulkString(b []byte) Value {
	return Value{Type: TypeBulkString, Bulk: b}
}

// NewNullBulkString creates a null bulk string
func NewNullBulkString() Value {
	return Value{Type: TypeBulkString, Null: true}
}

// NewArray creates an array value
func NewArray(values []Value) Value {
	return Value{Type: TypeArray, Array: values}
}

// NewNullArray creates a null array
func NewNullArray() Value {
	return Value{Type: TypeArray, Null: true}
}
