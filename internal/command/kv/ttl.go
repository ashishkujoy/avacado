package kv

import (
	"avacado/internal/protocol"
	"avacado/internal/storage"
	"context"
)

// TTL returns the time to live for a command in seconds
type TTL struct {
	Key string
}

func (t *TTL) Execute(ctx context.Context, storage storage.Storage) *protocol.Response {
	ttl, err := storage.KV().GetTTL(t.Key)
	if err != nil {
		return protocol.NewNumberResponse(-2)
	}
	if ttl == -1 {
		return protocol.NewNumberResponse(-1)
	}
	ttlInSeconds := ttl / 1000
	return protocol.NewNumberResponse(ttlInSeconds)
}
