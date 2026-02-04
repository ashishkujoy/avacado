package resp

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
)

// Parser parses the RESP protocol message
type Parser struct {
	reader *bufio.Reader
}

// Parse reads and parses a single RESP value from the reader
func (p *Parser) Parse() (Value, error) {
	typeByte, err := p.reader.ReadByte()
	if err != nil {
		return Value{}, err
	}
	switch typeByte {
	case TypeSimpleString:
		return p.parseSimpleString()
	case TypeError:
		return p.parseError()
	case TypeInteger:
		return p.parseInteger()
	case TypeBulkString:
		return p.parseBulkString()
	case TypeArray:
		return p.parseArray()
	case TypeMap:
		return p.parseMap()
	default:
		return Value{}, fmt.Errorf("unknown RESP type: %c", typeByte)
	}
}

// parseSimpleString parse a simple string
func (p *Parser) parseSimpleString() (Value, error) {
	line, err := p.readLine()
	if err != nil {
		return Value{}, fmt.Errorf("failed to read simple string: %w", err)
	}
	return NewSimpleString(string(line)), err
}

// readLine reads a line until \r\n (excluding the \r\n)
func (p *Parser) readLine() ([]byte, error) {
	bytes, err := p.reader.ReadBytes('\n')
	if err != nil {
		return nil, err
	}

	// Check for \r\n ending
	if len(bytes) < 2 || bytes[len(bytes)-2] != '\r' {
		return nil, fmt.Errorf("line does not end with CRLF")
	}

	return bytes[:len(bytes)-2], nil
}

// parseError parse an error
func (p *Parser) parseError() (Value, error) {
	line, err := p.readLine()
	if err != nil {
		return Value{}, fmt.Errorf("failed to read error: %w", err)
	}
	return NewError(string(line)), err
}

// parseInteger parses a RESP integer (:...\r\n)
func (p *Parser) parseInteger() (Value, error) {
	line, err := p.readLine()
	if err != nil {
		return Value{}, fmt.Errorf("failed to read integer: %w", err)
	}

	num, err := strconv.ParseInt(string(line), 10, 64)
	if err != nil {
		return Value{}, fmt.Errorf("invalid integer: %w", err)
	}

	return NewInteger(num), nil
}

// parseBulkString parses a RESP bulk string ($...\r\n...\r\n)
func (p *Parser) parseBulkString() (Value, error) {
	line, err := p.readLine()
	if err != nil {
		return Value{}, fmt.Errorf("failed to read bulk string length: %w", err)
	}

	length, err := strconv.Atoi(string(line))
	if err != nil {
		return Value{}, fmt.Errorf("invalid bulk string length: %w", err)
	}

	// Handle null bulk string
	if length == -1 {
		return NewNullBulkString(), nil
	}

	if length < 0 {
		return Value{}, fmt.Errorf("invalid bulk string length: %d", length)
	}

	// Read the bulk data
	bulk := make([]byte, length)
	_, err = io.ReadFull(p.reader, bulk)
	if err != nil {
		return Value{}, fmt.Errorf("failed to read bulk string data: %w", err)
	}

	// Read the trailing \r\n
	if err := p.expectCRLF(); err != nil {
		return Value{}, fmt.Errorf("bulk string missing CRLF: %w", err)
	}

	return NewBulkString(bulk), nil
}

// expectCRLF reads and validates a CRLF sequence
func (p *Parser) expectCRLF() error {
	cr, err := p.reader.ReadByte()
	if err != nil {
		return err
	}
	lf, err := p.reader.ReadByte()
	if err != nil {
		return err
	}

	if cr != '\r' || lf != '\n' {
		return fmt.Errorf("expected CRLF, got: %c%c", cr, lf)
	}

	return nil
}

// parseMap parses a RESP map
func (p *Parser) parseMap() (Value, error) {
	line, err := p.readLine()
	if err != nil {
		return Value{}, fmt.Errorf("failed to read map size: %w", err)
	}

	size, err := strconv.Atoi(string(line))
	if err != nil {
		return Value{}, fmt.Errorf("invalid map size: %w", err)
	}

	entries := make(map[string]Value)
	for i := 0; i < size; i++ {
		key, err := p.Parse()
		if err != nil {
			return Value{}, fmt.Errorf("failed to parse map key at index %d", i)
		}
		keyName, err := key.AsString()
		if err != nil {
			return Value{}, fmt.Errorf("unexpected key type at index %d", i)
		}
		value, err := p.Parse()
		if err != nil {
			return Value{}, fmt.Errorf("failed to parse value for key %s", keyName)
		}
		entries[keyName] = value
	}
	return NewMap(entries), nil
}

// parseArray parses a RESP array (*...\r\n...)
func (p *Parser) parseArray() (Value, error) {
	line, err := p.readLine()
	if err != nil {
		return Value{}, fmt.Errorf("failed to read array length: %w", err)
	}

	length, err := strconv.Atoi(string(line))
	if err != nil {
		return Value{}, fmt.Errorf("invalid array length: %w", err)
	}

	// Handle null array
	if length == -1 {
		return NewNullArray(), nil
	}

	if length < 0 {
		return Value{}, fmt.Errorf("invalid array length: %d", length)
	}

	// Parse array elements
	array := make([]Value, length)
	for i := 0; i < length; i++ {
		value, err := p.Parse()
		if err != nil {
			return Value{}, fmt.Errorf("failed to parse array element %d: %w", i, err)
		}
		array[i] = value
	}

	return NewArray(array), nil
}

// NewParser creates a new buffered io based parser from the given io reader
func NewParser(r io.Reader) *Parser {
	return &Parser{reader: bufio.NewReader(r)}
}
