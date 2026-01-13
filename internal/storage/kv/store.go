package kv

import "context"

//go:generate mockgen -source=store.go -destination=mocks/store.go -package=mockkv
type Store interface {
	Set(ctx context.Context, key string, value []byte) error
	Get(ctx context.Context, key string) ([]byte, error)
}
