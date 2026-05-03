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

type BLPop struct {
	Keys    []string
	Timeout float64
}

func (b *BLPop) Execute(ctx context.Context, s storage.Storage) *protocol.Response {
	// Try immediate pop from each key in order.
	for _, key := range b.Keys {
		vals, _ := s.Lists().LPop(ctx, key, 1)
		if len(vals) > 0 {
			return protocol.NewArrayResponse([]interface{}{key, vals[0]})
		}
	}

	// No immediate data — register as a blocked client via the executor's BlockRegistry.
	registry, ok := command.BlockRegistryFromContext(ctx)
	if !ok {
		return protocol.NewNullBulkStringResponse()
	}
	blockCh, cancelFn := registry.RegisterBlockedClient(b.Keys, lists.Left)

	// Start a timeout goroutine only when a finite timeout is set.
	// For timeout=0 the client waits indefinitely until a push arrives.
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
