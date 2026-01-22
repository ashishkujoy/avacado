package kv

import "context"

type ErrorType string

const (
	KeyAlreadyExistsErrorType ErrorType = "KEY_ALREADY_EXISTS"
	KeyNotPresentErrorType              = "KEY_NOT_PRESENT"
)

// SetOptions represent options supported by set command
type SetOptions struct {
	NX bool
	XX bool
	EX int64
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

func (s *SetOptions) WithEX(time int64) *SetOptions {
	s.EX = time
	return s
}

//go:generate sh -c "rm -f mock/store.go && mockgen -source=store.go -destination=mock/store.go -package=mockkv"
type Store interface {
	Set(ctx context.Context, key string, value []byte, options *SetOptions) error
	Get(ctx context.Context, key string) ([]byte, error)
}
