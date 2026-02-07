package kv

import (
	"avacado/internal/command"
	"avacado/internal/protocol"
	"avacado/internal/storage"
	"context"
)

// Decr represent a decr command containing key as argument
// decrements the number stored at key by one.
// If the key does not exist, it is set to 0 before performing the operation.
// An error is returned if the value stored at key is not a string representing an integer.
type Decr struct {
	Key string
}

func (d *Decr) Execute(ctx context.Context, storage storage.Storage) *protocol.Response {
	v, err := storage.KV().Decr(ctx, d.Key)
	if err != nil {
		return protocol.NewErrorResponse(err)
	}
	return protocol.NewNumberResponse(v)
}

type DecrParser struct {
}

func NewDecrParser() *DecrParser {
	return &DecrParser{}
}

func (d *DecrParser) Parse(msg *protocol.Message) (command.Command, error) {
	key, err := msg.Args[0].AsString()
	if err != nil {
		return nil, err
	}
	return &Decr{Key: key}, nil
}

func (d *DecrParser) Name() string {
	return "DECR"
}
