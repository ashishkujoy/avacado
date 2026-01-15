package kv

import (
	"avacado/internal/command"
	"avacado/internal/protocol"
	"avacado/internal/storage"
	"context"
	"fmt"
)

// Set represent a set command containing key and value as arguments
type Set struct {
	Key   string
	Value []byte
}

func (s *Set) Execute(ctx context.Context, storage storage.Storage) *protocol.Response {
	err := storage.KV().Set(ctx, s.Key, s.Value)
	if err != nil {
		return protocol.NewNullBulkString()
	}
	return protocol.NewSimpleStringResponse("OK")
}

// SetParser parses the set command
type SetParser struct {
}

func NewSetParser() SetParser {
	return SetParser{}
}

func (s *SetParser) Parse(msg *protocol.Message) (command.Command, error) {
	key, err := msg.Args[0].AsString()
	if err != nil {
		return nil, fmt.Errorf("set command failed to parse key: %w", err)
	}
	value, err := msg.Args[1].AsBytes()
	if err != nil {
		return nil, fmt.Errorf("set command failed to parse value: %w", err)
	}
	return &Set{Key: key, Value: value}, nil
}

func (s *SetParser) Name() string {
	return "SET"
}
