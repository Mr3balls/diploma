package handler

import (
	"net/http"

	"esports-backend/internal/apperror"

	"github.com/go-chi/chi/v5"
)

type NotificationHandler struct{ deps Deps }

func NewNotificationHandler(deps Deps) *NotificationHandler { return &NotificationHandler{deps: deps} }

func (h *NotificationHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := mustUserID(r)
	if userID == "" {
		writeError(w, apperror.Unauthorized("missing auth context"))
		return
	}
	limit, offset := pageParams(r)
	items, err := h.deps.Notifications.List(r.Context(), userID, limit, offset)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"items": items})
}

func (h *NotificationHandler) UnreadCount(w http.ResponseWriter, r *http.Request) {
	userID := mustUserID(r)
	if userID == "" {
		writeError(w, apperror.Unauthorized("missing auth context"))
		return
	}
	count, err := h.deps.Notifications.UnreadCount(r.Context(), userID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]int{"count": count})
}

func (h *NotificationHandler) Read(w http.ResponseWriter, r *http.Request) {
	userID := mustUserID(r)
	if userID == "" {
		writeError(w, apperror.Unauthorized("missing auth context"))
		return
	}
	id := chi.URLParam(r, "id")
	if err := h.deps.Notifications.Read(r.Context(), userID, id); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "notification marked read"})
}

func (h *NotificationHandler) ReadAll(w http.ResponseWriter, r *http.Request) {
	userID := mustUserID(r)
	if userID == "" {
		writeError(w, apperror.Unauthorized("missing auth context"))
		return
	}
	if err := h.deps.Notifications.ReadAll(r.Context(), userID); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "all notifications marked read"})
}

func (h *NotificationHandler) Action(w http.ResponseWriter, r *http.Request) {
	userID := mustUserID(r)
	if userID == "" {
		writeError(w, apperror.Unauthorized("missing auth context"))
		return
	}
	id := chi.URLParam(r, "id")
	if err := h.deps.Notifications.Act(r.Context(), userID, id); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "notification action recorded"})
}
