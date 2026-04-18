package resp

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
)

// Parser parses incoming RESP commands from clients.
// Per the Redis protocol spec, clients only send arrays of bulk strings.
type Parser struct {
	reader *bufio.Reader
}

// Parse reads and parses a single RESP value from the reader.
// Only Array and BulkString types are accepted, matching the Redis client protocol.
func (p *Parser) Parse() (Value, error) {
	typeByte, err := p.reader.ReadByte()
	if err != nil {
		return Value{}, err
	}
	switch typeByte {
	case TypeBulkString:
		return p.parseBulkString()
	case TypeArray:
		return p.parseArray()
	default:
		return Value{}, fmt.Errorf("unsupported RESP type from client: %c (only arrays of bulk strings are allowed)", typeByte)
	}
}

// readLine reads a line until \r\n (excluding the \r\n)
func (p *Parser) readLine() ([]byte, error) {
	bytes, err := p.reader.ReadBytes('\n')
	if err != nil {
		return nil, err
	}
	if len(bytes) < 2 || bytes[len(bytes)-2] != '\r' {
		return nil, fmt.Errorf("line does not end with CRLF")
	}
	return bytes[:len(bytes)-2], nil
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

	if length == -1 {
		return NewNullBulkString(), nil
	}

	if length < 0 {
		return Value{}, fmt.Errorf("invalid bulk string length: %d", length)
	}

	bulk := make([]byte, length)
	_, err = io.ReadFull(p.reader, bulk)
	if err != nil {
		return Value{}, fmt.Errorf("failed to read bulk string data: %w", err)
	}

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

	if length == -1 {
		return NewNullArray(), nil
	}

	if length < 0 {
		return Value{}, fmt.Errorf("invalid array length: %d", length)
	}

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
