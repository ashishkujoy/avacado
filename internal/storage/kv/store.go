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
	PX int64
}

func NewSetOptions() *SetOptions {
	return &SetOptions{}
}

// WithNX set nx option which allow setting value if key does not exist already
func (s *SetOptions) WithNX() *SetOptions {
	s.NX = true
	return s
}

// WithXX set xx option which allow setting value only if the key already exists
func (s *SetOptions) WithXX() *SetOptions {
	s.XX = true
	return s
}

// WithEX set ex option set value expiry time in seconds
func (s *SetOptions) WithEX(time int64) *SetOptions {
	s.EX = time
	return s
}

//go:generate sh -c "rm -f mock/store.go && mockgen -source=store.go -destination=mock/store.go -package=mockkv"
type Store interface {
	Set(ctx context.Context, key string, value []byte, options *SetOptions) error
	Get(ctx context.Context, key string) ([]byte, error)
	GetTTL(key string) (int64, error)
}
