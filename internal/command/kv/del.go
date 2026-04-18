package kv

import (
	"avacado/internal/command"
	"avacado/internal/protocol"
	"avacado/internal/storage"
	"context"
)

// Del represent a del command containing one or more keys as arguments
// removes the specified keys. A key is ignored if it does not exist.
// Returns the number of keys that were removed.
type Del struct {
	Keys []string
}

func (d *Del) Execute(ctx context.Context, storage storage.Storage) *protocol.Response {
	count, err := storage.KV().Del(ctx, d.Keys...)
	if err != nil {
		return protocol.NewErrorResponse(err)
	}
	return protocol.NewNumberResponse(count)
}

type DelParser struct {
}

func NewDelParser() *DelParser {
	return &DelParser{}
}

func (d *DelParser) Parse(msg *protocol.Message) (command.Command, error) {
	return &Del{Keys: msg.Args}, nil
}

func (d *DelParser) Name() string {
	return "DEL"
}
