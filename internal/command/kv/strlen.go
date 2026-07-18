package kv

import (
	"avacado/internal/command"
	"avacado/internal/protocol"
	"avacado/internal/storage"
	"context"
)

type Strlen struct {
	Key string
}

func (s *Strlen) Execute(ctx context.Context, storage storage.Storage) *protocol.Response {
	length, err := storage.KV().Len(ctx, s.Key)
	if err != nil {
		return protocol.NewErrorResponse(err)
	}
	return protocol.NewNumberResponse(length)
}

type StrlenParser struct{}

func NewStrlenParser() *StrlenParser {
	return &StrlenParser{}
}

func (p *StrlenParser) Parse(msg *protocol.Message) (command.Command, error) {
	return &Strlen{Key: msg.Args[0]}, nil
}

func (p *StrlenParser) Name() string {
	return "STRLEN"
}
