package kv

import "context"

type ErrorType string

const (
	KeyAlreadyExistsErrorType ErrorType = "KEY_ALREADY_EXISTS"
	KeyNotPresentErrorType              = "KEY_NOT_PRESENT"
)

type SetOptions struct {
	NX bool
	XX bool
}

func NewSetOptions(nx bool, xx bool) *SetOptions {
	return &SetOptions{NX: nx, XX: xx}
}

//go:generate mockgen -source=store.go -destination=mocks/store.go -package=mockkv
type Store interface {
	Set(ctx context.Context, key string, value []byte, options *SetOptions) error
	Get(ctx context.Context, key string) ([]byte, error)
}
