package memory

import (
	"avacado/internal/storage/kv"
	"context"
	"fmt"
	"sync"
	"time"
)

type value struct {
	data   []byte
	expiry *time.Time
}

func (v *value) isExpired() bool {
	if v.expiry == nil {
		return false
	}
	return time.Now().After(*v.expiry)
}

type KVMemoryStore struct {
	store map[string]*value
	mu    sync.RWMutex
}

func NewKVMemoryStore() *KVMemoryStore {
	return &KVMemoryStore{
		store: make(map[string]*value),
		mu:    sync.RWMutex{},
	}
}

func (k *KVMemoryStore) Set(_ context.Context, key string, data []byte, options *kv.SetOptions) error {
	k.mu.Lock()
	defer k.mu.Unlock()
	_, keyAlreadyExists := k.store[key]
	if keyAlreadyExists && options.NX {
		return NewKeyAlreadyExistsError(key)
	}
	if !keyAlreadyExists && options.XX {
		return NewKeyNotPresentError(key)
	}
	k.store[key] = &value{data: data, expiry: nil}
	return nil
}

func (k *KVMemoryStore) Get(_ context.Context, key string) ([]byte, error) {
	k.mu.RLock()
	defer k.mu.RUnlock()
	if value, ok := k.store[key]; ok {
		return value.data, nil
	}
	return nil, nil
}

func NewKeyAlreadyExistsError(key string) error {
	return fmt.Errorf("set operation failed: key = %s, %s", key, kv.KeyAlreadyExistsErrorType)
}

func NewKeyNotPresentError(key string) error {
	return fmt.Errorf("set operation failed: key = %s, %s", key, kv.KeyNotPresentErrorType)
}
