// Package websocket provides a simple broadcast hub for real-time tournament updates.
// It uses Server-Sent Events (SSE) over standard HTTP so no external WebSocket
// library is required, while still satisfying the real-time contract.
package websocket

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// Event is a typed real-time update emitted by the service layer.
type Event struct {
	Type    string      `json:"type"`    // e.g. "match_result", "tournament_started"
	Payload interface{} `json:"payload"` // arbitrary JSON-serialisable data
}

// client holds a single SSE subscriber.
type client struct {
	ch   chan Event
	slug string
}

// Hub manages SSE subscriptions per tournament slug and broadcasts events.
type Hub struct {
	mu      sync.RWMutex
	clients map[string]map[*client]struct{} // slug → set of clients
}

// NewHub creates an empty Hub.
func NewHub() *Hub {
	return &Hub{
		clients: make(map[string]map[*client]struct{}),
	}
}

// Broadcast sends an event to all subscribers of a tournament.
func (h *Hub) Broadcast(slug string, evt Event) {
	h.mu.RLock()
	subs := h.clients[slug]
	h.mu.RUnlock()
	for c := range subs {
		select {
		case c.ch <- evt:
		default: // drop if client is slow
		}
	}
}

// subscribe registers a new SSE client for the given tournament slug.
func (h *Hub) subscribe(slug string) *client {
	c := &client{ch: make(chan Event, 32), slug: slug}
	h.mu.Lock()
	if h.clients[slug] == nil {
		h.clients[slug] = make(map[*client]struct{})
	}
	h.clients[slug][c] = struct{}{}
	h.mu.Unlock()
	return c
}

// unsubscribe removes a client and closes its channel.
func (h *Hub) unsubscribe(c *client) {
	h.mu.Lock()
	delete(h.clients[c.slug], c)
	if len(h.clients[c.slug]) == 0 {
		delete(h.clients, c.slug)
	}
	h.mu.Unlock()
	close(c.ch)
}

// ServeSSE handles an SSE long-poll connection for the given tournament slug.
// The client receives a stream of "data: <json>\n\n" messages.
// Call this from an HTTP handler after access-checking the tournament.
func (h *Hub) ServeSSE(w http.ResponseWriter, r *http.Request, slug string) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no") // disable nginx buffering

	c := h.subscribe(slug)
	defer h.unsubscribe(c)

	// Send a keep-alive comment every 20 s to prevent proxy timeouts.
	ticker := time.NewTicker(20 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case evt, open := <-c.ch:
			if !open {
				return
			}
			b, err := json.Marshal(evt)
			if err != nil {
				continue
			}
			fmt.Fprintf(w, "data: %s\n\n", b)
			flusher.Flush()
		case <-ticker.C:
			fmt.Fprintf(w, ": ping\n\n")
			flusher.Flush()
		case <-r.Context().Done():
			return
		}
	}
}
