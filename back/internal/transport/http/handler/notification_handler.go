package handler

import (
	"net/http"
	"strings"

	"esports-backend/internal/apperror"
	tok "esports-backend/internal/pkg/tokens"

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

// Stream holds an SSE connection for the authenticated user.
// Token is passed as ?token=<access_token> because EventSource doesn't support headers.
func (h *NotificationHandler) Stream(w http.ResponseWriter, r *http.Request) {
	token := strings.TrimSpace(r.URL.Query().Get("token"))
	if token == "" {
		writeError(w, apperror.Unauthorized("missing token"))
		return
	}
	claims, err := tok.ParseAccessToken(h.deps.JWTSecret, token)
	if err != nil {
		writeError(w, apperror.Unauthorized("invalid or expired token"))
		return
	}
	h.deps.Hub.ServeUserSSE(w, r, claims.UserID)
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

// Delete soft-deletes a single notification.
func (h *NotificationHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := mustUserID(r)
	if userID == "" {
		writeError(w, apperror.Unauthorized("missing auth context"))
		return
	}
	id := chi.URLParam(r, "id")
	if err := h.deps.Notifications.Delete(r.Context(), userID, id); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "notification deleted"})
}

// DeleteAll soft-deletes all notifications for the current user.
func (h *NotificationHandler) DeleteAll(w http.ResponseWriter, r *http.Request) {
	userID := mustUserID(r)
	if userID == "" {
		writeError(w, apperror.Unauthorized("missing auth context"))
		return
	}
	if err := h.deps.Notifications.DeleteAll(r.Context(), userID); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "all notifications deleted"})
}

// GetPreferences returns the list of disabled notification types for the user.
func (h *NotificationHandler) GetPreferences(w http.ResponseWriter, r *http.Request) {
	userID := mustUserID(r)
	if userID == "" {
		writeError(w, apperror.Unauthorized("missing auth context"))
		return
	}
	disabled, err := h.deps.Notifications.GetPreferences(r.Context(), userID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string][]string{"disabled": disabled})
}

type setPreferencesRequest struct {
	Disabled []string `json:"disabled"`
}

// SetPreferences updates the disabled notification types for the user.
func (h *NotificationHandler) SetPreferences(w http.ResponseWriter, r *http.Request) {
	userID := mustUserID(r)
	if userID == "" {
		writeError(w, apperror.Unauthorized("missing auth context"))
		return
	}
	var req setPreferencesRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}
	if req.Disabled == nil {
		req.Disabled = []string{}
	}
	if err := h.deps.Notifications.SetPreferences(r.Context(), userID, req.Disabled); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "preferences updated"})
}

// GetVAPIDPublicKey returns the VAPID public key for Web Push subscription.
func (h *NotificationHandler) GetVAPIDPublicKey(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"public_key": h.deps.Notifications.VAPIDPublicKey()})
}

type registerPushRequest struct {
	Endpoint string `json:"endpoint"`
	P256dh   string `json:"p256dh"`
	Auth     string `json:"auth"`
}

// RegisterPush stores a Web Push subscription for the current user.
func (h *NotificationHandler) RegisterPush(w http.ResponseWriter, r *http.Request) {
	userID := mustUserID(r)
	if userID == "" {
		writeError(w, apperror.Unauthorized("missing auth context"))
		return
	}
	var req registerPushRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}
	if req.Endpoint == "" || req.P256dh == "" || req.Auth == "" {
		writeError(w, apperror.BadRequest("invalid_body", "endpoint, p256dh and auth are required", nil))
		return
	}
	if err := h.deps.Notifications.RegisterPush(r.Context(), userID, req.Endpoint, req.P256dh, req.Auth); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "push subscription registered"})
}

type unregisterPushRequest struct {
	Endpoint string `json:"endpoint"`
}

// UnregisterPush removes a Web Push subscription.
func (h *NotificationHandler) UnregisterPush(w http.ResponseWriter, r *http.Request) {
	userID := mustUserID(r)
	if userID == "" {
		writeError(w, apperror.Unauthorized("missing auth context"))
		return
	}
	var req unregisterPushRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}
	if err := h.deps.Notifications.UnregisterPush(r.Context(), userID, req.Endpoint); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "push subscription removed"})
}
