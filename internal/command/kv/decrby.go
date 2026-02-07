package kv

import (
	"avacado/internal/command"
	"avacado/internal/protocol"
	"avacado/internal/storage"
	"context"
	"strconv"
)

// DecrBy represent a decrby command containing key and decrement as arguments
// decrements the number stored at key by the specified decrement.
// If the key does not exist, it is set to 0 before performing the operation.
// An error is returned if the value stored at key is not a string representing an integer.
type DecrBy struct {
	Key       string
	Decrement int64
}

func (d *DecrBy) Execute(ctx context.Context, storage storage.Storage) *protocol.Response {
	v, err := storage.KV().DecrBy(ctx, d.Key, d.Decrement)
	if err != nil {
		return protocol.NewErrorResponse(err)
	}
	return protocol.NewNumberResponse(v)
}

type DecrByParser struct {
}

func NewDecrByParser() *DecrByParser {
	return &DecrByParser{}
}

func (d *DecrByParser) Parse(msg *protocol.Message) (command.Command, error) {
	key, err := msg.Args[0].AsString()
	if err != nil {
		return nil, err
	}

	decrementStr, err := msg.Args[1].AsString()
	if err != nil {
		return nil, err
	}

	decrement, err := strconv.ParseInt(decrementStr, 10, 64)
	if err != nil {
		return nil, err
	}

	return &DecrBy{Key: key, Decrement: decrement}, nil
}

func (d *DecrByParser) Name() string {
	return "DECRBY"
}
