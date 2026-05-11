package handler

import (
	"net/http"
	"strconv"
	"time"

	"esports-backend/internal/apperror"
	"esports-backend/internal/entity"
	"esports-backend/internal/service"

	"github.com/go-chi/chi/v5"
)

type TournamentHandler struct{ deps Deps }

func NewTournamentHandler(deps Deps) *TournamentHandler { return &TournamentHandler{deps: deps} }

type createTournamentRequest struct {
	Title                string  `json:"title" validate:"required,min=2,max=200"`
	Discipline           string  `json:"discipline"`
	Description          *string `json:"description"`
	Rules                *string `json:"rules"`
	Location             *string `json:"location"`
	MaxTeams             int     `json:"max_teams"`
	Format               string  `json:"format" validate:"omitempty,oneof=single_elimination double_elimination group_stage"`
	GroupCount           *int    `json:"group_count"`
	RegistrationDeadline *string `json:"registration_deadline"`
	StartAt              *string `json:"start_at"`
	Visibility           string  `json:"visibility" validate:"required,oneof=public private"`
	RegistrationMode     string  `json:"registration_mode" validate:"omitempty,oneof=team individual"`
}

type updateTournamentRequest = createTournamentRequest

type changeStatusRequest struct {
	Status string `json:"status" validate:"required,oneof=draft registration_open registration_closed bracket_generated in_progress finished cancelled ready completed"`
}

type addManagerRequest struct {
	UserID string `json:"user_id" validate:"required,uuid"`
}

func (h *TournamentHandler) ListPublic(w http.ResponseWriter, r *http.Request) {
	limit, offset := pageParams(r)
	items, err := h.deps.Tournaments.ListPublic(r.Context(), limit, offset)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"items": items})
}

func (h *TournamentHandler) GetPublic(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	requester := mustUserID(r)
	tournament, err := h.deps.Tournaments.GetPublic(r.Context(), id, requester)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, tournament)
}

func (h *TournamentHandler) GetPublicTeams(w http.ResponseWriter, r *http.Request) {
	tournamentID := chi.URLParam(r, "id")
	teams, err := h.deps.Tournaments.ListTournamentTeams(r.Context(), tournamentID, false)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"items": teams})
}

func (h *TournamentHandler) GetPublicMatches(w http.ResponseWriter, r *http.Request) {
	tournamentID := chi.URLParam(r, "id")
	matches, err := h.deps.Tournaments.ListTournamentMatches(r.Context(), tournamentID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"items": matches})
}

func (h *TournamentHandler) GetBracket(w http.ResponseWriter, r *http.Request) {
	tournamentID := chi.URLParam(r, "id")
	bracket, matches, err := h.deps.Brackets.GetBracket(r.Context(), tournamentID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"bracket": bracket, "matches": matches})
}

func (h *TournamentHandler) Create(w http.ResponseWriter, r *http.Request) {
	actorUserID := mustUserID(r)
	if actorUserID == "" {
		writeError(w, apperror.Unauthorized("missing auth context"))
		return
	}
	var req createTournamentRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}
	if err := h.deps.Validate.Struct(req); err != nil {
		writeError(w, err)
		return
	}
	registrationDeadline, err := parseOptionalTime(req.RegistrationDeadline)
	if err != nil {
		writeError(w, apperror.BadRequest("invalid_datetime", "invalid registration_deadline", nil))
		return
	}
	startAt, err := parseOptionalTime(req.StartAt)
	if err != nil {
		writeError(w, apperror.BadRequest("invalid_datetime", "invalid start_at", nil))
		return
	}
	if req.Format == "" {
		req.Format = "single_elimination"
	}
	maxTeams := req.MaxTeams
	if maxTeams < 2 {
		maxTeams = 8
	}
	tournament := &entity.Tournament{Title: req.Title, Discipline: req.Discipline, Description: req.Description, Rules: req.Rules, Location: req.Location, MaxTeams: maxTeams, Format: req.Format, GroupCount: req.GroupCount, RegistrationDeadline: registrationDeadline, StartAt: startAt, Visibility: req.Visibility, RegistrationMode: req.RegistrationMode}
	created, err := h.deps.Tournaments.Create(r.Context(), actorUserID, toCreateTournamentInput(tournament))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (h *TournamentHandler) Update(w http.ResponseWriter, r *http.Request) {
	actorUserID := mustUserID(r)
	if actorUserID == "" {
		writeError(w, apperror.Unauthorized("missing auth context"))
		return
	}
	id := chi.URLParam(r, "id")
	var req updateTournamentRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}
	if err := h.deps.Validate.Struct(req); err != nil {
		writeError(w, err)
		return
	}
	registrationDeadline, err := parseOptionalTime(req.RegistrationDeadline)
	if err != nil {
		writeError(w, apperror.BadRequest("invalid_datetime", "invalid registration_deadline", nil))
		return
	}
	startAt, err := parseOptionalTime(req.StartAt)
	if err != nil {
		writeError(w, apperror.BadRequest("invalid_datetime", "invalid start_at", nil))
		return
	}
	if req.Format == "" {
		req.Format = "single_elimination"
	}
	maxTeams := req.MaxTeams
	if maxTeams < 2 {
		maxTeams = 8
	}
	tournament := &entity.Tournament{ID: id, Title: req.Title, Discipline: req.Discipline, Description: req.Description, Rules: req.Rules, Location: req.Location, MaxTeams: maxTeams, Format: req.Format, GroupCount: req.GroupCount, RegistrationDeadline: registrationDeadline, StartAt: startAt, Visibility: req.Visibility, RegistrationMode: req.RegistrationMode}
	updated, err := h.deps.Tournaments.Update(r.Context(), actorUserID, tournament)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

