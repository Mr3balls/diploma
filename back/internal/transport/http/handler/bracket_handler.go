package handler

import (
	"net/http"

	"esports-backend/internal/apperror"

	"github.com/go-chi/chi/v5"
)

type BracketHandler struct{ deps Deps }

func NewBracketHandler(deps Deps) *BracketHandler { return &BracketHandler{deps: deps} }

type reseedRequest struct {
	OrderedTeamIDs []string `json:"ordered_team_ids" validate:"required,min=2,dive,uuid"`
}

func (h *BracketHandler) Generate(w http.ResponseWriter, r *http.Request) {
	actorUserID := mustUserID(r)
	if actorUserID == "" {
		writeError(w, apperror.Unauthorized("missing auth context"))
		return
	}
	tournamentID := chi.URLParam(r, "id")
	bracket, matches, err := h.deps.Brackets.Generate(r.Context(), actorUserID, tournamentID, false, nil)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"bracket": bracket, "matches": matches})
}

func (h *BracketHandler) Regenerate(w http.ResponseWriter, r *http.Request) {
	actorUserID := mustUserID(r)
	if actorUserID == "" {
		writeError(w, apperror.Unauthorized("missing auth context"))
		return
	}
	tournamentID := chi.URLParam(r, "id")
	bracket, matches, err := h.deps.Brackets.Generate(r.Context(), actorUserID, tournamentID, true, nil)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"bracket": bracket, "matches": matches})
}

func (h *BracketHandler) Reseed(w http.ResponseWriter, r *http.Request) {
	actorUserID := mustUserID(r)
	if actorUserID == "" {
		writeError(w, apperror.Unauthorized("missing auth context"))
		return
	}
	tournamentID := chi.URLParam(r, "id")
	var req reseedRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}
	bracket, matches, err := h.deps.Brackets.Generate(r.Context(), actorUserID, tournamentID, true, req.OrderedTeamIDs)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"bracket": bracket, "matches": matches})
}

func (h *BracketHandler) AdvanceToPlayoff(w http.ResponseWriter, r *http.Request) {
	actorUserID := mustUserID(r)
	if actorUserID == "" {
		writeError(w, apperror.Unauthorized("missing auth context"))
		return
	}
	tournamentID := chi.URLParam(r, "id")
	resp, err := h.deps.Brackets.AdvanceToPlayoff(r.Context(), actorUserID, tournamentID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *BracketHandler) ResetMatch(w http.ResponseWriter, r *http.Request) {
	actorUserID := mustUserID(r)
	if actorUserID == "" {
		writeError(w, apperror.Unauthorized("missing auth context"))
		return
	}
	matchID := chi.URLParam(r, "id")
	if err := h.deps.Brackets.ResetMatch(r.Context(), actorUserID, matchID); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
}

func (h *BracketHandler) GetPlacements(w http.ResponseWriter, r *http.Request) {
	tournamentID := chi.URLParam(r, "id")
	placements, err := h.deps.Brackets.ComputePlacements(r.Context(), tournamentID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"placements": placements})
}
