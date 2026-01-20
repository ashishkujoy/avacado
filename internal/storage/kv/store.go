package kv

import "context"

type ErrorType string

const KeyAlreadyExistsErrorType = "KEY_ALREADY_EXISTS"

type SetOptions struct {
	NX bool
}

func NewSetOptions(nx bool) *SetOptions {
	return &SetOptions{NX: nx}
}

//go:generate mockgen -source=store.go -destination=mocks/store.go -package=mockkv
type Store interface {
	Set(ctx context.Context, key string, value []byte, options *SetOptions) error
	Get(ctx context.Context, key string) ([]byte, error)
}