func (h *TournamentHandler) Delete(w http.ResponseWriter, r *http.Request) {
	actorUserID := mustUserID(r)
	if actorUserID == "" {
		writeError(w, apperror.Unauthorized("missing auth context"))
		return
	}
	id := chi.URLParam(r, "id")
	if err := h.deps.Tournaments.Delete(r.Context(), actorUserID, id); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "tournament deleted"})
}

func (h *TournamentHandler) ChangeStatus(w http.ResponseWriter, r *http.Request) {
	actorUserID := mustUserID(r)
	if actorUserID == "" {
		writeError(w, apperror.Unauthorized("missing auth context"))
		return
	}
	id := chi.URLParam(r, "id")
	var req changeStatusRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}
	if err := h.deps.Validate.Struct(req); err != nil {
		writeError(w, err)
		return
	}
	if err := h.deps.Tournaments.ChangeStatus(r.Context(), actorUserID, id, req.Status); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "status updated"})
}

func (h *TournamentHandler) AddManager(w http.ResponseWriter, r *http.Request) {
	actorUserID := mustUserID(r)
	if actorUserID == "" {
		writeError(w, apperror.Unauthorized("missing auth context"))
		return
	}
	id := chi.URLParam(r, "id")
	var req addManagerRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}
	if err := h.deps.Validate.Struct(req); err != nil {
		writeError(w, err)
		return
	}
	if err := h.deps.Tournaments.AddManager(r.Context(), actorUserID, id, req.UserID); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "manager added"})
}

func (h *TournamentHandler) RemoveManager(w http.ResponseWriter, r *http.Request) {
	actorUserID := mustUserID(r)
	if actorUserID == "" {
		writeError(w, apperror.Unauthorized("missing auth context"))
		return
	}
	id := chi.URLParam(r, "id")
	userID := chi.URLParam(r, "userId")
	if err := h.deps.Tournaments.RemoveManager(r.Context(), actorUserID, id, userID); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "manager removed"})
}

// ── Individual-mode participant management ────────────────────────────────────

func (h *TournamentHandler) GetParticipants(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	userID := mustUserID(r)
	data, err := h.deps.Challonge.GetBracket(r.Context(), id, userID)
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]interface{}{"items": []struct{}{}})
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"items": data.Participants})
}

func (h *TournamentHandler) AddParticipant(w http.ResponseWriter, r *http.Request) {
	actorID := mustUserID(r)
	id := chi.URLParam(r, "id")
	ok, err := h.deps.Tournaments.CanManageTournament(r.Context(), id, actorID)
	if err != nil || !ok {
		writeError(w, apperror.Forbidden("insufficient permissions"))
		return
	}
	var req struct {
		Name string `json:"name" validate:"required"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}
	if err := h.deps.Validate.Struct(req); err != nil {
		writeError(w, err)
		return
	}
	p, err := h.deps.Challonge.AddParticipant(r.Context(), id, actorID, service.AddParticipantReq{Name: req.Name})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, p)
}

func (h *TournamentHandler) BulkAddParticipants(w http.ResponseWriter, r *http.Request) {
	actorID := mustUserID(r)
	id := chi.URLParam(r, "id")
	ok, err := h.deps.Tournaments.CanManageTournament(r.Context(), id, actorID)
	if err != nil || !ok {
		writeError(w, apperror.Forbidden("insufficient permissions"))
		return
	}
	var req struct {
		Names []string `json:"names" validate:"required,min=1"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}
	if err := h.deps.Validate.Struct(req); err != nil {
		writeError(w, err)
		return
	}
	participants, err := h.deps.Challonge.BulkAddParticipants(r.Context(), id, actorID, req.Names)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]interface{}{"items": participants})
}

