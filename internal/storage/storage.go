package storage

import (
	"avacado/internal/storage/kv"
	"avacado/internal/storage/kv/memory"
)

//go:generate sh -c "rm -f mock/storage.go && mockgen -source=storage.go -destination=mock/storage.go -package=mocksstorage"
type Storage interface {
	KV() kv.Store
}

type DefaultStorage struct {
	kv *memory.KVMemoryStore
}

func (d DefaultStorage) KV() kv.Store {
	return d.kv
}

func NewDefaultStorage() DefaultStorage {
	return DefaultStorage{
		kv: memory.NewKVMemoryStore(),
	}
}
