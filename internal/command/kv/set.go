package kv

import (
	"avacado/internal/command"
	"avacado/internal/protocol"
	"avacado/internal/storage"
	"avacado/internal/storage/kv"
	"context"
	"fmt"
	"strings"
)

// Set represent a set command containing key and value as arguments
type Set struct {
	Key     string
	Value   []byte
	Options *kv.SetOptions
}

func (s *Set) Execute(ctx context.Context, storage storage.Storage) *protocol.Response {
	err := storage.KV().Set(ctx, s.Key, s.Value, s.Options)
	if err != nil {
		return protocol.NewNullBulkStringResponse()
	}
	return protocol.NewSimpleStringResponse("OK")
}

// SetParser parses the set command
type SetParser struct {
}

func NewSetParser() SetParser {
	return SetParser{}
}

func (s SetParser) Parse(msg *protocol.Message) (command.Command, error) {
	key, err := msg.Args[0].AsString()
	if err != nil {
		return nil, fmt.Errorf("set command failed to parse key: %w", err)
	}
	value, err := msg.Args[1].AsBytes()
	if err != nil {
		return nil, fmt.Errorf("set command failed to parse value: %w", err)
	}
	cmd := &Set{Key: key, Value: value}
	options := kv.NewSetOptions()
	for i := 2; i < len(msg.Args); i++ {
		argName, _ := msg.Args[i].AsString()
		argName = strings.ToUpper(argName)
		if argName == "NX" {
			options = options.WithNX()
		}
		if argName == "XX" {
			options = options.WithXX()
		}
		if argName == "EX" {
			// EX expects a numeric value in the next argument
			if i+1 >= len(msg.Args) {
				return nil, fmt.Errorf("set command: EX option requires a value")
			}
			i++ // Move to the next argument
			exSeconds, err := msg.Args[i].AsInt64()
			if err != nil {
				return nil, fmt.Errorf("set command: EX value must be an integer: %w", err)
			}
			options = options.WithEX(exSeconds)
		}
		// TODO: error handling for unknown arg
	}
	cmd.Options = options
	return cmd, nil
}

func (s SetParser) Name() string {
	return "SET"
}
