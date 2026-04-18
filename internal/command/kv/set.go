package kv

import (
	"avacado/internal/command"
	"avacado/internal/protocol"
	"avacado/internal/storage"
	"avacado/internal/storage/kv"
	"context"
	"fmt"
	"strconv"
	"strings"
)

// Set represent a set command containing key and value as arguments
type Set struct {
	Key     string
	Value   []byte
	Options *kv.SetOptions
}

func (s *Set) Execute(ctx context.Context, storage storage.Storage) *protocol.Response {
	oldValue, err := storage.KV().Set(ctx, s.Key, s.Value, s.Options)
	if err != nil {
		return protocol.NewNullBulkStringResponse()
	}
	if s.Options.Get {
		if oldValue == nil {
			return protocol.NewNullBulkStringResponse()
		}
		return protocol.NewBulkStringResponse(oldValue)
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
	cmd := &Set{Key: msg.Args[0], Value: []byte(msg.Args[1])}
	options := kv.NewSetOptions()
	for i := 2; i < len(msg.Args); i++ {
		argName := strings.ToUpper(msg.Args[i])
		if argName == "NX" {
			options = options.WithNX()
		}
		if argName == "XX" {
			options = options.WithXX()
		}
		if argName == "EX" {
			if i+1 >= len(msg.Args) {
				return nil, fmt.Errorf("set command: EX option requires a value")
			}
			i++
			exSeconds, err := strconv.ParseInt(msg.Args[i], 10, 64)
			if err != nil {
				return nil, fmt.Errorf("set command: EX value must be an integer: %w", err)
			}
			options = options.WithEX(exSeconds)
		}
		if argName == "GET" {
			options = options.WithGet()
		}
		if argName == "IFEQ" {
			if i+1 >= len(msg.Args) {
				return nil, fmt.Errorf("set command: IFEQ option requires a value")
			}
			i++
			options = options.WithIFEQ([]byte(msg.Args[i]))
		}
	}
	cmd.Options = options
	return cmd, nil
}

func (s SetParser) Name() string {
	return "SET"
}
