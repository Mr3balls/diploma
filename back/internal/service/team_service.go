package service

import (
	"context"
	"fmt"
	"strings"

	"esports-backend/internal/apperror"
	"esports-backend/internal/entity"
	"esports-backend/internal/pkg/xjson"
	"esports-backend/internal/repository"

	"github.com/google/uuid"
)

type TeamService struct {
	tournaments   *TournamentService
	teams         *repository.TeamRepository
	users         *repository.UserRepository
	notifications *repository.NotificationRepository
	audits        *repository.AuditRepository
}

func NewTeamService(tournaments *TournamentService, teams *repository.TeamRepository, users *repository.UserRepository, notifications *repository.NotificationRepository, audits *repository.AuditRepository) *TeamService {
	return &TeamService{tournaments: tournaments, teams: teams, users: users, notifications: notifications, audits: audits}
}

type TeamDetails struct {
	Team    *entity.Team        `json:"team"`
	Members []entity.TeamMember `json:"members"`
}

type ReplaceMemberInput struct {
	Nickname string
}

func (s *TeamService) GetTeam(ctx context.Context, teamID string) (*TeamDetails, error) {
	team, err := s.teams.GetTeamByID(ctx, teamID)
	if err != nil {
		return nil, apperror.NotFound("team not found")
	}
	members, err := s.teams.ListMembersByTeamID(ctx, teamID)
	if err != nil {
		return nil, err
	}
	return &TeamDetails{Team: team, Members: members}, nil
}

func (s *TeamService) UpdateTeam(ctx context.Context, actorUserID, teamID, name string) (*TeamDetails, error) {
	team, err := s.teams.GetTeamByID(ctx, teamID)
	if err != nil {
		return nil, apperror.NotFound("team not found")
	}
	ok, err := s.tournaments.CanManageTournament(ctx, team.TournamentID, actorUserID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, apperror.Forbidden("insufficient tournament permissions")
	}
	team.Name = name
	if err := s.teams.UpdateTeam(ctx, team); err != nil {
		return nil, err
	}
	return s.GetTeam(ctx, teamID)
}

func (s *TeamService) ApproveTeam(ctx context.Context, actorUserID, teamID string) error {
	details, err := s.GetTeam(ctx, teamID)
	if err != nil {
		return err
	}
	ok, err := s.tournaments.CanManageTournament(ctx, details.Team.TournamentID, actorUserID)
	if err != nil {
		return err
	}
	if !ok {
		return apperror.Forbidden("insufficient tournament permissions")
	}
	if !IsTeamReadyForReview(details.Members) {
		return apperror.BadRequest("team_not_ready", "captain must confirm and at least 4 players must confirm", nil)
	}
	approved := true
	if err := s.teams.SetApproval(ctx, teamID, entity.TeamStatusApproved, &approved); err != nil {
		return err
	}
	_ = s.audits.Create(ctx, &entity.AuditLog{ID: uuid.NewString(), ActorUserID: &actorUserID, TournamentID: &details.Team.TournamentID, EntityType: "team", EntityID: teamID, ActionType: "team_approved", Description: "Team approved by manager", MetadataJSON: xjson.MustMarshal(map[string]string{"team_id": teamID})})
	return nil
}

func (s *TeamService) RejectTeam(ctx context.Context, actorUserID, teamID string, reason string) error {
	details, err := s.GetTeam(ctx, teamID)
	if err != nil {
		return err
	}
	ok, err := s.tournaments.CanManageTournament(ctx, details.Team.TournamentID, actorUserID)
	if err != nil {
		return err
	}
	if !ok {
		return apperror.Forbidden("insufficient tournament permissions")
	}
	approved := false
	if err := s.teams.SetApproval(ctx, teamID, entity.TeamStatusRejected, &approved); err != nil {
		return err
	}
	_ = s.audits.Create(ctx, &entity.AuditLog{ID: uuid.NewString(), ActorUserID: &actorUserID, TournamentID: &details.Team.TournamentID, EntityType: "team", EntityID: teamID, ActionType: "team_rejected", Description: "Team rejected by manager", MetadataJSON: xjson.MustMarshal(map[string]string{"reason": reason})})
	return nil
}

func (s *TeamService) RemoveMember(ctx context.Context, actorUserID, teamID, memberID string) error {
	details, err := s.GetTeam(ctx, teamID)
	if err != nil {
		return err
	}
	ok, err := s.tournaments.CanManageTournament(ctx, details.Team.TournamentID, actorUserID)
	if err != nil {
		return err
	}
	if !ok {
		return apperror.Forbidden("insufficient tournament permissions")
	}
	return s.teams.RemoveMember(ctx, memberID)
}

