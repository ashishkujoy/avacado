package memory

import (
	"sync"
)

type ListAvailabilityNotifier struct {
	mu             sync.Mutex
	awaitingClient map[string][]chan string
}

func NewListAvailabilityNotifier() *ListAvailabilityNotifier {
	return &ListAvailabilityNotifier{
		mu:             sync.Mutex{},
		awaitingClient: make(map[string][]chan string),
	}
}

// AwaitFor register for the availability notification of lists specified by the keys
func (l *ListAvailabilityNotifier) AwaitFor(keys []string) <-chan string {
	l.mu.Lock()
	defer l.mu.Unlock()
	ch := make(chan string)
	// register for all the list keys
	for _, key := range keys {
		clients, ok := l.awaitingClient[key]
		if !ok {
			clients = make([]chan string, 0)
		}
		clients = append(clients, ch)
		l.awaitingClient[key] = clients
	}
	return ch
}

// NotifyAvailable notifies the first client about the availability of element in list given key
func (l *ListAvailabilityNotifier) NotifyAvailable(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	clients, ok := l.awaitingClient[key]
	if !ok || len(clients) == 0 {
		return
	}
	clients[0] <- key
}

// DeregisterClients remove the given client from waiting lists specified by keys
// It should be called by consumer after successful data received on one of the blocked
// client or on timeout
func (l *ListAvailabilityNotifier) DeregisterClients(ch <-chan string, keys []string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	for _, key := range keys {
		clients, ok := l.awaitingClient[key]
		if !ok {
			continue
		}
		var newClients []chan string
		for _, client := range clients {
			if client != ch {
				newClients = append(newClients, client)
			}
		}
		l.awaitingClient[key] = newClients
	}
}
