package handler

import (
	"net/http"

	"esports-backend/internal/apperror"
	"esports-backend/internal/entity"
	"esports-backend/internal/service"
	mw "esports-backend/internal/transport/http/middleware"

	"github.com/go-chi/chi/v5"
)

type ChallongeHandler struct {
	svc *service.ChallongeService
	deps Deps
}

func NewChallongeHandler(deps Deps) *ChallongeHandler {
	return &ChallongeHandler{svc: deps.Challonge, deps: deps}
}

// ── helpers ───────────────────────────────────────────────────────────────────

func (h *ChallongeHandler) resolveSlug(w http.ResponseWriter, r *http.Request) (*entity.Tournament, bool) {
	slug := chi.URLParam(r, "slug")
	t, err := h.svc.GetBySlug(r.Context(), slug)
	if err != nil {
		writeError(w, err)
		return nil, false
	}
	return t, true
}

func (h *ChallongeHandler) currentUser(r *http.Request) *mw.AuthUser {
	return mw.CurrentUser(r.Context())
}

func (h *ChallongeHandler) requireUser(w http.ResponseWriter, r *http.Request) (*mw.AuthUser, bool) {
	u := h.currentUser(r)
	if u == nil {
		writeError(w, apperror.Unauthorized("authentication required"))
		return nil, false
	}
	return u, true
}

func (h *ChallongeHandler) requireManager(w http.ResponseWriter, r *http.Request, tournamentID, userID string) bool {
	role := h.svc.GetMyRole(r.Context(), tournamentID, userID)
	if role != entity.TournamentRoleOrganizer && role != entity.TournamentRoleCoOrganizer {
		writeError(w, apperror.Forbidden("organizer or co-organizer role required"))
		return false
	}
	return true
}

func (h *ChallongeHandler) requireOrganizer(w http.ResponseWriter, r *http.Request, tournamentID, userID string) bool {
	role := h.svc.GetMyRole(r.Context(), tournamentID, userID)
	if role != entity.TournamentRoleOrganizer {
		writeError(w, apperror.Forbidden("organizer role required"))
		return false
	}
	return true
}

// ── Public endpoints ──────────────────────────────────────────────────────────

// GET /c/{slug}
func (h *ChallongeHandler) GetBracket(w http.ResponseWriter, r *http.Request) {
	t, ok := h.resolveSlug(w, r)
	if !ok {
		return
	}
	userID := mustUserID(r)
	resp, err := h.svc.GetBracket(r.Context(), t.ID, userID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

// GET /c/{slug}/standings
func (h *ChallongeHandler) GetStandings(w http.ResponseWriter, r *http.Request) {
	t, ok := h.resolveSlug(w, r)
	if !ok {
		return
	}
	standings, err := h.svc.GetStandings(r.Context(), t.ID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"standings": standings})
}

// GET /c/{slug}/events  (SSE)
func (h *ChallongeHandler) ServeEvents(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	h.deps.Hub.ServeSSE(w, r, slug)
}

// ── Auth-required ─────────────────────────────────────────────────────────────

// POST /c
func (h *ChallongeHandler) CreateTournament(w http.ResponseWriter, r *http.Request) {
	u, ok := h.requireUser(w, r)
	if !ok {
		return
	}
	var req service.CreateChallongeTournamentReq
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, apperror.BadRequest("invalid_body", err.Error(), nil))
		return
	}
	t, err := h.svc.CreateTournament(r.Context(), u.UserID, req)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, t)
}

// GET /c/my-matches
func (h *ChallongeHandler) GetMyMatches(w http.ResponseWriter, r *http.Request) {
	u, ok := h.requireUser(w, r)
	if !ok {
		return
	}
	matches, err := h.svc.GetMyMatches(r.Context(), u.UserID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"items": matches})
}

// POST /c/invites/{token}/accept
func (h *ChallongeHandler) AcceptInvite(w http.ResponseWriter, r *http.Request) {
	u, ok := h.requireUser(w, r)
	if !ok {
		return
	}
	token := chi.URLParam(r, "token")
	if err := h.svc.AcceptInvite(r.Context(), token, u.UserID); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "accepted"})
}

// POST /c/{slug}/join
func (h *ChallongeHandler) Join(w http.ResponseWriter, r *http.Request) {
	u, ok := h.requireUser(w, r)
	if !ok {
		return
	}
	t, ok := h.resolveSlug(w, r)
	if !ok {
		return
	}
	if err := h.svc.Join(r.Context(), t.ID, u.UserID); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "joined"})
}

// ── Organizer / co-organizer endpoints ───────────────────────────────────────

