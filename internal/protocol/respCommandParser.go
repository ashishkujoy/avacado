package protocol

import (
	"avacado/internal/protocol/resp"
	"fmt"
	"io"
)

// RESPCommandParser implements a parser for RESP commands
type RESPCommandParser struct {
}

// toProtocolValue converts resp value to protocol value
func toProtocolValue(value resp.Value) (Value, error) {
	switch {
	case value.IsString():
		{
			str, err := value.AsString()
			if err != nil {
				return Value{}, err
			}
			return NewStringProtocolValue(str), nil
		}
	case value.IsNumber():
		{
			num, err := value.AsNumber()
			if err != nil {
				return Value{}, err
			}
			return NewNumberProtocolValue(num), nil
		}
	case value.IsNumber():
		{
			num, err := value.AsNumber()
			if err != nil {
				return Value{}, err
			}
			return NewNumberProtocolValue(num), nil
		}
	case value.IsBulk():
		{
			bulk, err := value.AsBulk()
			if err != nil {
				return Value{}, err
			}
			return NewBytesProtocolValue(bulk), nil
		}
	case value.IsArray():
		{
			array, err := value.AsArray()
			if err != nil {
				return Value{}, err
			}
			values := make([]Value, len(array))
			for i := 0; i < len(array); i++ {
				values[i], err = toProtocolValue(array[i])
				if err != nil {
					return Value{}, err
				}
			}
			return NewArrayProtocolValue(values), nil
		}
	}
	return Value{}, fmt.Errorf("unreachable")
}

// Parse parses a RESP command from the given io reader
func (c *RESPCommandParser) Parse(r io.Reader) (*Message, error) {
	parser := resp.NewParser(r)
	value, err := parser.Parse()
	if err != nil {
		return nil, fmt.Errorf("failed to parse command: %w", err)
	}
	array, err := value.AsArray()
	if err != nil {
		return nil, fmt.Errorf("command is not an array: %w", err)
	}
	if len(array) == 0 {
		return nil, fmt.Errorf("command is empty")
	}

	cmdBytes, err := array[0].AsBulk()
	if err != nil {
		return nil, fmt.Errorf("command name is not a bulk string: %w", err)
	}
	command := string(cmdBytes)
	args := make([]Value, len(array)-1)
	for i := 1; i < len(array); i++ {
		args[i-1], err = toProtocolValue(array[i])
		if err != nil {
			return nil, fmt.Errorf("failed to parse command argument: %w", err)
		}
	}
	return &Message{Command: command, Args: args}, nil
}
