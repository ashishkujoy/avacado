package hashmap

import (
	"avacado/internal/command"
	"avacado/internal/protocol"
	"avacado/internal/storage"
	"context"
)

type hMGet struct {
	key    string
	fields []string
}

func (h *hMGet) Execute(ctx context.Context, storage storage.Storage) *protocol.Response {
	values := storage.Maps().HMGet(ctx, h.key, h.fields)
	return protocol.NewArrayResponse(values)
}

type HMGetParser struct{}

func (p *HMGetParser) Parse(msg *protocol.Message) (command.Command, error) {
	if len(msg.Args) < 2 {
		return nil, command.NewInvalidArgumentsCount(p.Name(), 2, len(msg.Args))
	}
	return &hMGet{key: msg.Args[0], fields: msg.Args[1:]}, nil
}

func (p *HMGetParser) Name() string {
	return "HMGET"
}

func NewHMGetParser() *HMGetParser {
	return &HMGetParser{}
}
