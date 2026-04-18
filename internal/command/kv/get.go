package kv

import (
	"avacado/internal/command"
	"avacado/internal/protocol"
	"avacado/internal/storage"
	"context"
)

type Get struct {
	key string
}

func (g *Get) Execute(ctx context.Context, storage storage.Storage) *protocol.Response {
	data, err := storage.KV().Get(ctx, g.key)
	if err != nil || data == nil {
		return protocol.NewNullBulkStringResponse()
	}
	return protocol.NewBulkStringResponse(data)
}

type GetParser struct {
}

func NewGetParser() GetParser {
	return GetParser{}
}

func (s GetParser) Parse(msg *protocol.Message) (command.Command, error) {
	return &Get{key: msg.Args[0]}, nil
}

func (s GetParser) Name() string {
	return "GET"
}
