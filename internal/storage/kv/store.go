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

func NewSetOptions() *SetOptions {
	return &SetOptions{}
}

func (s *SetOptions) WithNX() *SetOptions {
	s.NX = true
	return s
}

func (s *SetOptions) WithXX() *SetOptions {
	s.XX = true
	return s
}

//go:generate mockgen -source=store.go -destination=mock/store.go -package=mockkv
type Store interface {
	Set(ctx context.Context, key string, value []byte, options *SetOptions) error
	Get(ctx context.Context, key string) ([]byte, error)
}
