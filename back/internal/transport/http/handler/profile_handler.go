package handler

import (
	"net/http"

	"esports-backend/internal/apperror"
	"esports-backend/internal/service"
)

type ProfileHandler struct{ deps Deps }

func NewProfileHandler(deps Deps) *ProfileHandler { return &ProfileHandler{deps: deps} }

type updateProfileRequest struct {
	FirstName string  `json:"first_name" validate:"required,min=2,max=100"`
	LastName  string  `json:"last_name" validate:"required,min=2,max=100"`
	Phone     string  `json:"phone" validate:"required,min=5,max=30"`
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
	user, err := h.deps.Users.UpdateMe(r.Context(), userID, service.UpdateProfileInput{FirstName: req.FirstName, LastName: req.LastName, Phone: req.Phone, AvatarURL: req.AvatarURL})
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
