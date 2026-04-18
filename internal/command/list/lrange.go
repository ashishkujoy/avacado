package list

import (
	"avacado/internal/command"
	"avacado/internal/protocol"
	"avacado/internal/storage"
	"context"
	"strconv"
)

type LRange struct {
	Key   string
	Start int64
	End   int64
}

func (l *LRange) Execute(ctx context.Context, storage storage.Storage) *protocol.Response {
	elements, err := storage.Lists().LRange(ctx, l.Key, l.Start, l.End)
	if err != nil {
		return protocol.NewErrorResponse(err)
	}
	return protocol.NewArrayResponse(elements)
}

type LRangeParser struct {
}

func NewLRangeParser() *LRangeParser {
	return &LRangeParser{}
}

func (l *LRangeParser) Parse(msg *protocol.Message) (command.Command, error) {
	if len(msg.Args) != 3 {
		return nil, command.NewInvalidArgumentsCount(l.Name(), 2, len(msg.Args))
	}
	start, err := strconv.ParseInt(msg.Args[1], 10, 64)
	if err != nil {
		return nil, command.NewInvalidTypeError(l.Name(), "start")
	}
	end, err := strconv.ParseInt(msg.Args[2], 10, 64)
	if err != nil {
		return nil, command.NewInvalidTypeError(l.Name(), "end")
	}
	return &LRange{Key: msg.Args[0], Start: start, End: end}, nil
}

func (l *LRangeParser) Name() string {
	return "LRANGE"
}
