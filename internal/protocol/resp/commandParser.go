package resp

import (
	"avacado/internal/protocol"
	"fmt"
	"io"
)

// CommandParser implements a parser for RESP commands
type CommandParser struct {
	respParser *Parser
}

// toProtocolValue converts resp value to protocol value
func toProtocolValue(value Value) (protocol.Value, error) {
	switch {
	case value.IsString():
		{
			str, err := value.AsString()
			if err != nil {
				return protocol.Value{}, err
			}
			return protocol.NewStringProtocolValue(str), nil
		}
	case value.IsNumber():
		{
			num, err := value.AsNumber()
			if err != nil {
				return protocol.Value{}, err
			}
			return protocol.NewNumberProtocolValue(num), nil
		}
	case value.IsNumber():
		{
			num, err := value.AsNumber()
			if err != nil {
				return protocol.Value{}, err
			}
			return protocol.NewNumberProtocolValue(num), nil
		}
	case value.IsBulk():
		{
			bulk, err := value.AsBulk()
			if err != nil {
				return protocol.Value{}, err
			}
			return protocol.NewBulkStringProtocolValue(bulk), nil
		}
	case value.IsArray():
		{
			array, err := value.AsArray()
			if err != nil {
				return protocol.Value{}, err
			}
			values := make([]protocol.Value, len(array))
			for i := 0; i < len(array); i++ {
				values[i], err = toProtocolValue(array[i])
				if err != nil {
					return protocol.Value{}, err
				}
			}
			return protocol.NewArrayProtocolValue(values), nil
		}
	case value.IsMap():
		{
			entries, err := value.AsMap()
			if err != nil {
				return protocol.Value{}, err
			}
			protocolEntries := make([]protocol.MapEntry, len(entries))
			for i, entry := range entries {
				protocolVal, err := toProtocolValue(entry.Val)
				if err != nil {
					return protocol.Value{}, err
				}
				protocolEntries[i] = protocol.MapEntry{
					Key: entry.Key,
					Val: protocolVal,
				}
			}
			return protocol.NewMapProtocolValue(protocolEntries), nil
		}
	}
	return protocol.Value{}, fmt.Errorf("unreachable")
}

// Parse parses a RESP command from the given io reader
func (c *CommandParser) Parse() (*protocol.Message, error) {
	parser := c.respParser
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
	args := make([]protocol.Value, len(array)-1)
	for i := 1; i < len(array); i++ {
		args[i-1], err = toProtocolValue(array[i])
		if err != nil {
			return nil, fmt.Errorf("failed to parse command argument: %w", err)
		}
	}
	return &protocol.Message{Command: command, Args: args}, nil
}

func NewCommandParser(reader io.Reader) *CommandParser {
	return &CommandParser{
		respParser: NewParser(reader),
	}
}
