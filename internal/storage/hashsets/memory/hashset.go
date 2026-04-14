package memory

import (
	"avacado/internal/storage/listpack"
	"strconv"
	"sync"
)

const defaultMaxListPackSize = 1024 * 8

type HashSet interface {
	Set(key string, value string)
	Get(key string) ([]byte, bool)
	Size() int
}

type ListPackBasedHashSet struct {
	mu sync.RWMutex
	lp *listpack.ListPack
}

func NewListPackBasedHashSet() *ListPackBasedHashSet {
	return &ListPackBasedHashSet{
		mu: sync.RWMutex{},
		lp: listpack.NewListPack(defaultMaxListPackSize),
	}
}

func (h *ListPackBasedHashSet) Set(key, value string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	keyIndex, keyExists := h.lp.IndexOf(key, true)

	if !keyExists {
		_, _ = h.lp.Push([]byte(key))
		_, _ = h.lp.Push([]byte(value))
		return
	}

	_ = h.lp.ReplaceAt(keyIndex+1, []byte(value))
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

	return h.lp.Length() / 2
}