func (s *TeamService) ReplaceMember(ctx context.Context, actorUserID, teamID, memberID string, in ReplaceMemberInput) error {
	details, err := s.GetTeam(ctx, teamID)
	if err != nil {
		return err
	}
	ok, err := s.tournaments.CanManageTournament(ctx, details.Team.TournamentID, actorUserID)
	if err != nil {
		return err
	}
	if !ok {
		return apperror.Forbidden("insufficient tournament permissions")
	}

	var userID *string
	if matchedUser, err := s.users.GetByNickname(ctx, in.Nickname); err == nil {
		userID = &matchedUser.ID
	}
	if err := s.teams.ReplaceMember(ctx, memberID, userID, in.Nickname); err != nil {
		return err
	}
	if userID != nil {
		payload := map[string]string{"team_id": teamID, "team_member_id": memberID, "tournament_id": details.Team.TournamentID}
		_ = s.notifications.Create(ctx, &entity.Notification{ID: uuid.NewString(), UserID: *userID, Type: entity.NotificationAddedToTeam, Title: "Вас добавили в команду", Message: fmt.Sprintf("Подтвердите участие в команде %s", details.Team.Name), PayloadJSON: xjson.MustMarshal(payload), ActionPayloadJSON: xjson.MustMarshal(payload)})
	}
	return nil
}

func (s *TeamService) AcceptMembership(ctx context.Context, actorUserID, memberID string) error {
	member, err := s.teams.GetMemberByID(ctx, memberID)
	if err != nil {
		return apperror.NotFound("team member not found")
	}
	if member.UserID == nil || *member.UserID != actorUserID {
		return apperror.Forbidden("membership does not belong to current user")
	}
	if err := s.teams.SetMemberConfirmation(ctx, memberID, entity.MemberConfirmationConfirmed); err != nil {
		return err
	}
	team, err := s.teams.GetTeamByID(ctx, member.TeamID)
	if err == nil {
		members, _ := s.teams.ListMembersByTeamID(ctx, member.TeamID)
		if IsTeamReadyForReview(members) {
			_ = s.teams.SetApproval(ctx, team.ID, entity.TeamStatusReadyForReview, team.ApprovedByManager)
		}
		_ = s.audits.Create(ctx, &entity.AuditLog{ID: uuid.NewString(), ActorUserID: &actorUserID, TournamentID: &team.TournamentID, EntityType: "team_member", EntityID: memberID, ActionType: "team_member_accepted", Description: "Team participation confirmed", MetadataJSON: xjson.MustMarshal(map[string]string{"member_id": memberID})})
	}
	return nil
}

func (s *TeamService) DeclineMembership(ctx context.Context, actorUserID, memberID string) error {
	member, err := s.teams.GetMemberByID(ctx, memberID)
	if err != nil {
		return apperror.NotFound("team member not found")
	}
	if member.UserID == nil || *member.UserID != actorUserID {
		return apperror.Forbidden("membership does not belong to current user")
	}
	if err := s.teams.SetMemberConfirmation(ctx, memberID, entity.MemberConfirmationDeclined); err != nil {
		return err
	}
	team, err := s.teams.GetTeamByID(ctx, member.TeamID)
	if err == nil {
		_ = s.teams.SetApproval(ctx, team.ID, entity.TeamStatusAwaitingConfirmation, team.ApprovedByManager)
		_ = s.audits.Create(ctx, &entity.AuditLog{ID: uuid.NewString(), ActorUserID: &actorUserID, TournamentID: &team.TournamentID, EntityType: "team_member", EntityID: memberID, ActionType: "team_member_declined", Description: "Team participation declined", MetadataJSON: xjson.MustMarshal(map[string]string{"member_id": memberID})})
	}
	return nil
}

func IsTeamReadyForReview(members []entity.TeamMember) bool {
	captainConfirmed := false
	confirmedMainPlayers := 0
	for _, m := range members {
		if m.IsSubstitute || strings.EqualFold(m.MemberRole, entity.MemberRoleSubstitute) {
			continue
		}
		if m.ConfirmationStatus == entity.MemberConfirmationConfirmed {
			confirmedMainPlayers++
		}
		if m.IsCaptain && m.ConfirmationStatus == entity.MemberConfirmationConfirmed {
			captainConfirmed = true
		}
	}
	return captainConfirmed && confirmedMainPlayers >= 4
}
