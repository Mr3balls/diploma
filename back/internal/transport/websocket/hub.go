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

	"esports-backend/internal/entity"
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

// chatClient holds a single per-tournament chat SSE subscriber.
type chatClient struct {
	ch           chan *entity.TournamentMessage
	tournamentID string
}

// Hub manages SSE subscriptions per tournament slug and broadcasts events.
type Hub struct {
	mu           sync.RWMutex
	clients      map[string]map[*client]struct{}      // slug → set of clients
	userClients  map[string]map[*userClient]struct{}  // userID → set of clients
	chatClients  map[string]map[*chatClient]struct{}  // tournamentID → set of chat clients
}

// userClient holds a single per-user SSE subscriber.
type userClient struct {
	ch     chan Event
	userID string
}

// NewHub creates an empty Hub.
func NewHub() *Hub {
	return &Hub{
		clients:     make(map[string]map[*client]struct{}),
		userClients: make(map[string]map[*userClient]struct{}),
		chatClients: make(map[string]map[*chatClient]struct{}),
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

// BroadcastToUser sends a "notification_new" event to all SSE clients of a specific user.
func (h *Hub) BroadcastToUser(userID string) {
	h.mu.RLock()
	subs := h.userClients[userID]
	h.mu.RUnlock()
	evt := Event{Type: "notification_new", Payload: nil}
	for c := range subs {
		select {
		case c.ch <- evt:
		default:
		}
	}
}

func (h *Hub) subscribeUser(userID string) *userClient {
	c := &userClient{ch: make(chan Event, 16), userID: userID}
	h.mu.Lock()
	if h.userClients[userID] == nil {
		h.userClients[userID] = make(map[*userClient]struct{})
	}
	h.userClients[userID][c] = struct{}{}
	h.mu.Unlock()
	return c
}

func (h *Hub) unsubscribeUser(c *userClient) {
	h.mu.Lock()
	delete(h.userClients[c.userID], c)
	if len(h.userClients[c.userID]) == 0 {
		delete(h.userClients, c.userID)
	}
	h.mu.Unlock()
	close(c.ch)
}

// ServeUserSSE holds an SSE connection for the given userID,
// sending "notification_new" events whenever a notification is created for that user.
func (h *Hub) ServeUserSSE(w http.ResponseWriter, r *http.Request, userID string) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	c := h.subscribeUser(userID)
	defer h.unsubscribeUser(c)

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

// BroadcastToTournament sends a chat message event to all SSE clients watching a tournament's chat.
func (h *Hub) BroadcastToTournament(tournamentID string, msg *entity.TournamentMessage) {
	h.mu.RLock()
	subs := h.chatClients[tournamentID]
	h.mu.RUnlock()
	for c := range subs {
		select {
		case c.ch <- msg:
		default:
		}
	}
}

func (h *Hub) subscribeChat(tournamentID string) *chatClient {
	c := &chatClient{ch: make(chan *entity.TournamentMessage, 32), tournamentID: tournamentID}
	h.mu.Lock()
	if h.chatClients[tournamentID] == nil {
		h.chatClients[tournamentID] = make(map[*chatClient]struct{})
	}
	h.chatClients[tournamentID][c] = struct{}{}
	h.mu.Unlock()
	return c
}

func (h *Hub) unsubscribeChat(c *chatClient) {
	h.mu.Lock()
	delete(h.chatClients[c.tournamentID], c)
	if len(h.chatClients[c.tournamentID]) == 0 {
		delete(h.chatClients, c.tournamentID)
	}
	h.mu.Unlock()
	close(c.ch)
}

// ServeChatSSE holds an SSE connection for tournament chat, pushing new messages as JSON.
func (h *Hub) ServeChatSSE(w http.ResponseWriter, r *http.Request, tournamentID string) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	c := h.subscribeChat(tournamentID)
	defer h.unsubscribeChat(c)

	ticker := time.NewTicker(20 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case msg, open := <-c.ch:
			if !open {
				return
			}
			b, err := json.Marshal(msg)
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
