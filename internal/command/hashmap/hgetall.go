package hashmap

import (
	"avacado/internal/command"
	"avacado/internal/config"
	"avacado/internal/protocol"
	"avacado/internal/storage"
	"context"
)

type hGetAll struct {
	key string
}

func (h *hGetAll) Execute(ctx context.Context, storage storage.Storage) *protocol.Response {
	entries, err := storage.Maps().HGetAll(ctx, h.key)
	if err != nil {
		return protocol.NewErrorResponse(err)
	}
	if config.IsProto3(ctx) {
		values := make([]protocol.MapEntry, 0, len(entries))
		for key, value := range entries {
			values = append(
				values,
				protocol.MapEntry{Key: key, Val: protocol.NewStringProtocolValue(value)},
			)
		}
		return protocol.NewMapResponse(values)
	}
	values := make([]string, 0, len(entries)*2)
	for key, value := range entries {
		values = append(values, key, value)
	}
	return protocol.NewArrayResponse(values)
}

type HGetAllParser struct {
}

func (p *HGetAllParser) Parse(msg *protocol.Message) (command.Command, error) {
	if len(msg.Args) != 1 {
		return nil, command.NewInvalidArgumentsCount(p.Name(), 1, len(msg.Args))
	}
	key, err := msg.Args[0].AsString()
	if err != nil {
		return nil, command.NewInvalidTypeError(p.Name(), "key")
	}
	return &hGetAll{key: key}, nil
}

func (p *HGetAllParser) Name() string {
	return "HGETALL"
}
