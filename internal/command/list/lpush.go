package list

import (
	"avacado/internal/command"
	"avacado/internal/protocol"
	"avacado/internal/storage"
	"context"
)

type LPush struct {
	Key    string
	Values [][]byte
}

func (l *LPush) Execute(ctx context.Context, storage storage.Storage) *protocol.Response {
	size, err := storage.Lists().LPush(ctx, l.Key, l.Values...)
	if err != nil {
		return protocol.NewErrorResponse(err)
	}
	return protocol.NewNumberResponse(int64(size))
}

type LPushParser struct {
}

func NewLPushParser() *LPushParser {
	return &LPushParser{}
}

func (l *LPushParser) Parse(msg *protocol.Message) (command.Command, error) {
	if len(msg.Args) < 2 {
		return nil, command.NewInvalidArgumentsCount("LPUSH", 2, len(msg.Args))
	}
	values := make([][]byte, len(msg.Args)-1)
	for i, arg := range msg.Args[1:] {
		values[i] = []byte(arg)
	}
	return &LPush{Key: msg.Args[0], Values: values}, nil
}

func (l *LPushParser) Name() string {
	return "LPUSH"
}
