package handler

import (
	"net/http"

	"esports-backend/internal/apperror"

	"github.com/go-chi/chi/v5"
)

type AuditHandler struct{ deps Deps }

func NewAuditHandler(deps Deps) *AuditHandler { return &AuditHandler{deps: deps} }

func (h *AuditHandler) ListTournamentAudit(w http.ResponseWriter, r *http.Request) {
	actorUserID := mustUserID(r)
	if actorUserID == "" {
		writeError(w, apperror.Unauthorized("missing auth context"))
		return
	}
	tournamentID := chi.URLParam(r, "id")
	items, err := h.deps.Audits.ListByTournament(r.Context(), actorUserID, tournamentID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"items": items})
}
