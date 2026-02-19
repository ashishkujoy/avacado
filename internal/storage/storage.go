package storage

import (
	"avacado/internal/storage/kv"
	"avacado/internal/storage/kv/memory"
	"avacado/internal/storage/lists"
	memlist "avacado/internal/storage/lists/memory"
)

//go:generate sh -c "rm -f mock/storage.go && mockgen -source=storage.go -destination=mock/storage.go -package=mocksstorage"
type Storage interface {
	KV() kv.Store
	Lists() lists.Lists
}

type DefaultStorage struct {
	kv    *memory.KVMemoryStore
	lists *memlist.ListMemoryStore
}

func (d DefaultStorage) KV() kv.Store {
	return d.kv
}

func (d DefaultStorage) Lists() lists.Lists {
	return d.lists
}

func NewDefaultStorage() DefaultStorage {
	return DefaultStorage{
		kv: memory.NewKVMemoryStore(),
	}
}
