package hashmap

import (
	"avacado/internal/command"
	"avacado/internal/protocol"
	"avacado/internal/storage"
	"context"
)

type HSet struct {
	name      string
	keyValues []string
}

func (h *HSet) Execute(ctx context.Context, storage storage.Storage) *protocol.Response {
	n := storage.Maps().HSet(ctx, h.name, h.keyValues)
	return protocol.NewNumberResponse(int64(n))
}

type HSetParser struct {
}

func (p *HSetParser) Parse(msg *protocol.Message) (command.Command, error) {
	if len(msg.Args) < 3 {
		return nil, command.NewInvalidArgumentsCount(p.Name(), 3, len(msg.Args))
	}
	// name + even count of (field + value)
	if len(msg.Args)%2 != 1 {
		return nil, command.NewInvalidArgumentsCount(p.Name(), len(msg.Args)+1, len(msg.Args))
	}

	name, err := msg.Args[0].AsString()
	if err != nil {
		return nil, command.NewInvalidTypeError(name, "name")
	}

	keyValues := make([]string, len(msg.Args)-1)
	for i := 0; i < len(msg.Args)-1; i++ {
		entry, err := msg.Args[i+1].AsString()
		if err != nil {
			return nil, command.NewInvalidTypeError(p.Name(), "field-value")
		}
		keyValues[i] = entry
	}

	return &HSet{name: name, keyValues: keyValues}, nil
}

func (p *HSetParser) Name() string {
	return "HSET"
}

func NewHSetParser() *HSetParser {
	return &HSetParser{}
}
