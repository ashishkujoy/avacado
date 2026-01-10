package resp

import "fmt"

type Type = byte

const (
	TypeSimpleString Type = '+'
	TypeError        Type = '-'
	TypeInteger      Type = ':'
	TypeBulkString   Type = '$'
	TypeArray        Type = '*'
)

// Value represent a RESP value
type Value struct {
	Type  Type
	Str   string
	Num   int64
	Bulk  []byte
	Array []Value
	Null  bool
}

func (v *Value) IsString() bool {
	return v.Type == TypeSimpleString
}

func (v *Value) AsString() (string, error) {
	if v.Type != TypeSimpleString {
		return "", fmt.Errorf("value is not a string, got type: %c", v.Type)
	}
	if v.Null {
		return "", fmt.Errorf("null string")
	}
	return v.Str, nil
}

func (v *Value) IsNumber() bool {
	return v.Type == TypeInteger
}

func (v *Value) AsNumber() (int64, error) {
	if v.Type != TypeInteger {
		return 0, fmt.Errorf("value is not a number, got type: %c", v.Type)
	}
	if v.Null {
		return 0, fmt.Errorf("null number")
	}
	return v.Num, nil
}

func (v *Value) IsBulk() bool {
	return v.Type == TypeBulkString
}

// AsBulk returns the value as a byte array
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

// NewSimpleString creates a simple string value
func NewSimpleString(s string) Value {
	return Value{Type: TypeSimpleString, Str: s}
}

// NewError creates an error value
func NewError(s string) Value {
	return Value{Type: TypeError, Str: s}
}

// NewInteger creates an integer value
func NewInteger(n int64) Value {
	return Value{Type: TypeInteger, Num: n}
}

// NewBulkString creates a bulk string value
func NewBulkString(b []byte) Value {
	return Value{Type: TypeBulkString, Bulk: b}
}

// NewBulkStringFromString creates a bulk string value from a string
func NewBulkStringFromString(s string) Value {
	return Value{Type: TypeBulkString, Bulk: []byte(s)}
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
