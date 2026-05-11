package handler

import (
	"net/http"

	"esports-backend/internal/apperror"
	"esports-backend/internal/entity"
	"esports-backend/internal/transport/http/middleware"

	"github.com/go-chi/chi/v5"
)

type AdminHandler struct{ deps Deps }

func NewAdminHandler(deps Deps) *AdminHandler { return &AdminHandler{deps: deps} }

func (h *AdminHandler) guardAdmin(r *http.Request) error {
	user := middleware.CurrentUser(r.Context())
	if user == nil {
		return apperror.Unauthorized("missing auth context")
	}
	if !middleware.HasPlatformRole(user, entity.PlatformRolePlatformAdmin) {
		return apperror.Forbidden("platform admin role required")
	}
	return nil
}

func (h *AdminHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	if err := h.guardAdmin(r); err != nil {
		writeError(w, err)
		return
	}
	limit, offset := pageParams(r)
	result, err := h.deps.Admin.ListUsers(r.Context(), limit, offset)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"items": result.Items, "total": result.Total})
}

func (h *AdminHandler) BlockUser(w http.ResponseWriter, r *http.Request) {
	if err := h.guardAdmin(r); err != nil {
		writeError(w, err)
		return
	}
	userID := chi.URLParam(r, "id")
	if err := h.deps.Admin.BlockUser(r.Context(), userID); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "user blocked"})
}

func (h *AdminHandler) UnblockUser(w http.ResponseWriter, r *http.Request) {
	if err := h.guardAdmin(r); err != nil {
		writeError(w, err)
		return
	}
	userID := chi.URLParam(r, "id")
	if err := h.deps.Admin.UnblockUser(r.Context(), userID); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "user unblocked"})
}

func (h *AdminHandler) ListTournaments(w http.ResponseWriter, r *http.Request) {
	if err := h.guardAdmin(r); err != nil {
		writeError(w, err)
		return
	}
	limit, offset := pageParams(r)
	result, err := h.deps.Admin.ListTournaments(r.Context(), limit, offset)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"items": result.Items, "total": result.Total})
}
