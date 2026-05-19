package handler

import (
	"net/http"

	"esports-backend/internal/apperror"
	"esports-backend/internal/service"
)

type ProfileHandler struct{ deps Deps }

func NewProfileHandler(deps Deps) *ProfileHandler { return &ProfileHandler{deps: deps} }

type updateProfileRequest struct {
	FirstName string  `json:"first_name" validate:"omitempty,max=100"`
	LastName  string  `json:"last_name" validate:"omitempty,max=100"`
	Nickname  string  `json:"nickname" validate:"omitempty,min=2,max=50"`
	Phone     string  `json:"phone" validate:"omitempty,phone_ru"`
	AvatarURL *string `json:"avatar_url"`
}

func (h *ProfileHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	userID := mustUserID(r)
	if userID == "" {
		writeError(w, apperror.Unauthorized("missing auth context"))
		return
	}
	user, err := h.deps.Users.GetMe(r.Context(), userID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, user)
}

func (h *ProfileHandler) UpdateMe(w http.ResponseWriter, r *http.Request) {
	userID := mustUserID(r)
	if userID == "" {
		writeError(w, apperror.Unauthorized("missing auth context"))
		return
	}
	var req updateProfileRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}
	if err := h.deps.Validate.Struct(req); err != nil {
		writeError(w, err)
		return
	}
	user, err := h.deps.Users.UpdateMe(r.Context(), userID, service.UpdateProfileInput{FirstName: req.FirstName, LastName: req.LastName, Nickname: req.Nickname, Phone: req.Phone, AvatarURL: req.AvatarURL})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, user)
}

func (h *ProfileHandler) DeleteMe(w http.ResponseWriter, r *http.Request) {
	userID := mustUserID(r)
	if userID == "" {
		writeError(w, apperror.Unauthorized("missing auth context"))
		return
	}
	if err := h.deps.Users.DeleteMe(r.Context(), userID); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "profile deleted"})
}

func (h *ProfileHandler) GetMyStats(w http.ResponseWriter, r *http.Request) {
	userID := mustUserID(r)
	if userID == "" {
		writeError(w, apperror.Unauthorized("missing auth context"))
		return
	}
	stats, err := h.deps.Users.GetMyStats(r.Context(), userID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, stats)
}

func (h *ProfileHandler) GetMyTournaments(w http.ResponseWriter, r *http.Request) {
	userID := mustUserID(r)
	if userID == "" {
		writeError(w, apperror.Unauthorized("missing auth context"))
		return
	}
	items, err := h.deps.Users.GetMyTournaments(r.Context(), userID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"items": items})
}
