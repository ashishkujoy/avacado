package memory

import (
	"avacado/internal/storage/kv"
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"strconv"
	"sync"
	"time"
)

type encoding byte

const (
	encodingString  encoding = iota
	encodingInteger          = 1
)

type value struct {
	data   []byte
	enc    encoding
	expiry *time.Time
}

func encodeNumber(n int64) []byte {
	buf := new(bytes.Buffer)
	if err := binary.Write(buf, binary.BigEndian, n); err != nil {
		return []byte(strconv.FormatInt(n, 10))
	}
	return buf.Bytes()
}

func newIntegerValue(n int64) *value {
	return &value{data: encodeNumber(n), enc: encodingInteger}
}

func newValue(data []byte, expiry *time.Time) *value {
	// Try to parse as integer for optimized storage
	if n, err := strconv.ParseInt(string(data), 10, 64); err == nil {
		return &value{data: encodeNumber(n), enc: encodingInteger, expiry: expiry}
	}
	return &value{data: data, enc: encodingString, expiry: expiry}
}

func (v *value) AsInt64() (int64, error) {
	if v.enc == encodingInteger {
		var n int64
		err := binary.Read(bytes.NewReader(v.data), binary.BigEndian, &n)
		return n, err
	}
	return strconv.ParseInt(string(v.data), 10, 64)
}

func (v *value) Bytes() []byte {
	if v.enc == encodingString {
		return v.data
	}

	n, err := v.AsInt64()
	if err != nil {
		return v.data
	}

	return []byte(strconv.FormatInt(n, 10))
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
		k.store[key] = newValue([]byte("1"), nil)
		return 1, nil
	}

	oldValue, err := v.AsInt64()

	if err != nil {
		return 0, NewExpectsValidNumberError()
	}

	v.data = encodeNumber(oldValue + 1)
	return oldValue + 1, nil
}

func (k *KVMemoryStore) Decr(ctx context.Context, key string) (int64, error) {
	return k.DecrBy(ctx, key, 1)
}

func (k *KVMemoryStore) DecrBy(ctx context.Context, key string, decrement int64) (int64, error) {
	k.mu.Lock()
	defer k.mu.Unlock()

	v, ok := k.store[key]
	if !ok || v.isExpired() {
		// If key does not exist or is expired, initialize it to 0 - decrement
		nv := 0 - decrement
		k.store[key] = newIntegerValue(nv)
		return nv, nil
	}

	oldValue, err := v.AsInt64()
	if err != nil {
		return 0, NewExpectsValidNumberError()
	}

	nv := oldValue - decrement
	v.data = encodeNumber(nv)
	return nv, nil
}

func (k *KVMemoryStore) Del(ctx context.Context, keys ...string) (int64, error) {
	k.mu.Lock()
	defer k.mu.Unlock()

	var deletedCount int64
	for _, key := range keys {
		v, ok := k.store[key]
		// Only count as deleted if key exists and is not expired
		if ok && !v.isExpired() {
			delete(k.store, key)
			deletedCount++
		}
	}
	return deletedCount, nil
}

func (k *KVMemoryStore) Exists(ctx context.Context, keys ...string) (int64, error) {
	k.mu.RLock()
	defer k.mu.RUnlock()

	var existsCount int64
	for _, key := range keys {
		v, ok := k.store[key]
		// Only count as existing if key exists and is not expired
		if ok && !v.isExpired() {
			existsCount++
		}
	}
	return existsCount, nil
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

	k.store[key] = newValue(data, expiry)
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
