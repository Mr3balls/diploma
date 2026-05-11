package handler

import (
	"net/http"

	"esports-backend/internal/apperror"
	"esports-backend/internal/service"

	"github.com/go-chi/chi/v5"
)

type TeamHandler struct{ deps Deps }

func NewTeamHandler(deps Deps) *TeamHandler { return &TeamHandler{deps: deps} }

type patchTeamRequest struct {
	Name string `json:"name" validate:"required,min=2,max=100"`
}

type rejectTeamRequest struct {
	Reason string `json:"reason" validate:"required,min=2,max=300"`
}

type replaceMemberRequest struct {
	Nickname string `json:"nickname" validate:"required,min=2,max=50"`
}

func (h *TeamHandler) AdminCreateTeam(w http.ResponseWriter, r *http.Request) {
	actorUserID := mustUserID(r)
	if actorUserID == "" {
		writeError(w, apperror.Unauthorized("missing auth context"))
		return
	}
	tournamentID := chi.URLParam(r, "id")
	var req struct {
		TeamName string   `json:"team_name" validate:"required,min=2,max=100"`
		Members  []string `json:"members" validate:"required,min=1"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}
	if err := h.deps.Validate.Struct(req); err != nil {
		writeError(w, err)
		return
	}
	result, err := h.deps.Teams.AdminCreateTeam(r.Context(), tournamentID, service.AdminCreateTeamInput{
		AdminUserID: actorUserID,
		TeamName:    req.TeamName,
		Members:     req.Members,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, result)
}

func (h *TeamHandler) GetAdminTeams(w http.ResponseWriter, r *http.Request) {
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
	items, err := h.deps.Tournaments.ListTournamentTeams(r.Context(), tournamentID, true)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"items": items})
}

func (h *TeamHandler) GetTeam(w http.ResponseWriter, r *http.Request) {
	teamID := chi.URLParam(r, "id")
	result, err := h.deps.Teams.GetTeam(r.Context(), teamID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (h *TeamHandler) PatchTeam(w http.ResponseWriter, r *http.Request) {
	actorUserID := mustUserID(r)
	if actorUserID == "" {
		writeError(w, apperror.Unauthorized("missing auth context"))
		return
	}
	teamID := chi.URLParam(r, "id")
	var req patchTeamRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}
	result, err := h.deps.Teams.UpdateTeam(r.Context(), actorUserID, teamID, req.Name)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (h *TeamHandler) ApproveTeam(w http.ResponseWriter, r *http.Request) {
	actorUserID := mustUserID(r)
	if actorUserID == "" {
		writeError(w, apperror.Unauthorized("missing auth context"))
		return
	}
	teamID := chi.URLParam(r, "id")
	if err := h.deps.Teams.ApproveTeam(r.Context(), actorUserID, teamID); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "team approved"})
}

func (h *TeamHandler) RejectTeam(w http.ResponseWriter, r *http.Request) {
	actorUserID := mustUserID(r)
	if actorUserID == "" {
		writeError(w, apperror.Unauthorized("missing auth context"))
		return
	}
	teamID := chi.URLParam(r, "id")
	var req rejectTeamRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}
	if err := h.deps.Teams.RejectTeam(r.Context(), actorUserID, teamID, req.Reason); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "team rejected"})
}

func (h *TeamHandler) RemoveMember(w http.ResponseWriter, r *http.Request) {
	actorUserID := mustUserID(r)
	if actorUserID == "" {
		writeError(w, apperror.Unauthorized("missing auth context"))
		return
	}
	teamID := chi.URLParam(r, "id")
	memberID := chi.URLParam(r, "memberId")
	if err := h.deps.Teams.RemoveMember(r.Context(), actorUserID, teamID, memberID); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "member removed"})
}

func (h *TeamHandler) ReplaceMember(w http.ResponseWriter, r *http.Request) {
	actorUserID := mustUserID(r)
	if actorUserID == "" {
		writeError(w, apperror.Unauthorized("missing auth context"))
		return
	}
	teamID := chi.URLParam(r, "id")
	memberID := chi.URLParam(r, "memberId")
	var req replaceMemberRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}
	if err := h.deps.Teams.ReplaceMember(r.Context(), actorUserID, teamID, memberID, service.ReplaceMemberInput{Nickname: req.Nickname}); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "member replaced"})
}

func (h *TeamHandler) AcceptMembership(w http.ResponseWriter, r *http.Request) {
	actorUserID := mustUserID(r)
	if actorUserID == "" {
		writeError(w, apperror.Unauthorized("missing auth context"))
		return
	}
	memberID := chi.URLParam(r, "id")
	if err := h.deps.Teams.AcceptMembership(r.Context(), actorUserID, memberID); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "participation confirmed"})
}

func (h *TeamHandler) DeclineMembership(w http.ResponseWriter, r *http.Request) {
	actorUserID := mustUserID(r)
	if actorUserID == "" {
		writeError(w, apperror.Unauthorized("missing auth context"))
		return
	}
	memberID := chi.URLParam(r, "id")
	if err := h.deps.Teams.DeclineMembership(r.Context(), actorUserID, memberID); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "participation declined"})
}
