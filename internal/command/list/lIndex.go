package list

import (
	"avacado/internal/command"
	"avacado/internal/protocol"
	"avacado/internal/storage"
	"context"
	"errors"
)

type LIndex struct {
	Key   string
	Index int
}

func (l *LIndex) Execute(ctx context.Context, storage storage.Storage) *protocol.Response {
	_, err := storage.KV().Get(ctx, l.Key)
	// Element represented by key is not list
	if err == nil {
		return protocol.NewErrorResponse(errors.New("(error) WRONGTYPE Operation against a key holding the wrong kind of value"))
	}
	element, err := storage.Lists().LIndex(ctx, l.Key, l.Index)
	if err != nil {
		return protocol.NewNullBulkStringResponse()
	}
	return protocol.NewBulkStringResponse(element)
}

type LIndexParser struct {
}

func NewLIndexParser() *LIndexParser {
	return &LIndexParser{}
}

func (l *LIndexParser) Parse(msg *protocol.Message) (command.Command, error) {
	if len(msg.Args) != 2 {
		return nil, command.NewInvalidArgumentsCount(l.Name(), 2, len(msg.Args))
	}
	key, err := msg.Args[0].AsString()
	if err != nil {
		return nil, command.NewInvalidTypeError(l.Name(), "KEY")
	}
	index, err := msg.Args[1].AsInt64()
	if err != nil {
		return nil, command.NewInvalidTypeError(l.Name(), "INDEX")
	}
	return &LIndex{
		Key:   key,
		Index: int(index),
	}, nil
}

func (l *LIndexParser) Name() string {
	return "LINDEX"
}
