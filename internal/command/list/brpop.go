package list

import (
	"avacado/internal/command"
	"avacado/internal/protocol"
	"avacado/internal/storage"
	"avacado/internal/storage/lists"
	"context"
	"strconv"
	"time"
)

type BRPop struct {
	Keys    []string
	Timeout float64
}

func (b *BRPop) Execute(ctx context.Context, s storage.Storage) *protocol.Response {
	// Try immediate pop from each key in order.
	for _, key := range b.Keys {
		vals, _ := s.Lists().RPop(ctx, key, 1)
		if len(vals) > 0 {
			return protocol.NewArrayResponse([]interface{}{key, vals[0]})
		}
	}

	// No immediate data — register as a blocked client via the executor's BlockRegistry.
	registry, ok := command.BlockRegistryFromContext(ctx)
	if !ok {
		return protocol.NewNullBulkStringResponse()
	}
	blockCh, cancelFn := registry.RegisterBlockedClient(b.Keys, lists.Right)

	// Start a timeout goroutine only when a finite timeout is set.
	if b.Timeout > 0 {
		go func() {
			select {
			case <-time.After(time.Duration(b.Timeout * float64(time.Second))):
			case <-ctx.Done():
			}
			cancelFn()
		}()
	}

	return &protocol.Response{BlockCh: blockCh}
}

type BRPopParser struct{}

func NewBRPopParser() *BRPopParser {
	return &BRPopParser{}
}

func (p *BRPopParser) Parse(msg *protocol.Message) (command.Command, error) {
	if len(msg.Args) < 2 {
		return nil, command.NewInvalidArgumentsCount(p.Name(), 2, len(msg.Args))
	}

	timeout, err := strconv.ParseFloat(msg.Args[len(msg.Args)-1], 64)
	if err != nil || timeout < 0 {
		return nil, command.NewInvalidTypeError(p.Name(), "timeout")
	}

	keys := msg.Args[:len(msg.Args)-1]
	return &BRPop{Keys: keys, Timeout: timeout}, nil
}

func (p *BRPopParser) Name() string {
	return "BRPOP"
}
