package hashmap

import (
	"avacado/internal/command"
	"avacado/internal/protocol"
	"avacado/internal/storage"
	"context"
)

type HExists struct {
	key   string
	field string
}

func (h *HExists) Execute(ctx context.Context, storage storage.Storage) *protocol.Response {
	result := storage.Maps().HExists(ctx, h.key, h.field)
	return protocol.NewNumberResponse(int64(result))
}

type HExistsParser struct{}

func NewHExistsParser() *HExistsParser {
	return &HExistsParser{}
}

func (p *HExistsParser) Parse(msg *protocol.Message) (command.Command, error) {
	if len(msg.Args) < 2 {
		return nil, command.NewInvalidArgumentsCount(p.Name(), 2, len(msg.Args))
	}
	return &HExists{key: msg.Args[0], field: msg.Args[1]}, nil
}

func (p *HExistsParser) Name() string {
	return "HEXISTS"
}
