package kv

import (
	"avacado/internal/command"
	"avacado/internal/protocol"
	"avacado/internal/storage"
	"context"
	"strconv"
)

type SetRange struct {
	Key    string
	Offset int
	Value  []byte
}

func (s *SetRange) Execute(ctx context.Context, storage storage.Storage) *protocol.Response {
	length, err := storage.KV().SetRange(ctx, s.Key, s.Offset, s.Value)
	if err != nil {
		return protocol.NewErrorResponse(err)
	}
	return protocol.NewNumberResponse(int64(length))
}

type SetRangeParser struct{}

func (s *SetRangeParser) Parse(msg *protocol.Message) (command.Command, error) {
	args := msg.Args
	if len(args) != 3 {
		return nil, command.NewInvalidArgumentsCount(msg.Command, 3, len(args))
	}
	offset, err := strconv.Atoi(args[1])
	if err != nil {
		return nil, err
	}
	return &SetRange{
		Key:    args[0],
		Offset: offset,
		Value:  []byte(args[2]),
	}, nil
}

func (s *SetRangeParser) Name() string {
	return "SETRANGE"
}

func NewSetRangeParser() *SetRangeParser {
	return &SetRangeParser{}
}
