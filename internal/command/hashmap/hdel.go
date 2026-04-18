package hashmap

import (
	"avacado/internal/command"
	"avacado/internal/protocol"
	"avacado/internal/storage"
	"context"
)

type HDel struct {
	key    string
	fields []string
}

func (h *HDel) Execute(ctx context.Context, storage storage.Storage) *protocol.Response {
	deletedCount, err := storage.Maps().HDel(ctx, h.key, h.fields)
	if err != nil {
		return protocol.NewErrorResponse(err)
	}
	return protocol.NewNumberResponse(int64(deletedCount))
}

type HDelParser struct {
}

func (h *HDelParser) Parse(msg *protocol.Message) (command.Command, error) {
	if len(msg.Args) < 2 {
		return nil, command.NewInvalidArgumentsCount(h.Name(), 2, len(msg.Args))
	}
	key, err := msg.Args[0].AsString()
	if err != nil {
		return nil, command.NewInvalidTypeError(h.Name(), "key")
	}
	var fields []string
	for _, arg := range msg.Args[1:] {
		field, err := arg.AsString()
		if err != nil {
			return nil, command.NewInvalidTypeError(h.Name(), "field")
		}
		fields = append(fields, field)
	}
	return &HDel{key: key, fields: fields}, nil
}

func (h *HDelParser) Name() string {
	return "HDEL"
}

func NewHDelParser() *HDelParser {
	return &HDelParser{}
}
