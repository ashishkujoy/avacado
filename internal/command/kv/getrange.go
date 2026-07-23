package kv

import (
	"avacado/internal/command"
	"avacado/internal/protocol"
	"avacado/internal/storage"
	"context"
	"strconv"
)

type GetRange struct {
	Key   string
	Start int64
	End   int64
}

func (g *GetRange) Execute(ctx context.Context, storage storage.Storage) *protocol.Response {
	data, err := storage.KV().GetRange(ctx, g.Key, g.Start, g.End)
	if err != nil {
		return protocol.NewErrorResponse(err)
	}
	return protocol.NewBulkStringResponse(data)
}

type GetRangeParser struct {
}

func NewGetRangeParser() *GetRangeParser {
	return &GetRangeParser{}
}

func (g *GetRangeParser) Parse(msg *protocol.Message) (command.Command, error) {
	if len(msg.Args) != 3 {
		return nil, command.NewInvalidArgumentsCount(g.Name(), 3, len(msg.Args))
	}

	start, err := strconv.ParseInt(msg.Args[1], 10, 64)
	if err != nil {
		return nil, command.NewInvalidTypeError(g.Name(), "start")
	}
	end, err := strconv.ParseInt(msg.Args[2], 10, 64)
	if err != nil {
		return nil, command.NewInvalidTypeError(g.Name(), "end")
	}
	return &GetRange{Key: msg.Args[0], Start: start, End: end}, nil
}

func (g *GetRangeParser) Name() string {
	return "GETRANGE"
}
