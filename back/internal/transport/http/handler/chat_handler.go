package handler

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"esports-backend/internal/apperror"
	tok "esports-backend/internal/pkg/tokens"

	"github.com/go-chi/chi/v5"
)

type ChatHandler struct{ deps Deps }

func NewChatHandler(deps Deps) *ChatHandler { return &ChatHandler{deps: deps} }

func (h *ChatHandler) GetMessages(w http.ResponseWriter, r *http.Request) {
	tournamentID := chi.URLParam(r, "id")
	userID := mustUserID(r)
	if userID == "" {
		writeError(w, apperror.Unauthorized("missing auth context"))
		return
	}

	limit := 50
	if l := r.URL.Query().Get("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 {
			limit = n
		}
	}

	var before time.Time
	if b := r.URL.Query().Get("before"); b != "" {
		if t, err := time.Parse(time.RFC3339Nano, b); err == nil {
			before = t
		}
	}

	msgs, err := h.deps.Chat.GetMessages(r.Context(), tournamentID, userID, limit, before)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"items": msgs})
}

func (h *ChatHandler) SendMessage(w http.ResponseWriter, r *http.Request) {
	tournamentID := chi.URLParam(r, "id")
	userID := mustUserID(r)
	if userID == "" {
		writeError(w, apperror.Unauthorized("missing auth context"))
		return
	}

	var req struct {
		Content string `json:"content"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}

	msg, err := h.deps.Chat.SendMessage(r.Context(), tournamentID, userID, req.Content)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, msg)
}

// Stream is a public SSE endpoint; auth is done via ?token= query param.
func (h *ChatHandler) Stream(w http.ResponseWriter, r *http.Request) {
	tournamentID := chi.URLParam(r, "id")

	token := strings.TrimSpace(r.URL.Query().Get("token"))
	if token == "" {
		http.Error(w, "missing token", http.StatusUnauthorized)
		return
	}
	claims, err := tok.ParseAccessToken(h.deps.JWTSecret, token)
	if err != nil {
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return
	}
	userID := claims.UserID

	// Basic access check before opening the stream
	if _, err := h.deps.Chat.GetMessages(r.Context(), tournamentID, userID, 1, time.Time{}); err != nil {
		writeError(w, err)
		return
	}

	h.deps.Hub.ServeChatSSE(w, r, tournamentID)
}
