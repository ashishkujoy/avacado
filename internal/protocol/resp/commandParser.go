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
	args := make([]string, len(array)-1)
	for i := 1; i < len(array); i++ {
		bulk, err := array[i].AsBulk()
		if err != nil {
			return nil, fmt.Errorf("command argument is not a bulk string: %w", err)
		}
		args[i-1] = string(bulk)
	}
	return &protocol.Message{Command: command, Args: args}, nil
}

func NewCommandParser(reader io.Reader) *CommandParser {
	return &CommandParser{
		respParser: NewParser(reader),
	}
}
