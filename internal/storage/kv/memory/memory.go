package memory

import (
	"avacado/internal/storage/kv"
	"context"
	"fmt"
	"strconv"
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

func (k *KVMemoryStore) Incr(ctx context.Context, key string) (int64, error) {
	k.mu.Lock()
	defer k.mu.Unlock()

	v, ok := k.store[key]
	if !ok || v.isExpired() {
		// If key does not exist or is expired, initialize it to 1
		k.store[key] = &value{data: []byte("1")}
		return 1, nil
	}

	// Try to parse existing value as integer
	oldValue, err := strconv.ParseInt(string(v.data), 10, 64)

	if err != nil {
		return 0, NewExpectsValidNumberError()
	}

	// Increment the value
	newValue := oldValue + 1
	v.data = []byte(fmt.Sprintf("%d", newValue))
	return newValue, nil
}

func NewKVMemoryStore() *KVMemoryStore {
	store := &KVMemoryStore{
		store: make(map[string]*value),
		mu:    sync.RWMutex{},
		close: make(chan interface{}),
	}
	// Start background cleanup goroutine
	go store.startExpiredKeyCleanUp()
	return store
}

func (k *KVMemoryStore) startExpiredKeyCleanUp() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			k.removeExpiredKeys()
		case <-k.close:
			return
		}
	}
}

func (k *KVMemoryStore) Set(_ context.Context, key string, data []byte, options *kv.SetOptions) ([]byte, error) {
	k.mu.Lock()
	defer k.mu.Unlock()
	oldValue, keyAlreadyExists := k.store[key]
	if keyAlreadyExists && options.NX {
		return nil, NewKeyAlreadyExistsError(key)
	}
	if !keyAlreadyExists && options.XX {
		return nil, NewKeyNotPresentError(key)
	}
	var expiry *time.Time
	if options.EX != 0 {
		expiryTime := time.Now().Add(time.Duration(options.EX) * time.Second)
		expiry = &expiryTime
	}

	k.store[key] = &value{data: data, expiry: expiry}
	var d []byte
	if keyAlreadyExists && options.Get {
		d = oldValue.data
	}
	return d, nil
}

func (k *KVMemoryStore) Get(_ context.Context, key string) ([]byte, error) {
	// First, try with read lock (common case: key exists and not expired)
	k.mu.RLock()
	v, ok := k.store[key]
	if !ok {
		k.mu.RUnlock()
		return nil, nil
	}

	// Check if expired while holding read lock
	if !v.isExpired() {
		// Happy path: key exists and not expired
		data := v.data
		k.mu.RUnlock()
		return data, nil
	}

	// Key is expired - need to delete it
	// Unlock read lock first
	k.mu.RUnlock()

	// Acquire write lock for deletion
	k.mu.Lock()
	defer k.mu.Unlock()

	// CRITICAL: Recheck after acquiring write lock (TOCTOU protection)
	// Another goroutine might have deleted or updated this key
	v, ok = k.store[key]
	if ok && v.isExpired() {
		delete(k.store, key)
	}

	return nil, nil
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

// Close gracefully shuts down the KVMemoryStore and stops background cleanup
func (k *KVMemoryStore) Close() error {
	close(k.close)
	return nil
}

// GetTTL returns the time to live for key in milliseconds
func (k *KVMemoryStore) GetTTL(key string) (int64, error) {
	k.mu.RLock()
	defer k.mu.Unlock()
	v, ok := k.store[key]
	if !ok {
		return -2, NewKeyNotPresentError(key)
	}
	if v.expiry == nil {
		return -1, nil
	}
	now := time.Now()
	return v.expiry.UnixMilli() - now.UnixMilli(), nil
}

func NewKeyAlreadyExistsError(key string) error {
	return fmt.Errorf("set operation failed: key = %s, %s", key, kv.KeyAlreadyExistsErrorType)
}

func NewKeyNotPresentError(key string) error {
	return fmt.Errorf("set operation failed: key = %s, %s", key, kv.KeyNotPresentErrorType)
}

func NewExpectsValidNumberError() error {
	return fmt.Errorf("value is not an integer or out of range")
}
