package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"esports-backend/internal/apperror"
	"esports-backend/internal/entity"
	"esports-backend/internal/pkg/notif"
	"esports-backend/internal/pkg/xjson"
	"esports-backend/internal/repository"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

type TeamService struct {
	tournaments   *TournamentService
	teams         *repository.TeamRepository
	users         repository.UserStore
	notifications *repository.NotificationRepository
	audits        *repository.AuditRepository
	email         *EmailService
}

func NewTeamService(tournaments *TournamentService, teams *repository.TeamRepository, users repository.UserStore, notifications *repository.NotificationRepository, audits *repository.AuditRepository, email *EmailService) *TeamService {
	return &TeamService{tournaments: tournaments, teams: teams, users: users, notifications: notifications, audits: audits, email: email}
}

type TeamDetails struct {
	Team    *entity.Team        `json:"team"`
	Members []entity.TeamMember `json:"members"`
}

type ReplaceMemberInput struct {
	Email string
}

func emailPrefix(email string) string {
	if idx := strings.Index(email, "@"); idx > 0 {
		return email[:idx]
	}
	return email
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if e, ok := err.(*pgconn.PgError); ok {
		pgErr = e
		return pgErr.Code == "23505"
	}
	return strings.Contains(err.Error(), "23505")
}

type RegisterTeamInput struct {
	CaptainUserID   string
	CaptainNickname string
	TeamName        string
	MemberEmails    []string
}

func (s *TeamService) RegisterTeam(ctx context.Context, tournamentID string, in RegisterTeamInput) (*TeamDetails, error) {
	tournament, err := s.tournaments.GetTournament(ctx, tournamentID)
	if err != nil {
		return nil, err
	}
	if tournament.Status != entity.TournamentStatusRegistrationOpen {
		return nil, apperror.BadRequest("not_open", "tournament is not accepting registrations", nil)
	}
	if tournament.RegistrationMode != "team" {
		return nil, apperror.BadRequest("wrong_mode", "tournament is not team-based", nil)
	}

	if tournament.MaxTeams > 0 {
		count, err := s.teams.CountByTournament(ctx, tournamentID)
		if err != nil {
			return nil, err
		}
		if count >= tournament.MaxTeams {
			return nil, apperror.BadRequest("max_teams_reached",
				fmt.Sprintf("достигнут лимит команд (%d)", tournament.MaxTeams), nil)
		}
	}

	existing, _ := s.teams.FindCaptainMembership(ctx, in.CaptainUserID, tournamentID)
	if len(existing) > 0 {
		return nil, apperror.BadRequest("already_registered", "you already have a team in this tournament", nil)
	}

	team := &entity.Team{
		ID:           uuid.NewString(),
		TournamentID: tournamentID,
		Name:         in.TeamName,
		Status:       entity.TeamStatusPending,
	}
	if err := s.teams.CreateTeam(ctx, team); err != nil {
		if isUniqueViolation(err) {
			return nil, apperror.BadRequest("name_taken", "команда с таким названием уже зарегистрирована", nil)
		}
		return nil, err
	}

	now := time.Now()
	captainMember := &entity.TeamMember{
		ID:                 uuid.NewString(),
		TeamID:             team.ID,
		UserID:             &in.CaptainUserID,
		Nickname:           in.CaptainNickname,
		MemberRole:         entity.MemberRolePlayer,
		IsCaptain:          true,
		ConfirmationStatus: entity.MemberConfirmationConfirmed,
		InvitedAt:          &now,
		RespondedAt:        &now,
	}
	if err := s.teams.CreateMember(ctx, captainMember); err != nil {
		return nil, err
	}

	for _, memberEmail := range in.MemberEmails {
		if memberEmail == "" {
			continue
		}
		var userID *string
		var userLang string
		nickname := emailPrefix(memberEmail)
		storedEmail := memberEmail
		if u, err := s.users.GetByEmail(ctx, memberEmail); err == nil {
			userID = &u.ID
			userLang = u.Lang
			if u.Nickname != "" {
				nickname = u.Nickname
			} else if u.FirstName != "" {
				nickname = u.FirstName
			}
		}
		member := &entity.TeamMember{
			ID:                 uuid.NewString(),
			TeamID:             team.ID,
			UserID:             userID,
			Nickname:           nickname,
			Email:              &storedEmail,
			MemberRole:         entity.MemberRolePlayer,
			IsCaptain:          false,
			ConfirmationStatus: entity.MemberConfirmationPendingConfirmation,
			InvitedAt:          &now,
		}
		if err := s.teams.CreateMember(ctx, member); err != nil {
			return nil, err
		}
		if tournament != nil {
			go s.email.SendTeamInvite(memberEmail, team.Name, tournament.Title)
		}
		if userID != nil {
			payload := map[string]string{"team_id": team.ID, "team_member_id": member.ID, "tournament_id": tournamentID}
			texts := notif.AddedToTeam(userLang, team.Name)
			_ = s.notifications.Create(ctx, &entity.Notification{
				ID: uuid.NewString(), UserID: *userID,
				Type:              entity.NotificationAddedToTeam,
				Title:             texts.Title,
				Message:           texts.Message,
				PayloadJSON:       xjson.MustMarshal(payload),
				ActionPayloadJSON: xjson.MustMarshal(payload),
			})
		}
	}

	return s.GetTeam(ctx, team.ID)
}

