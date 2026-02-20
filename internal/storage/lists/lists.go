package lists

import "context"

// Lists represent list data structure supported by the storage
//
//go:generate sh -c "rm -f mock/lists.go && mockgen -source=lists.go -destination=mock/lists.go -package=mocklists"
type Lists interface {
	LPush(ctx context.Context, key string, values ...[]byte) (int, error)
	RPush(ctx context.Context, key string, values ...[]byte) (int, error)
	RPop(ctx context.Context, key string, count int) ([][]byte, error)
	Len(ctx context.Context, key string) (int, error)
}
