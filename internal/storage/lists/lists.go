package lists

import "context"

type ListNameToItem struct {
	Key   string
	Value []byte
}

type Direction = string

const (
	Left  = "left"
	Right = "right"
)

// Lists represent list data structure supported by the storage
//
//go:generate sh -c "rm -f mock/lists.go && mockgen -source=lists.go -destination=mock/lists.go -package=mocklists"
type Lists interface {
	LPush(ctx context.Context, key string, values ...[]byte) (int, error)
	RPush(ctx context.Context, key string, values ...[]byte) (int, error)
	LPop(ctx context.Context, key string, count int) ([][]byte, error)
	RPop(ctx context.Context, key string, count int) ([][]byte, error)
	Len(ctx context.Context, key string) (int, error)
	LIndex(ctx context.Context, key string, index int) ([]byte, error)
	BlPop(ctx context.Context, keys []string) <-chan ListNameToItem
	LRange(ctx context.Context, key string, start, end int64) ([][]byte, error)
	LMove(ctx context.Context, source, destination string, sourceDirection, destinationDirection Direction) ([]byte, error)
}
