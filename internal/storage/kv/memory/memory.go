package memory

import (
	"avacado/internal/storage/kv"
	"context"
	"fmt"
	"sync"
)

type KVMemoryStore struct {
	store map[string][]byte
	mu    sync.RWMutex
}

func NewKVMemoryStore() *KVMemoryStore {
	return &KVMemoryStore{
		store: make(map[string][]byte),
		mu:    sync.RWMutex{},
	}
}

func (k *KVMemoryStore) Set(_ context.Context, key string, value []byte, options *kv.SetOptions) error {
	k.mu.Lock()
	defer k.mu.Unlock()
	_, keyAlreadyExists := k.store[key]
	if keyAlreadyExists && options.NX {
		return NewKeyAlreadyExistsError(key)
	}
	if !keyAlreadyExists && options.XX {
		return NewKeyNotPresentError(key)
	}
	k.store[key] = value
	return nil
}

func (k *KVMemoryStore) Get(_ context.Context, key string) ([]byte, error) {
	k.mu.RLock()
	defer k.mu.RUnlock()
	return k.store[key], nil
}

func NewKeyAlreadyExistsError(key string) error {
	return fmt.Errorf("set operation failed: key = %s, %s", key, kv.KeyAlreadyExistsErrorType)
}

func NewKeyNotPresentError(key string) error {
	return fmt.Errorf("set operation failed: key = %s, %s", key, kv.KeyNotPresentErrorType)
}
