package memory

import (
	"avacado/internal/storage/listpack"
	"strconv"
)

const defaultMaxListPackSize = 1024 * 8
const maxEntryCount = 128
const maxEntrySize = 64

type encodingType = int

const (
	listpackEncoding encodingType = 0
	hashEncoding     encodingType = 1
)

// HashMap stores key-value pairs, internally using either listpack or hash encoding.
// All methods are called exclusively by the executor goroutine — no locking needed.
type HashMap struct {
	lp       *listpack.ListPack
	hash     map[string]string
	encoding encodingType
}

func NewHashMap() *HashMap {
	return &HashMap{
		lp:       listpack.NewListPack(defaultMaxListPackSize),
		encoding: listpackEncoding,
	}
}

func (h *HashMap) Set(key, value string) int {
	existingSize := h.size()

	if existingSize >= maxEntryCount && h.encoding != hashEncoding {
		_ = h.migrateToHashMap()
	}

	switch h.encoding {
	case hashEncoding:
		h.hash[key] = value
	case listpackEncoding:
		h.setInListPack(key, value)
	}

	return h.size() - existingSize
}

func (h *HashMap) Get(key string) ([]byte, bool) {
	if h.encoding == hashEncoding {
		v, ok := h.hash[key]
		if !ok {
			return nil, false
		}
		return []byte(v), true
	}

	i := 0
	var v []byte
	keyFound := false
	incrementI := func() { i++ }
	_ = h.lp.Traverse(func(value interface{}, _, _, _ int) (bool, error) {
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

func (h *HashMap) GetAll() map[string]string {
	if h.encoding == hashEncoding {
		return copyHashMap(h.hash)
	}
	return toHashMap(h.lp)
}

func (h *HashMap) Size() int {
	return h.size()
}

func (h *HashMap) Delete(fields []string) int {
	currentSize := h.size()

	if h.encoding == hashEncoding {
		for _, key := range fields {
			delete(h.hash, key)
		}
	} else {
		for _, key := range fields {
			keyIndex, keyPresent := h.lp.IndexOf(key, true)
			if !keyPresent {
				continue
			}
			h.lp.DeleteFromIndex(keyIndex, 2)
		}
	}
	return currentSize - h.size()
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

func (h *HashMap) size() int {
	if h.encoding == hashEncoding {
		return len(h.hash)
	}

	return h.lp.Length() / 2
}

// migrateToHashMap converts the underlying listPack to a hashmap.
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

func copyHashMap(source map[string]string) map[string]string {
	destination := make(map[string]string)
	for k, v := range source {
		destination[k] = v
	}
	return destination
}

func toHashMap(source *listpack.ListPack) map[string]string {
	entries, _ := source.LRange(0, int64(source.Length()))
	destination := make(map[string]string)
	for i := 0; i < len(entries); i += 2 {
		destination[string(entries[i])] = string(entries[i+1])
	}
	return destination
}
