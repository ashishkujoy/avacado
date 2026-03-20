package memory

import "avacado/internal/storage/lists"

// waitEntry represents a BLPOP client blocked on one or more keys.
// All operations on listWaiter must be called under ListMemoryStore.mu (write lock).
type waitEntry struct {
	keys     []string
	resultCh chan lists.ListNameToItem
	served   chan struct{} // closed when a result is delivered
}

// listWaiter tracks clients blocked in BLPOP.
// It has no mutex of its own; callers must hold ListMemoryStore.mu.Lock().
type listWaiter struct {
	awaitingClients map[string][]*waitEntry
}

func newListWaiter() *listWaiter {
	return &listWaiter{
		awaitingClients: make(map[string][]*waitEntry),
	}
}

func (w *listWaiter) register(entry *waitEntry) {
	for _, key := range entry.keys {
		w.awaitingClients[key] = append(w.awaitingClients[key], entry)
	}
}

// deregister removes entry from all of its watched keys. Safe to call multiple times.
func (w *listWaiter) deregister(entry *waitEntry) {
	for _, key := range entry.keys {
		clients := w.awaitingClients[key]
		kept := clients[:0]
		for _, c := range clients {
			if c != entry {
				kept = append(kept, c)
			}
		}
		w.awaitingClients[key] = kept
	}
}

// firstWaiter returns the first waiting client for key, or nil if none.
func (w *listWaiter) firstWaiter(key string) *waitEntry {
	if entries := w.awaitingClients[key]; len(entries) > 0 {
		return entries[0]
	}
	return nil
}
