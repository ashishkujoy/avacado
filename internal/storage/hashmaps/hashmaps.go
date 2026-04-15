package hashmaps

import "context"

//go:generate sh -c "rm -f mock/hashmaps.go && mockgen -source=hashmaps.go -destination=mock/hashmaps.go -package=mockhashmaps"
type HashMaps interface {
	HSet(ctx context.Context, name string, keyValues []string) int
}
