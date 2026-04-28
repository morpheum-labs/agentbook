package events

import (
	"context"
	"encoding/json"
	"sync"
)

// Hub broadcasts JSON events to all SSE subscribers.
type Hub struct {
	mu      sync.RWMutex
	clients map[chan []byte]struct{}
}

// NewHub creates an empty hub.
func NewHub() *Hub {
	return &Hub{clients: make(map[chan []byte]struct{})}
}

// Subscribe registers a client until ctx is done.
func (h *Hub) Subscribe(ctx context.Context, buf int) <-chan []byte {
	if buf < 1 {
		buf = 8
	}
	ch := make(chan []byte, buf)
	h.mu.Lock()
	h.clients[ch] = struct{}{}
	h.mu.Unlock()
	go func() {
		<-ctx.Done()
		h.mu.Lock()
		delete(h.clients, ch)
		close(ch)
		h.mu.Unlock()
	}()
	return ch
}

// PublishJSON broadcasts a value to all subscribers (best-effort; drops if buffer full).
func (h *Hub) PublishJSON(v any) {
	b, err := json.Marshal(v)
	if err != nil {
		return
	}
	h.mu.RLock()
	defer h.mu.RUnlock()
	for ch := range h.clients {
		select {
		case ch <- b:
		default:
		}
	}
}
