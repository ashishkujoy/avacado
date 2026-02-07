package kv

import (
	"avacado/internal/command"
	"avacado/internal/protocol"
	"avacado/internal/storage"
	"context"
)

// Incr represent an incr command containing key as argument
// increments the number stored at key by one.
// If the key does not exist, it is set to 0 before performing the operation.
// An error is returned if the value stored at key is not a string representing an integer.
type Incr struct {
	Key string
}

func (i *Incr) Execute(ctx context.Context, storage storage.Storage) *protocol.Response {
	v, err := storage.KV().Incr(ctx, i.Key)
	if err != nil {
		return protocol.NewErrorResponse(err)
	}
	return protocol.NewNumberResponse(v)
}

type IncrParser struct {
}

func NewIncrParser() *IncrParser {
	return &IncrParser{}
}

func (i *IncrParser) Parse(msg *protocol.Message) (command.Command, error) {
	key, err := msg.Args[0].AsString()
	if err != nil {
		return nil, err
	}
	return &Incr{Key: key}, nil
}

func (i *IncrParser) Name() string {
	return "INCR"
}
