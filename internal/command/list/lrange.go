package list

import (
	"avacado/internal/command"
	"avacado/internal/protocol"
	"avacado/internal/storage"
	"context"
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
	key, err := msg.Args[0].AsString()
	if err != nil {
		return nil, command.NewInvalidTypeError(l.Name(), "key")
	}
	start, err := msg.Args[1].AsInt64()
	if err != nil {
		return nil, command.NewInvalidTypeError(l.Name(), "start")
	}
	end, err := msg.Args[2].AsInt64()
	if err != nil {
		return nil, command.NewInvalidTypeError(l.Name(), "end")
	}
	return &LRange{
		Key:   key,
		Start: start,
		End:   end,
	}, nil
}

func (l *LRangeParser) Name() string {
	return "LRANGE"
}
