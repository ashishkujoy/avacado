package storage

import (
	"avacado/internal/storage/hashmaps"
	memhash "avacado/internal/storage/hashmaps/memory"
	"avacado/internal/storage/kv"
	"avacado/internal/storage/kv/memory"
	"avacado/internal/storage/lists"
	memlist "avacado/internal/storage/lists/memory"
	"os"
	"strconv"
)

//go:generate sh -c "rm -f mock/storage.go && mockgen -source=storage.go -destination=mock/storage.go -package=mocksstorage"
type Storage interface {
	KV() kv.Store
	Lists() lists.Lists
	Maps() hashmaps.HashMaps
}

type DefaultStorage struct {
	kv    *memory.KVMemoryStore
	lists *memlist.ListMemoryStore
	maps  *memhash.HashMaps
}

func (d DefaultStorage) KV() kv.Store {
	return d.kv
}

func (d DefaultStorage) Lists() lists.Lists {
	return d.lists
}

func (d DefaultStorage) Maps() hashmaps.HashMaps {
	return d.maps
}

const defaultMaxListPackSize = 8192

func NewDefaultStorage() DefaultStorage {
	maxListPackSize := defaultMaxListPackSize
	if v, err := strconv.Atoi(os.Getenv("MAX_LIST_PACK_SIZE")); err == nil && v > 0 {
		maxListPackSize = v
	}
	return DefaultStorage{
		kv:    memory.NewKVMemoryStore(),
		lists: memlist.NewListMemoryStore(maxListPackSize),
		maps:  memhash.NewHashMaps(),
	}
}
