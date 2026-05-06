package hashmap

import (
	"avacado/internal/command"
	"avacado/internal/protocol"
	"avacado/internal/storage"
	"context"
	"strconv"
)

type HIncrBy struct {
	key       string
	field     string
	increment int64
}

func (h *HIncrBy) Execute(ctx context.Context, storage storage.Storage) *protocol.Response {
	value, err := storage.Maps().HIncrBy(ctx, h.key, h.field, h.increment)
	if err != nil {
		return protocol.NewErrorResponse(err)
	}
	return protocol.NewNumberResponse(value)
}

type HIncrByParser struct {
}

func (p *HIncrByParser) Parse(msg *protocol.Message) (command.Command, error) {
	if len(msg.Args) != 3 {
		return nil, command.NewInvalidArgumentsCount(p.Name(), 3, len(msg.Args))
	}

	increment, err := strconv.ParseInt(msg.Args[2], 10, 64)
	if err != nil {
		return nil, err
	}

	return &HIncrBy{key: msg.Args[0], field: msg.Args[1], increment: increment}, nil
}

func (p *HIncrByParser) Name() string {
	return "HINCRBY"
}

func NewHIncrByParser() *HIncrByParser {
	return &HIncrByParser{}
}