// POST /c/{slug}/participants
func (h *ChallongeHandler) AddParticipant(w http.ResponseWriter, r *http.Request) {
	u, ok := h.requireUser(w, r)
	if !ok {
		return
	}
	t, ok := h.resolveSlug(w, r)
	if !ok {
		return
	}
	if !h.requireManager(w, r, t.ID, u.UserID) {
		return
	}
	var req service.AddParticipantReq
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, apperror.BadRequest("invalid_body", err.Error(), nil))
		return
	}
	p, err := h.svc.AddParticipant(r.Context(), t.ID, u.UserID, req)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, p)
}

// POST /c/{slug}/participants/bulk
func (h *ChallongeHandler) BulkAddParticipants(w http.ResponseWriter, r *http.Request) {
	u, ok := h.requireUser(w, r)
	if !ok {
		return
	}
	t, ok := h.resolveSlug(w, r)
	if !ok {
		return
	}
	if !h.requireManager(w, r, t.ID, u.UserID) {
		return
	}
	var body struct {
		Names []string `json:"names"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, apperror.BadRequest("invalid_body", err.Error(), nil))
		return
	}
	ps, err := h.svc.BulkAddParticipants(r.Context(), t.ID, u.UserID, body.Names)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]interface{}{"participants": ps})
}

// DELETE /c/{slug}/participants/{participantID}
func (h *ChallongeHandler) RemoveParticipant(w http.ResponseWriter, r *http.Request) {
	u, ok := h.requireUser(w, r)
	if !ok {
		return
	}
	t, ok := h.resolveSlug(w, r)
	if !ok {
		return
	}
	if !h.requireManager(w, r, t.ID, u.UserID) {
		return
	}
	pid := chi.URLParam(r, "participantID")
	if err := h.svc.RemoveParticipant(r.Context(), t.ID, pid, u.UserID); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// POST /c/{slug}/participants/shuffle
func (h *ChallongeHandler) ShuffleParticipants(w http.ResponseWriter, r *http.Request) {
	u, ok := h.requireUser(w, r)
	if !ok {
		return
	}
	t, ok := h.resolveSlug(w, r)
	if !ok {
		return
	}
	if !h.requireManager(w, r, t.ID, u.UserID) {
		return
	}
	if err := h.svc.ShuffleParticipants(r.Context(), t.ID, u.UserID); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "shuffled"})
}

// PUT /c/{slug}/participants/reorder
func (h *ChallongeHandler) ReorderParticipants(w http.ResponseWriter, r *http.Request) {
	u, ok := h.requireUser(w, r)
	if !ok {
		return
	}
	t, ok := h.resolveSlug(w, r)
	if !ok {
		return
	}
	if !h.requireManager(w, r, t.ID, u.UserID) {
		return
	}
	var body struct {
		Seeds []service.ParticipantSeedItem `json:"seeds"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, apperror.BadRequest("invalid_body", err.Error(), nil))
		return
	}
	if err := h.svc.ReorderParticipants(r.Context(), t.ID, u.UserID, body.Seeds); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "reordered"})
}

// POST /c/{slug}/start
func (h *ChallongeHandler) Start(w http.ResponseWriter, r *http.Request) {
	u, ok := h.requireUser(w, r)
	if !ok {
		return
	}
	t, ok := h.resolveSlug(w, r)
	if !ok {
		return
	}
	if !h.requireManager(w, r, t.ID, u.UserID) {
		return
	}
	if err := h.svc.Start(r.Context(), t.ID, u.UserID); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "started"})
}

// POST /c/{slug}/reset
func (h *ChallongeHandler) Reset(w http.ResponseWriter, r *http.Request) {
	u, ok := h.requireUser(w, r)
	if !ok {
		return
	}
	t, ok := h.resolveSlug(w, r)
	if !ok {
		return
	}
	if !h.requireManager(w, r, t.ID, u.UserID) {
		return
	}
	if err := h.svc.Reset(r.Context(), t.ID, u.UserID); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "reset"})
}

// POST /c/{slug}/unfinalize
func (h *ChallongeHandler) Unfinalize(w http.ResponseWriter, r *http.Request) {
	u, ok := h.requireUser(w, r)
	if !ok {
		return
	}
	t, ok := h.resolveSlug(w, r)
	if !ok {
		return
	}
	if !h.requireManager(w, r, t.ID, u.UserID) {
		return
	}
	if err := h.svc.Unfinalize(r.Context(), t.ID, u.UserID); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "unfinalized"})
}

