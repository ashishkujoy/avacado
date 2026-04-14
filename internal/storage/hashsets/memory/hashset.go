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

type HashSet interface {
	Set(key string, value string)
	Get(key string) ([]byte, bool)
	Size() int
}

type ListPackBasedHashSet struct {
	mu       sync.RWMutex
	lp       *listpack.ListPack
	hash     map[string]string
	encoding encodingType
	size     int
}

func NewHashSet() *ListPackBasedHashSet {
	return &ListPackBasedHashSet{
		mu:       sync.RWMutex{},
		lp:       listpack.NewListPack(defaultMaxListPackSize),
		encoding: listpackEncoding,
	}
}

func (h *ListPackBasedHashSet) Set(key, value string) {
	h.mu.Lock()
	defer func() {
		h.size++
		h.mu.Unlock()
	}()

	if h.size == maxEntryCount && h.encoding != hashEncoding {
		_ = h._migrateToHashMap()
	}

	switch h.encoding {
	case hashEncoding:
		h.hash[key] = value
	case listpackEncoding:
		h.setInListPack(key, value)
	}
}

func (h *ListPackBasedHashSet) setInListPack(key string, value string) {
	keyIndex, keyExists := h.lp.IndexOf(key, true)
	var insertionError error

	if !keyExists {
		if h.lp.EncodedSize([]byte(key)) > maxEntrySize || h.lp.EncodedSize([]byte(value)) > maxEntrySize {
			_ = h._migrateToHashMap()
			h.hash[key] = value
			return
		}
		_, insertionError = h.lp.PushAllOrNone([]byte(key), []byte(value))
	} else {
		if h.lp.EncodedSize([]byte(value)) > maxEntrySize {
			_ = h._migrateToHashMap()
			h.hash[key] = value
			return
		}
		insertionError = h.lp.ReplaceAt(keyIndex+1, []byte(value))
	}

	if insertionError != nil {
		_ = h._migrateToHashMap()
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

func (h *ListPackBasedHashSet) Get(key string) ([]byte, bool) {
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

func (h *ListPackBasedHashSet) Size() int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	return h.size
}

// _migrateToHashMap converts the underlying listPack to hashmap
// it should be called after acquiring lp.mu
// ideally it should be called from set methods only
func (h *ListPackBasedHashSet) _migrateToHashMap() error {
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
