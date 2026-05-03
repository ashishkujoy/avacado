package executor

import (
	"avacado/internal/command"
	"avacado/internal/protocol"
	"avacado/internal/storage"
	"avacado/internal/storage/lists"
	"context"
	"sync/atomic"
)

// blockedClient represents a BLPOP/BRPOP client waiting for data on one or more keys.
// The executor owns all blocked clients; they are only accessed from the executor goroutine
// except for cancelled (atomic) which is written by the timeout goroutine.
type blockedClient struct {
	keys      []string
	direction string
	resultCh  chan *protocol.Response
	cancelled atomic.Bool
}

type commandRequest struct {
	cmd    command.Command
	ctx    context.Context
	respCh chan *protocol.Response
}

// Executor serialises command execution through a single goroutine so storage
// needs no internal locking. It also manages the blocked-client queue for
// BLPOP/BRPOP: when a push arrives the executor delivers to any waiting client.
type Executor struct {
	queue          chan commandRequest
	store          storage.Storage
	blockedClients map[string][]*blockedClient // key → waiting clients (FIFO)
}

func New(store storage.Storage) *Executor {
	return &Executor{
		queue:          make(chan commandRequest, 1024),
		store:          store,
		blockedClients: make(map[string][]*blockedClient),
	}
}

// Run processes commands one at a time. Call as a goroutine; exits when ctx is cancelled.
func (e *Executor) Run(ctx context.Context) {
	for {
		select {
		case req := <-e.queue:
			execCtx := command.ContextWithBlockRegistry(req.ctx, e)
			resp := req.cmd.Execute(execCtx, e.store)
			// After a push command, check if any blocked BLPOP/BRPOP can be served.
			if pusher, ok := req.cmd.(interface{ PushedKey() string }); ok {
				if key := pusher.PushedKey(); key != "" {
					e.tryUnblockClient(key)
				}
			}
			req.respCh <- resp
		case <-ctx.Done():
			return
		}
	}
}

// Submit enqueues cmd and blocks until the executor returns a response.
// For blocking commands (BLPOP/BRPOP with no immediate data), the returned
// Response has a non-nil BlockCh; the caller must wait on that channel.
func (e *Executor) Submit(ctx context.Context, cmd command.Command) *protocol.Response {
	respCh := make(chan *protocol.Response, 1)
	e.queue <- commandRequest{cmd: cmd, ctx: ctx, respCh: respCh}
	return <-respCh
}

// SubmitAsync enqueues cmd without waiting for a response (fire-and-forget).
func (e *Executor) SubmitAsync(cmd command.Command) {
	respCh := make(chan *protocol.Response, 1)
	select {
	case e.queue <- commandRequest{cmd: cmd, ctx: context.Background(), respCh: respCh}:
	default:
		// Queue full — skip (e.g. a TTL cleanup tick); will retry on next tick.
	}
}

// RegisterBlockedClient implements command.BlockRegistry. It is called from
// within Execute(), which runs inside the executor goroutine, so no locking
// is needed for the blocked-clients map.
func (e *Executor) RegisterBlockedClient(keys []string, direction string) (<-chan *protocol.Response, context.CancelFunc) {
	resultCh := make(chan *protocol.Response, 1)
	client := &blockedClient{keys: keys, direction: direction, resultCh: resultCh}
	for _, key := range keys {
		e.blockedClients[key] = append(e.blockedClients[key], client)
	}
	// The cancel function is safe to call from any goroutine: it uses CAS to
	// ensure only one caller (timeout goroutine OR executor) delivers to resultCh.
	cancelFn := func() {
		if client.cancelled.CompareAndSwap(false, true) {
			client.resultCh <- protocol.NewNullBulkStringResponse()
		}
	}
	return resultCh, cancelFn
}

// tryUnblockClient is called by the executor after a successful push to key.
// It finds the first non-cancelled waiting client, pops one element for it,
// and removes it from the blocked-clients map.
func (e *Executor) tryUnblockClient(key string) {
	clients := e.blockedClients[key]
	if len(clients) == 0 {
		return
	}

	for _, client := range clients {
		// CAS(false→true): if we win, we are responsible for delivering.
		// If already true, the timeout goroutine already cancelled this client.
		if client.cancelled.CompareAndSwap(false, true) {
			var elements [][]byte
			if client.direction == lists.Left {
				elements, _ = e.store.Lists().LPop(context.Background(), key, 1)
			} else {
				elements, _ = e.store.Lists().RPop(context.Background(), key, 1)
			}
			if len(elements) > 0 {
				client.resultCh <- protocol.NewArrayResponse([]any{key, elements[0]})
			} else {
				client.resultCh <- protocol.NewNullBulkStringResponse()
			}
			e.removeBlockedClient(key, client)
			return
		}
	}

	// All clients in the list are cancelled — prune them.
	kept := clients[:0]
	for _, c := range clients {
		if !c.cancelled.Load() {
			kept = append(kept, c)
		}
	}
	e.blockedClients[key] = kept
}

// removeBlockedClient removes target from the blocked-clients map for all of its watched keys.
func (e *Executor) removeBlockedClient(_ string, target *blockedClient) {
	for _, key := range target.keys {
		others := e.blockedClients[key]
		kept := others[:0]
		for _, c := range others {
			if c != target {
				kept = append(kept, c)
			}
		}
		e.blockedClients[key] = kept
	}
}
