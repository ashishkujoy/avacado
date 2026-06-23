package kv

import (
	"avacado/internal/command"
	"avacado/internal/protocol"
	"avacado/internal/storage"
	"context"
)

type Append struct {
	Key   string
	Value []byte
}

func (a *Append) Execute(ctx context.Context, storage storage.Storage) *protocol.Response {
	length, err := storage.KV().Append(ctx, a.Key, a.Value)
	if err != nil {
		return protocol.NewErrorResponse(err)
	}
	return protocol.NewNumberResponse(length)
}

type AppendParser struct{}

func NewAppendParser() *AppendParser {
	return &AppendParser{}
}

func (p *AppendParser) Parse(msg *protocol.Message) (command.Command, error) {
	return &Append{Key: msg.Args[0], Value: []byte(msg.Args[1])}, nil
}

func (p *AppendParser) Name() string {
	return "APPEND"
}
