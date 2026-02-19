package list

import (
	"avacado/internal/command"
	"avacado/internal/protocol"
	"avacado/internal/storage"
	"context"
)

type RPush struct {
	Key    string
	Values [][]byte
}

func (r *RPush) Execute(ctx context.Context, storage storage.Storage) *protocol.Response {
	size, err := storage.Lists().RPush(ctx, r.Key, r.Values...)
	if err != nil {
		return protocol.NewErrorResponse(err)
	}
	return protocol.NewNumberResponse(int64(size))
}

type RPushParser struct {
}

func NewRPushParser() *RPushParser {
	return &RPushParser{}
}

func (r *RPushParser) Parse(msg *protocol.Message) (command.Command, error) {
	if len(msg.Args) < 2 {
		return nil, command.NewInvalidArgumentsCount("RPUSH", 2, len(msg.Args))
	}
	key, err := msg.Args[0].AsString()
	if err != nil {
		return nil, command.NewInvalidTypeError("RPUSH", "key")
	}
	args := msg.Args[1:]
	values := make([][]byte, len(args))
	for i, arg := range args {
		value, e := arg.AsBytes()
		if e != nil {
			return nil, command.NewInvalidTypeError("RPUSH", "values")
		}
		values[i] = value
	}
	return &RPush{Key: key, Values: values}, nil
}

func (r *RPushParser) Name() string {
	return "RPUSH"
}