type AdminCreateTeamInput struct {
	AdminUserID string
	TeamName    string
	Members     []string // first member is captain
}

func (s *TeamService) AdminCreateTeam(ctx context.Context, tournamentID string, in AdminCreateTeamInput) (*TeamDetails, error) {
	ok, err := s.tournaments.CanManageTournament(ctx, tournamentID, in.AdminUserID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, apperror.Forbidden("insufficient tournament permissions")
	}

	tournament, err := s.tournaments.GetTournament(ctx, tournamentID)
	if err != nil {
		return nil, err
	}
	if tournament.MaxTeams > 0 {
		count, err := s.teams.CountByTournament(ctx, tournamentID)
		if err != nil {
			return nil, err
		}
		if count >= tournament.MaxTeams {
			return nil, apperror.BadRequest("max_teams_reached",
				fmt.Sprintf("достигнут лимит команд (%d)", tournament.MaxTeams), nil)
		}
	}

	team := &entity.Team{
		ID:           uuid.NewString(),
		TournamentID: tournamentID,
		Name:         in.TeamName,
		Status:       entity.TeamStatusPending,
	}
	if err := s.teams.CreateTeam(ctx, team); err != nil {
		if isUniqueViolation(err) {
			return nil, apperror.BadRequest("name_taken", "команда с таким названием уже существует", nil)
		}
		return nil, err
	}

	now := time.Now()
	for i, memberEmail := range in.Members {
		if memberEmail == "" {
			continue
		}
		var userID *string
		var userLang string
		nickname := emailPrefix(memberEmail)
		storedEmail := memberEmail
		if u, err := s.users.GetByEmail(ctx, memberEmail); err == nil {
			userID = &u.ID
			userLang = u.Lang
			if u.Nickname != "" {
				nickname = u.Nickname
			} else if u.FirstName != "" {
				nickname = u.FirstName
			}
		}
		isCaptain := i == 0
		confirmStatus := entity.MemberConfirmationPendingConfirmation
		if isCaptain {
			confirmStatus = entity.MemberConfirmationConfirmed
		}
		member := &entity.TeamMember{
			ID:                 uuid.NewString(),
			TeamID:             team.ID,
			UserID:             userID,
			Nickname:           nickname,
			Email:              &storedEmail,
			MemberRole:         entity.MemberRolePlayer,
			IsCaptain:          isCaptain,
			ConfirmationStatus: confirmStatus,
			InvitedAt:          &now,
		}
		if isCaptain {
			member.RespondedAt = &now
		}
		if err := s.teams.CreateMember(ctx, member); err != nil {
			return nil, err
		}
		if !isCaptain {
			// Always send an email invite so the person knows they've been added
			// (works even if they don't have an account yet).
			go s.email.SendTeamInvite(memberEmail, team.Name, tournament.Title)

			// If the user has an account, also send an in-app notification.
			if userID != nil {
				payload := map[string]string{"team_id": team.ID, "team_member_id": member.ID, "tournament_id": tournamentID}
				texts := notif.AddedToTeam(userLang, in.TeamName)
				_ = s.notifications.Create(ctx, &entity.Notification{
					ID: uuid.NewString(), UserID: *userID,
					Type:              entity.NotificationAddedToTeam,
					Title:             texts.Title,
					Message:           texts.Message,
					PayloadJSON:       xjson.MustMarshal(payload),
					ActionPayloadJSON: xjson.MustMarshal(payload),
				})
			}
		}
	}

	return s.GetTeam(ctx, team.ID)
}

