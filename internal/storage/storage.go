package storage

import (
	"avacado/internal/storage/kv"
)

type Storage interface {
	KV() kv.Store
}