func (h *TournamentHandler) RemoveParticipant(w http.ResponseWriter, r *http.Request) {
	actorID := mustUserID(r)
	id := chi.URLParam(r, "id")
	participantID := chi.URLParam(r, "participantId")
	ok, err := h.deps.Tournaments.CanManageTournament(r.Context(), id, actorID)
	if err != nil || !ok {
		writeError(w, apperror.Forbidden("insufficient permissions"))
		return
	}
	if err := h.deps.Challonge.RemoveParticipant(r.Context(), id, participantID, actorID); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "participant removed"})
}

func (h *TournamentHandler) ShuffleParticipants(w http.ResponseWriter, r *http.Request) {
	actorID := mustUserID(r)
	id := chi.URLParam(r, "id")
	ok, err := h.deps.Tournaments.CanManageTournament(r.Context(), id, actorID)
	if err != nil || !ok {
		writeError(w, apperror.Forbidden("insufficient permissions"))
		return
	}
	if err := h.deps.Challonge.ShuffleParticipants(r.Context(), id, actorID); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "participants shuffled"})
}

func (h *TournamentHandler) StartIndividualBracket(w http.ResponseWriter, r *http.Request) {
	actorID := mustUserID(r)
	id := chi.URLParam(r, "id")
	ok, err := h.deps.Tournaments.CanManageTournament(r.Context(), id, actorID)
	if err != nil || !ok {
		writeError(w, apperror.Forbidden("insufficient permissions"))
		return
	}
	if err := h.deps.Challonge.Start(r.Context(), id, actorID); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "bracket started"})
}

func (h *TournamentHandler) RegisterTeam(w http.ResponseWriter, r *http.Request) {
	userID := mustUserID(r)
	id := chi.URLParam(r, "id")
	var req struct {
		TeamName string   `json:"team_name" validate:"required,min=2,max=100"`
		Members  []string `json:"members"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}
	if err := h.deps.Validate.Struct(req); err != nil {
		writeError(w, err)
		return
	}
	me, err := h.deps.Users.GetMe(r.Context(), userID)
	if err != nil {
		writeError(w, err)
		return
	}
	nickname := me.Nickname
	if nickname == "" {
		nickname = me.Email
	}
	result, err := h.deps.Teams.RegisterTeam(r.Context(), id, service.RegisterTeamInput{
		CaptainUserID:   userID,
		CaptainNickname: nickname,
		TeamName:        req.TeamName,
		MemberNicknames: req.Members,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, result)
}

func (h *TournamentHandler) JoinIndividual(w http.ResponseWriter, r *http.Request) {
	userID := mustUserID(r)
	id := chi.URLParam(r, "id")
	t, err := h.deps.Tournaments.GetTournament(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}
	if t.RegistrationMode == "team" {
		if t.Status != entity.TournamentStatusRegistrationOpen {
			writeError(w, apperror.BadRequest("not_open", "регистрация не открыта", nil))
			return
		}
		if err := h.deps.Challonge.JoinPool(r.Context(), id, userID); err != nil {
			writeError(w, err)
			return
		}
	} else {
		if err := h.deps.Challonge.Join(r.Context(), id, userID); err != nil {
			writeError(w, err)
			return
		}
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "joined"})
}

func pageParams(r *http.Request) (int, int) {
	limit := 20
	offset := 0
	if raw := r.URL.Query().Get("limit"); raw != "" {
		if v, err := strconv.Atoi(raw); err == nil && v > 0 && v <= 100 {
			limit = v
		}
	}
	if raw := r.URL.Query().Get("offset"); raw != "" {
		if v, err := strconv.Atoi(raw); err == nil && v >= 0 {
			offset = v
		}
	}
	return limit, offset
}

func parseOptionalTime(raw *string) (*time.Time, error) {
	if raw == nil || *raw == "" {
		return nil, nil
	}
	t, err := time.Parse(time.RFC3339, *raw)
	if err != nil {
		return nil, err
	}
	return &t, nil
}
