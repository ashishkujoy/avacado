package storage

import (
	"avacado/internal/storage/kv"
)

//go:generate mockgen -source=storage.go -destination=mocks/storage.go -package=mocksstorage
type Storage interface {
	KV() kv.Store
}
