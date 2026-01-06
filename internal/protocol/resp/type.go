package resp

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
