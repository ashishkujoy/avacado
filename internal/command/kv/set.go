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
	Key   string
	Value []byte
	NX    bool
}

func (s *Set) Execute(ctx context.Context, storage storage.Storage) *protocol.Response {
	err := storage.KV().Set(ctx, s.Key, s.Value, &kv.SetOptions{NX: s.NX})
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
	if len(msg.Args) > 2 {
		if nx, err := msg.Args[2].AsString(); err == nil {
			cmd.NX = strings.ToUpper(nx) == "NX"
		}
	}
	return cmd, nil
}

func (s SetParser) Name() string {
	return "SET"
}
