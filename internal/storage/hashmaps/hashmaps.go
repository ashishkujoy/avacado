package hashmaps

import "context"

//go:generate sh -c "rm -f mock/hashmaps.go && mockgen -source=hashmaps.go -destination=mock/hashmaps.go -package=mockhashmaps"
type HashMaps interface {
	HSet(ctx context.Context, name string, keyValues []string) int
	HGet(ctx context.Context, name string, field string) ([]byte, error)
	HGetAll(ctx context.Context, name string) (map[string]string, error)
	HDel(ctx context.Context, key string, fields []string) (int, error)
	HExists(ctx context.Context, key string, field string) int
	HIncrBy(ctx context.Context, key string, field string, increment int64) (int64, error)
	HMGet(ctx context.Context, key string, fields []string) []any
}
