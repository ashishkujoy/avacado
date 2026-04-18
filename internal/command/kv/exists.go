package kv

import (
	"avacado/internal/command"
	"avacado/internal/protocol"
	"avacado/internal/storage"
	"context"
)

// Exists represent an exists command containing one or more keys as arguments
// Returns the number of keys that exist. Keys mentioned multiple times are
// counted multiple times. Expired keys are treated as non-existent.
type Exists struct {
	Keys []string
}

func (e *Exists) Execute(ctx context.Context, storage storage.Storage) *protocol.Response {
	count, err := storage.KV().Exists(ctx, e.Keys...)
	if err != nil {
		return protocol.NewErrorResponse(err)
	}
	return protocol.NewNumberResponse(count)
}

type ExistsParser struct {
}

func NewExistsParser() *ExistsParser {
	return &ExistsParser{}
}

func (e *ExistsParser) Parse(msg *protocol.Message) (command.Command, error) {
	return &Exists{Keys: msg.Args}, nil
}

func (e *ExistsParser) Name() string {
	return "EXISTS"
}