func (s *TeamService) AdminDeleteTeam(ctx context.Context, actorUserID, teamID string) error {
	team, err := s.teams.GetTeamByID(ctx, teamID)
	if err != nil {
		return apperror.NotFound("team not found")
	}
	ok, err := s.tournaments.CanManageTournament(ctx, team.TournamentID, actorUserID)
	if err != nil {
		return err
	}
	if !ok {
		return apperror.Forbidden("insufficient tournament permissions")
	}
	if err := s.teams.DeleteTeam(ctx, teamID); err != nil {
		return err
	}
	_ = s.audits.Create(ctx, &entity.AuditLog{ID: uuid.NewString(), ActorUserID: &actorUserID, TournamentID: &team.TournamentID, EntityType: "team", EntityID: teamID, ActionType: "team_deleted", Description: "Team deleted by admin", MetadataJSON: xjson.MustMarshal(map[string]string{"team_id": teamID})})
	return nil
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

func (s *TeamService) GetMyTeam(ctx context.Context, userID, tournamentID string) (*TeamDetails, error) {
	memberships, err := s.teams.FindCaptainMembership(ctx, userID, tournamentID)
	if err != nil || len(memberships) == 0 {
		return nil, apperror.NotFound("no team found for this user in this tournament")
	}
	return s.GetTeam(ctx, memberships[0].TeamID)
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
	approved := true
	if err := s.teams.SetApproval(ctx, teamID, entity.TeamStatusApproved, &approved); err != nil {
		return err
	}
	_ = s.audits.Create(ctx, &entity.AuditLog{ID: uuid.NewString(), ActorUserID: &actorUserID, TournamentID: &details.Team.TournamentID, EntityType: "team", EntityID: teamID, ActionType: "team_approved", Description: "Team approved by manager", MetadataJSON: xjson.MustMarshal(map[string]string{"team_id": teamID})})
	payload := map[string]string{"team_id": teamID, "tournament_id": details.Team.TournamentID}
	for _, m := range details.Members {
		if m.UserID != nil {
			texts := notif.TeamApproved(m.UserLang, details.Team.Name)
			_ = s.notifications.Create(ctx, &entity.Notification{
				ID: uuid.NewString(), UserID: *m.UserID,
				Type:              entity.NotificationTeamParticipationConfirm,
				Title:             texts.Title,
				Message:           texts.Message,
				PayloadJSON:       xjson.MustMarshal(payload),
				ActionPayloadJSON: xjson.MustMarshal(payload),
			})
		}
	}
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
	payload := map[string]string{"team_id": teamID, "tournament_id": details.Team.TournamentID}
	for _, m := range details.Members {
		if m.UserID != nil {
			texts := notif.TeamDeclined(m.UserLang, details.Team.Name, reason)
			_ = s.notifications.Create(ctx, &entity.Notification{
				ID: uuid.NewString(), UserID: *m.UserID,
				Type:              entity.NotificationTeamParticipationDecline,
				Title:             texts.Title,
				Message:           texts.Message,
				PayloadJSON:       xjson.MustMarshal(payload),
				ActionPayloadJSON: xjson.MustMarshal(payload),
			})
		}
	}
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
	isAdmin, _ := s.tournaments.CanManageTournament(ctx, details.Team.TournamentID, actorUserID)
	isCaptain := false
	for _, m := range details.Members {
		if m.IsCaptain && m.UserID != nil && *m.UserID == actorUserID {
			isCaptain = true
			break
		}
	}
	if !isAdmin && !isCaptain {
		return apperror.Forbidden("insufficient tournament permissions")
	}

	var userID *string
	var userLang string
	nickname := emailPrefix(in.Email)
	if matchedUser, err := s.users.GetByEmail(ctx, in.Email); err == nil {
		userID = &matchedUser.ID
		userLang = matchedUser.Lang
		if matchedUser.Nickname != "" {
			nickname = matchedUser.Nickname
		} else if matchedUser.FirstName != "" {
			nickname = matchedUser.FirstName
		}
	}
	storedEmail := in.Email
	if err := s.teams.ReplaceMember(ctx, memberID, userID, nickname, &storedEmail); err != nil {
		return err
	}
	if userID != nil {
		payload := map[string]string{"team_id": teamID, "team_member_id": memberID, "tournament_id": details.Team.TournamentID}
		texts := notif.AddedToTeam(userLang, details.Team.Name)
		_ = s.notifications.Create(ctx, &entity.Notification{
			ID: uuid.NewString(), UserID: *userID,
			Type:              entity.NotificationAddedToTeam,
			Title:             texts.Title,
			Message:           texts.Message,
			PayloadJSON:       xjson.MustMarshal(payload),
			ActionPayloadJSON: xjson.MustMarshal(payload),
		})
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
		if user, err := s.users.GetByID(ctx, actorUserID); err == nil {
			tournament, _ := s.tournaments.GetTournament(ctx, team.TournamentID)
			title := ""
			if tournament != nil {
				title = tournament.Title
			}
			go s.email.SendTeamParticipationConfirmed(user.Email, team.Name, title)
		}
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
		if user, err := s.users.GetByID(ctx, actorUserID); err == nil {
			tournament, _ := s.tournaments.GetTournament(ctx, team.TournamentID)
			title := ""
			if tournament != nil {
				title = tournament.Title
			}
			go s.email.SendTeamParticipationDeclined(user.Email, team.Name, title)
		}
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