// POST /c/{slug}/matches/{matchID}/result
func (h *ChallongeHandler) SubmitResult(w http.ResponseWriter, r *http.Request) {
	u, ok := h.requireUser(w, r)
	if !ok {
		return
	}
	t, ok := h.resolveSlug(w, r)
	if !ok {
		return
	}
	if !h.requireManager(w, r, t.ID, u.UserID) {
		return
	}
	matchID := chi.URLParam(r, "matchID")
	var req service.SubmitResultReq
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, apperror.BadRequest("invalid_body", err.Error(), nil))
		return
	}
	if err := h.svc.SubmitResult(r.Context(), t.ID, matchID, u.UserID, req); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "result_submitted"})
}

// POST /c/{slug}/matches/{matchID}/reset
func (h *ChallongeHandler) ResetMatch(w http.ResponseWriter, r *http.Request) {
	u, ok := h.requireUser(w, r)
	if !ok {
		return
	}
	t, ok := h.resolveSlug(w, r)
	if !ok {
		return
	}
	if !h.requireManager(w, r, t.ID, u.UserID) {
		return
	}
	matchID := chi.URLParam(r, "matchID")
	if err := h.svc.ResetMatch(r.Context(), t.ID, matchID, u.UserID); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "match_reset"})
}

// GET /c/{slug}/log
func (h *ChallongeHandler) GetLog(w http.ResponseWriter, r *http.Request) {
	u, ok := h.requireUser(w, r)
	if !ok {
		return
	}
	t, ok := h.resolveSlug(w, r)
	if !ok {
		return
	}
	if !h.requireManager(w, r, t.ID, u.UserID) {
		return
	}
	entries, err := h.svc.GetLog(r.Context(), t.ID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"entries": entries})
}

// POST /c/{slug}/co-organizers/invite
func (h *ChallongeHandler) InviteCoOrganizer(w http.ResponseWriter, r *http.Request) {
	u, ok := h.requireUser(w, r)
	if !ok {
		return
	}
	t, ok := h.resolveSlug(w, r)
	if !ok {
		return
	}
	if !h.requireOrganizer(w, r, t.ID, u.UserID) {
		return
	}
	var req service.InviteCoOrgReq
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, apperror.BadRequest("invalid_body", err.Error(), nil))
		return
	}
	inv, err := h.svc.InviteCoOrganizer(r.Context(), t.ID, u.UserID, req)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, inv)
}

// DELETE /c/{slug}/co-organizers/{userID}
func (h *ChallongeHandler) RemoveCoOrganizer(w http.ResponseWriter, r *http.Request) {
	u, ok := h.requireUser(w, r)
	if !ok {
		return
	}
	t, ok := h.resolveSlug(w, r)
	if !ok {
		return
	}
	if !h.requireOrganizer(w, r, t.ID, u.UserID) {
		return
	}
	targetID := chi.URLParam(r, "userID")
	if err := h.svc.RemoveCoOrganizer(r.Context(), t.ID, targetID, u.UserID); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ── Participant result reporting ──────────────────────────────────────────────

// POST /c/{slug}/matches/{matchID}/report
func (h *ChallongeHandler) ReportResult(w http.ResponseWriter, r *http.Request) {
	u, ok := h.requireUser(w, r)
	if !ok {
		return
	}
	t, ok := h.resolveSlug(w, r)
	if !ok {
		return
	}
	matchID := chi.URLParam(r, "matchID")
	var req service.SubmitResultReq
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, apperror.BadRequest("invalid_body", err.Error(), nil))
		return
	}
	rr, err := h.svc.ReportResult(r.Context(), t.ID, matchID, u.UserID, req)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, rr)
}

// POST /c/{slug}/matches/{matchID}/reports/{reportID}/approve
func (h *ChallongeHandler) ApproveReport(w http.ResponseWriter, r *http.Request) {
	u, ok := h.requireUser(w, r)
	if !ok {
		return
	}
	t, ok := h.resolveSlug(w, r)
	if !ok {
		return
	}
	if !h.requireManager(w, r, t.ID, u.UserID) {
		return
	}
	matchID := chi.URLParam(r, "matchID")
	reportID := chi.URLParam(r, "reportID")
	if err := h.svc.ApproveReport(r.Context(), t.ID, matchID, reportID, u.UserID); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "approved"})
}

// POST /c/{slug}/matches/{matchID}/reports/{reportID}/reject
func (h *ChallongeHandler) RejectReport(w http.ResponseWriter, r *http.Request) {
	u, ok := h.requireUser(w, r)
	if !ok {
		return
	}
	t, ok := h.resolveSlug(w, r)
	if !ok {
		return
	}
	if !h.requireManager(w, r, t.ID, u.UserID) {
		return
	}
	matchID := chi.URLParam(r, "matchID")
	reportID := chi.URLParam(r, "reportID")
	if err := h.svc.RejectReport(r.Context(), t.ID, matchID, reportID, u.UserID); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "rejected"})
}
