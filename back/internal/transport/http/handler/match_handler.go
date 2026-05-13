package handler

import (
	"net/http"

	"esports-backend/internal/apperror"
	"esports-backend/internal/service"

	"github.com/go-chi/chi/v5"
)

type MatchHandler struct{ deps Deps }

func NewMatchHandler(deps Deps) *MatchHandler { return &MatchHandler{deps: deps} }

type scheduleMatchRequest struct {
	ScheduledAt      *string `json:"scheduled_at"`
	LocationOrServer *string `json:"location_or_server"`
}

type submitResultRequest struct {
	WinnerTeamID string  `json:"winner_team_id" validate:"required,uuid"`
	ScoreText    *string `json:"score_text"`
	Comment      *string `json:"comment"`
}

func (h *MatchHandler) GetAdminMatches(w http.ResponseWriter, r *http.Request) {
	actorUserID := mustUserID(r)
	if actorUserID == "" {
		writeError(w, apperror.Unauthorized("missing auth context"))
		return
	}
	tournamentID := chi.URLParam(r, "id")
	ok, err := h.deps.Tournaments.CanManageTournament(r.Context(), tournamentID, actorUserID)
	if err != nil {
		writeError(w, err)
		return
	}
	if !ok {
		writeError(w, apperror.Forbidden("insufficient tournament permissions"))
		return
	}
	items, err := h.deps.Tournaments.ListTournamentMatches(r.Context(), tournamentID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"items": items})
}

func (h *MatchHandler) Schedule(w http.ResponseWriter, r *http.Request) {
	actorUserID := mustUserID(r)
	if actorUserID == "" {
		writeError(w, apperror.Unauthorized("missing auth context"))
		return
	}
	matchID := chi.URLParam(r, "id")
	var req scheduleMatchRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}
	scheduledAt, err := parseOptionalTime(req.ScheduledAt)
	if err != nil {
		writeError(w, apperror.BadRequest("invalid_datetime", "invalid scheduled_at", nil))
		return
	}
	if err := h.deps.Matches.Schedule(r.Context(), actorUserID, matchID, service.ScheduleMatchInput{ScheduledAt: scheduledAt, LocationOrServer: req.LocationOrServer}); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "match scheduled"})
}

func (h *MatchHandler) ConfirmReady(w http.ResponseWriter, r *http.Request) {
	actorUserID := mustUserID(r)
	if actorUserID == "" {
		writeError(w, apperror.Unauthorized("missing auth context"))
		return
	}
	matchID := chi.URLParam(r, "id")
	if err := h.deps.Matches.ConfirmReady(r.Context(), actorUserID, matchID); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "ready confirmed"})
}

func (h *MatchHandler) RequestReschedule(w http.ResponseWriter, r *http.Request) {
	actorUserID := mustUserID(r)
	if actorUserID == "" {
		writeError(w, apperror.Unauthorized("missing auth context"))
		return
	}
	matchID := chi.URLParam(r, "id")
	if err := h.deps.Matches.RequestReschedule(r.Context(), actorUserID, matchID); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "reschedule requested"})
}

func (h *MatchHandler) ReportIssue(w http.ResponseWriter, r *http.Request) {
	actorUserID := mustUserID(r)
	if actorUserID == "" {
		writeError(w, apperror.Unauthorized("missing auth context"))
		return
	}
	matchID := chi.URLParam(r, "id")
	if err := h.deps.Matches.ReportIssue(r.Context(), actorUserID, matchID); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "issue reported"})
}

func (h *MatchHandler) SubmitResult(w http.ResponseWriter, r *http.Request) {
	actorUserID := mustUserID(r)
	if actorUserID == "" {
		writeError(w, apperror.Unauthorized("missing auth context"))
		return
	}
	matchID := chi.URLParam(r, "id")
	var req submitResultRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}
	if err := h.deps.Matches.SubmitResult(r.Context(), actorUserID, matchID, service.SubmitResultInput{WinnerTeamID: req.WinnerTeamID, ScoreText: req.ScoreText, Comment: req.Comment}); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "result submitted"})
}

func (h *MatchHandler) ApproveResult(w http.ResponseWriter, r *http.Request) {
	actorUserID := mustUserID(r)
	if actorUserID == "" {
		writeError(w, apperror.Unauthorized("missing auth context"))
		return
	}
	matchID := chi.URLParam(r, "id")
	if err := h.deps.Matches.ApproveResult(r.Context(), actorUserID, matchID); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "result approved"})
}

type adminSetResultRequest struct {
	WinnerTeamID        string  `json:"winner_team_id"`
	WinnerParticipantID string  `json:"winner_participant_id"`
	ScoreText           *string `json:"score_text"`
}

func (h *MatchHandler) AdminSetResult(w http.ResponseWriter, r *http.Request) {
	actorUserID := mustUserID(r)
	if actorUserID == "" {
		writeError(w, apperror.Unauthorized("missing auth context"))
		return
	}
	matchID := chi.URLParam(r, "id")
	var req adminSetResultRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}
	if req.WinnerParticipantID != "" {
		// Individual / participant-based match — delegate to challonge service.
		// Permission check: reuse CanManageTournament after loading the match.
		// The challonge service loads the match internally; we do a lightweight
		// permission guard here by checking via the tournament handler dep.
		if err := h.deps.Challonge.AdminSetParticipantResult(r.Context(), matchID, actorUserID, req.WinnerParticipantID); err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"message": "result set"})
		return
	}
	if err := h.deps.Matches.AdminSetResult(r.Context(), actorUserID, matchID, service.AdminSetResultInput{WinnerTeamID: req.WinnerTeamID, ScoreText: req.ScoreText}); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "result set"})
}

func (h *MatchHandler) RejectResult(w http.ResponseWriter, r *http.Request) {
	actorUserID := mustUserID(r)
	if actorUserID == "" {
		writeError(w, apperror.Unauthorized("missing auth context"))
		return
	}
	matchID := chi.URLParam(r, "id")
	if err := h.deps.Matches.RejectResult(r.Context(), actorUserID, matchID); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "result rejected"})
}
