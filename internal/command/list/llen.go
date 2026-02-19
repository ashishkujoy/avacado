package list

import (
	"avacado/internal/command"
	"avacado/internal/protocol"
	"avacado/internal/storage"
	"context"
)

type LLen struct {
	Key string
}

func (l *LLen) Execute(ctx context.Context, storage storage.Storage) *protocol.Response {
	length, err := storage.Lists().Len(ctx, l.Key)
	if err != nil {
		return protocol.NewErrorResponse(err)
	}
	return protocol.NewNumberResponse(int64(length))
}

type LLenParser struct {
}

func NewLLenParser() *LLenParser {
	return &LLenParser{}
}

func (l *LLenParser) Parse(msg *protocol.Message) (command.Command, error) {
	if len(msg.Args) != 1 {
		return nil, command.NewInvalidArgumentsCount("llen", 1, len(msg.Args))
	}
	key, err := msg.Args[0].AsString()
	if err != nil {
		return nil, command.NewInvalidTypeError("llen", "key")
	}
	return &LLen{Key: key}, nil
}

func (l *LLenParser) Name() string {
	return "LLEN"
}
