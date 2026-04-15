package hashmap

import (
	"avacado/internal/command"
	"avacado/internal/protocol"
	"avacado/internal/storage"
	"context"
)

type hGet struct {
	name  string
	field string
}

func (h *hGet) Execute(ctx context.Context, storage storage.Storage) *protocol.Response {
	value, err := storage.Maps().HGet(ctx, h.name, h.field)
	if err != nil {
		return protocol.NewNullBulkStringResponse()
	}
	return protocol.NewBulkStringResponse(value)
}

type HGetParser struct {
}

func (h *HGetParser) Parse(msg *protocol.Message) (command.Command, error) {
	if len(msg.Args) != 2 {
		return nil, command.NewInvalidArgumentsCount(h.Name(), 2, len(msg.Args))
	}
	name, err := msg.Args[0].AsString()
	if err != nil {
		return nil, command.NewInvalidTypeError(h.Name(), "name")
	}
	field, err := msg.Args[1].AsString()
	if err != nil {
		return nil, command.NewInvalidTypeError(h.Name(), "field")
	}
	return &hGet{
		name:  name,
		field: field,
	}, nil
}

func (h *HGetParser) Name() string {
	return "HGET"
}

func NewHGetParser() *HGetParser {
	return &HGetParser{}
}
