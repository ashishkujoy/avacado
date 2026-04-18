package list

import (
	"avacado/internal/command"
	"avacado/internal/protocol"
	"avacado/internal/storage"
	"context"
	"strconv"
	"time"
)

type BLPop struct {
	Keys    []string
	Timeout float64
}

func (b *BLPop) Execute(ctx context.Context, storage storage.Storage) *protocol.Response {
	lists := storage.Lists()

	timeout := ctx
	if b.Timeout != 0.0 {
		t, cancel := context.WithTimeout(ctx, time.Duration(b.Timeout*float64(time.Second)))
		timeout = t
		defer cancel()
	}
	ch := lists.BlPop(timeout, b.Keys)
	data, ok := <-ch
	if !ok {
		return protocol.NewNullBulkStringResponse()
	}
	return protocol.NewArrayResponse([]interface{}{data.Key, data.Value})
}

type BLPopParser struct{}

func NewBLPopParser() *BLPopParser {
	return &BLPopParser{}
}

func (p *BLPopParser) Parse(msg *protocol.Message) (command.Command, error) {
	if len(msg.Args) < 2 {
		return nil, command.NewInvalidArgumentsCount(p.Name(), 2, len(msg.Args))
	}

	timeout, err := strconv.ParseFloat(msg.Args[len(msg.Args)-1], 64)
	if err != nil || timeout < 0 {
		return nil, command.NewInvalidTypeError(p.Name(), "timeout")
	}

	keys := msg.Args[:len(msg.Args)-1]
	return &BLPop{Keys: keys, Timeout: timeout}, nil
}

func (p *BLPopParser) Name() string {
	return "BLPOP"
}
