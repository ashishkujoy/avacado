package hashsets

import "context"

//go:generate sh -c "rm -f mock/hashsets.go && mockgen -source=hashsets.go -destination=mock/hashsets.go -package=mockhashsets"
type HashSets interface {
	HSet(ctx context.Context, name string, keyValues []string)
}
