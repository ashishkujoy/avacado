package memory

import (
	"avacado/internal/storage/listpack"
	"strconv"
	"sync"
)

const defaultMaxListPackSize = 1024 * 8
const maxEntryCount = 128
const maxEntrySize = 64

type encodingType = int

const (
	listpackEncoding encodingType = 0
	hashEncoding     encodingType = 1
)

type HashMap struct {
	mu       sync.RWMutex
	lp       *listpack.ListPack
	hash     map[string]string
	encoding encodingType
	size     int
}

func NewHashMap() *HashMap {
	return &HashMap{
		mu:       sync.RWMutex{},
		lp:       listpack.NewListPack(defaultMaxListPackSize),
		encoding: listpackEncoding,
	}
}

func (h *HashMap) Set(key, value string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.size >= maxEntryCount && h.encoding != hashEncoding {
		_ = h.migrateToHashMap()
	}

	switch h.encoding {
	case hashEncoding:
		h.hash[key] = value
	case listpackEncoding:
		h.setInListPack(key, value)
	}

	h.size++
}

func (h *HashMap) setInListPack(key, value string) {
	keyIndex, keyExists := h.lp.IndexOf(key, true)

	var needsMigration bool
	if !keyExists {
		if h.lp.EncodedSize([]byte(key)) > maxEntrySize || h.lp.EncodedSize([]byte(value)) > maxEntrySize {
			needsMigration = true
		} else {
			_, err := h.lp.PushAllOrNone([]byte(key), []byte(value))
			needsMigration = err != nil
		}
	} else {
		if h.lp.EncodedSize([]byte(value)) > maxEntrySize {
			needsMigration = true
		} else {
			needsMigration = h.lp.ReplaceAt(keyIndex+1, []byte(value)) != nil
		}
	}

	if needsMigration {
		_ = h.migrateToHashMap()
		h.hash[key] = value
	}
}

func convertToString(value interface{}) string {
	switch v := value.(type) {
	case []byte:
		return string(v)
	case int:
		return strconv.Itoa(v)
	default:
		return ""
	}
}

func convertToBytes(value interface{}) []byte {
	switch v := value.(type) {
	case []byte:
		return v
	case int:
		return []byte(strconv.Itoa(v))
	default:
		return nil
	}
}

func (h *HashMap) Get(key string) ([]byte, bool) {
	i := 0
	var v []byte
	keyFound := false
	incrementI := func() { i++ }
	_ = h.lp.Traverse(func(value interface{}) (bool, error) {
		defer incrementI()
		if i%2 == 0 {
			k := convertToString(value)
			keyFound = k == key
			return true, nil
		}
		if keyFound {
			v = convertToBytes(value)
			return false, nil
		}
		return true, nil
	})
	return v, keyFound
}

func (h *HashMap) Size() int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	return h.size
}

// migrateToHashMap converts the underlying listPack to a hashmap.
// Must be called with h.mu held. Should only be called from set methods.
func (h *HashMap) migrateToHashMap() error {
	length := h.lp.Length()
	entries, err := h.lp.LRange(0, int64(length))
	if err != nil {
		return err
	}

	h.encoding = hashEncoding
	h.hash = make(map[string]string)
	for i := 0; i < length; i += 2 {
		h.hash[string(entries[i])] = string(entries[i+1])
	}

	h.lp = nil
	return nil
}
