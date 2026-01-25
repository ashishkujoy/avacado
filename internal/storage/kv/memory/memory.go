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
	close chan interface{}
}

func NewKVMemoryStore() *KVMemoryStore {
	return &KVMemoryStore{
		store: make(map[string]*value),
		mu:    sync.RWMutex{},
	}
}

func (k *KVMemoryStore) startExpiredKeyCleanUp() {
	tick := time.Tick(time.Second)
	select {
	case <-tick:
		k.removeExpiredKeys()
	case <-k.close:
		return
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
	var expiry *time.Time
	if options.EX != 0 {
		expiryTime := time.Now().Add(time.Duration(options.EX) * time.Second)
		expiry = &expiryTime
	}
	k.store[key] = &value{data: data, expiry: expiry}
	return nil
}

func (k *KVMemoryStore) Get(_ context.Context, key string) ([]byte, error) {
	k.mu.RLock()
	defer k.mu.RUnlock()
	v, ok := k.store[key]
	if !ok {
		return nil, nil
	}
	if v.isExpired() {
		delete(k.store, key)
		return nil, nil
	}
	return v.data, nil
}

func (k *KVMemoryStore) removeExpiredKeys() {
	k.mu.Lock()
	defer k.mu.Unlock()
	for s, v := range k.store {
		if v.isExpired() {
			delete(k.store, s)
		}
	}
}

func NewKeyAlreadyExistsError(key string) error {
	return fmt.Errorf("set operation failed: key = %s, %s", key, kv.KeyAlreadyExistsErrorType)
}

func NewKeyNotPresentError(key string) error {
	return fmt.Errorf("set operation failed: key = %s, %s", key, kv.KeyNotPresentErrorType)
}
