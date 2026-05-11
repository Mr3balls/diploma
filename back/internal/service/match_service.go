package service

import (
	"context"
	"fmt"
	"time"

	"esports-backend/internal/apperror"
	"esports-backend/internal/entity"
	"esports-backend/internal/pkg/xjson"
	"esports-backend/internal/repository"

	"github.com/google/uuid"
)

type MatchService struct {
	tournaments   *TournamentService
	brackets      *repository.BracketRepository
	teams         *repository.TeamRepository
	notifications *repository.NotificationRepository
	audits        *repository.AuditRepository
	bracketFlow   *BracketService
}

func NewMatchService(tournaments *TournamentService, brackets *repository.BracketRepository, teams *repository.TeamRepository, notifications *repository.NotificationRepository, audits *repository.AuditRepository, bracketFlow *BracketService) *MatchService {
	return &MatchService{tournaments: tournaments, brackets: brackets, teams: teams, notifications: notifications, audits: audits, bracketFlow: bracketFlow}
}

type ScheduleMatchInput struct {
	ScheduledAt      *time.Time
	LocationOrServer *string
}

type SubmitResultInput struct {
	WinnerTeamID string
	ScoreText    *string
	Comment      *string
}

func (s *MatchService) Schedule(ctx context.Context, actorUserID, matchID string, in ScheduleMatchInput) error {
	match, err := s.brackets.GetMatchByID(ctx, matchID)
	if err != nil {
		return apperror.NotFound("match not found")
	}
	ok, err := s.tournaments.CanManageTournament(ctx, match.TournamentID, actorUserID)
	if err != nil {
		return err
	}
	if !ok {
		return apperror.Forbidden("insufficient tournament permissions")
	}
	match.ScheduledAt = in.ScheduledAt
	match.LocationOrServer = in.LocationOrServer
	if err := s.brackets.UpdateMatchState(ctx, match); err != nil {
		return err
	}
	if err := s.notifyMatchTeams(ctx, match, entity.NotificationMatchTimeChanged, "Время матча обновлено", "Менеджер обновил данные матча"); err != nil {
		return err
	}
	return nil
}

func (s *MatchService) ConfirmReady(ctx context.Context, actorUserID, matchID string) error {
	return s.updateTeamConfirmation(ctx, actorUserID, matchID, entity.MatchTeamConfirmationReadyConfirmed, entity.MatchStatusConfirmed, entity.NotificationMatchAssigned, "Готовность подтверждена")
}

func (s *MatchService) RequestReschedule(ctx context.Context, actorUserID, matchID string) error {
	return s.updateTeamConfirmation(ctx, actorUserID, matchID, entity.MatchTeamConfirmationRescheduleRequested, entity.MatchStatusRescheduleRequested, entity.NotificationMatchRescheduled, "Запрошен перенос матча")
}

func (s *MatchService) ReportIssue(ctx context.Context, actorUserID, matchID string) error {
	return s.updateTeamConfirmation(ctx, actorUserID, matchID, entity.MatchTeamConfirmationIssueReported, entity.MatchStatusIssueReported, entity.NotificationMatchCancelled, "По матчу сообщено о проблеме")
}

func (s *MatchService) SubmitResult(ctx context.Context, actorUserID, matchID string, in SubmitResultInput) error {
	match, err := s.brackets.GetMatchByID(ctx, matchID)
	if err != nil {
		return apperror.NotFound("match not found")
	}
	isCaptain, err := s.userCaptainsTeamInMatch(ctx, actorUserID, match)
	if err != nil {
		return err
	}
	if !isCaptain {
		return apperror.Forbidden("only captain of participating team can submit result")
	}
	if (match.Team1ID == nil || *match.Team1ID != in.WinnerTeamID) && (match.Team2ID == nil || *match.Team2ID != in.WinnerTeamID) {
		return apperror.BadRequest("invalid_winner", "winner team must belong to the match", nil)
	}
	match.WinnerTeamID = &in.WinnerTeamID
	match.ScoreText = in.ScoreText
	match.ManagerComment = in.Comment
	match.Status = entity.MatchStatusAwaitingConfirmation
	if err := s.brackets.UpdateMatchState(ctx, match); err != nil {
		return err
	}
	_ = s.audits.Create(ctx, &entity.AuditLog{ID: uuid.NewString(), ActorUserID: &actorUserID, TournamentID: &match.TournamentID, EntityType: "match", EntityID: match.ID, ActionType: "result_submitted", Description: "Match result submitted", MetadataJSON: xjson.MustMarshal(in)})
	return s.notifyManagers(ctx, match.TournamentID, entity.NotificationResultSubmitted, "Результат матча отправлен", fmt.Sprintf("По матчу %s отправлен результат на подтверждение", match.ID))
}

type AdminSetResultInput struct {
	WinnerTeamID string
	ScoreText    *string
}

func (s *MatchService) AdminSetResult(ctx context.Context, actorUserID, matchID string, in AdminSetResultInput) error {
	match, err := s.brackets.GetMatchByID(ctx, matchID)
	if err != nil {
		return apperror.NotFound("match not found")
	}
	ok, err := s.tournaments.CanManageTournament(ctx, match.TournamentID, actorUserID)
	if err != nil {
		return err
	}
	if !ok {
		return apperror.Forbidden("insufficient tournament permissions")
	}
	if (match.Team1ID == nil || *match.Team1ID != in.WinnerTeamID) && (match.Team2ID == nil || *match.Team2ID != in.WinnerTeamID) {
		return apperror.BadRequest("invalid_winner", "winner team must belong to the match", nil)
	}
	winnerID := in.WinnerTeamID
	match.WinnerTeamID = &winnerID
	match.ScoreText = in.ScoreText
	match.Status = entity.MatchStatusFinished
	if err := s.brackets.UpdateMatchState(ctx, match); err != nil {
		return err
	}
	if err := s.bracketFlow.PropagateWinner(ctx, actorUserID, match.ID); err != nil {
		return err
	}
	_ = s.audits.Create(ctx, &entity.AuditLog{ID: uuid.NewString(), ActorUserID: &actorUserID, TournamentID: &match.TournamentID, EntityType: "match", EntityID: match.ID, ActionType: "admin_set_result", Description: "Admin directly set match result", MetadataJSON: xjson.MustMarshal(map[string]string{"winner_team_id": in.WinnerTeamID})})
	return s.notifyMatchTeams(ctx, match, entity.NotificationResultConfirmed, "Победитель матча установлен", "Менеджер установил победителя матча")
}

func (s *MatchService) ApproveResult(ctx context.Context, actorUserID, matchID string) error {
	match, err := s.brackets.GetMatchByID(ctx, matchID)
	if err != nil {
		return apperror.NotFound("match not found")
	}
	ok, err := s.tournaments.CanManageTournament(ctx, match.TournamentID, actorUserID)
	if err != nil {
		return err
	}
	if !ok {
		return apperror.Forbidden("insufficient tournament permissions")
	}
	if match.WinnerTeamID == nil {
		return apperror.BadRequest("winner_not_set", "winner team must be submitted before approval", nil)
	}
	match.Status = entity.MatchStatusFinished
	if err := s.brackets.UpdateMatchState(ctx, match); err != nil {
		return err
	}
	if err := s.bracketFlow.PropagateWinner(ctx, actorUserID, match.ID); err != nil {
		return err
	}
	_ = s.audits.Create(ctx, &entity.AuditLog{ID: uuid.NewString(), ActorUserID: &actorUserID, TournamentID: &match.TournamentID, EntityType: "match", EntityID: match.ID, ActionType: "result_approved", Description: "Match result approved", MetadataJSON: xjson.MustMarshal(map[string]string{"winner_team_id": *match.WinnerTeamID})})
	return s.notifyMatchTeams(ctx, match, entity.NotificationResultConfirmed, "Результат матча подтвержден", "Менеджер подтвердил результат матча")
}

func (s *MatchService) RejectResult(ctx context.Context, actorUserID, matchID string) error {
	match, err := s.brackets.GetMatchByID(ctx, matchID)
	if err != nil {
		return apperror.NotFound("match not found")
	}
	ok, err := s.tournaments.CanManageTournament(ctx, match.TournamentID, actorUserID)
	if err != nil {
		return err
	}
	if !ok {
		return apperror.Forbidden("insufficient tournament permissions")
	}
	match.WinnerTeamID = nil
	match.ScoreText = nil
	match.ManagerComment = nil
	match.Status = entity.MatchStatusConfirmed
	return s.brackets.UpdateMatchState(ctx, match)
}

func (s *MatchService) updateTeamConfirmation(ctx context.Context, actorUserID, matchID, teamStatus, matchStatus, notificationType, message string) error {
	match, err := s.brackets.GetMatchByID(ctx, matchID)
	if err != nil {
		return apperror.NotFound("match not found")
	}
	teamID, err := s.findCaptainsTeamInMatch(ctx, actorUserID, match)
	if err != nil {
		return err
	}
	if teamID == "" {
		return apperror.Forbidden("only captain of participating team may update readiness")
	}
	if match.Team1ID != nil && *match.Team1ID == teamID {
		match.Team1ConfirmationStatus = teamStatus
	}
	if match.Team2ID != nil && *match.Team2ID == teamID {
		match.Team2ConfirmationStatus = teamStatus
	}
	match.Status = matchStatus
	if match.Team1ConfirmationStatus == entity.MatchTeamConfirmationReadyConfirmed && match.Team2ConfirmationStatus == entity.MatchTeamConfirmationReadyConfirmed {
		match.Status = entity.MatchStatusConfirmed
	}
	if err := s.brackets.UpdateMatchState(ctx, match); err != nil {
		return err
	}
	_ = s.audits.Create(ctx, &entity.AuditLog{ID: uuid.NewString(), ActorUserID: &actorUserID, TournamentID: &match.TournamentID, EntityType: "match", EntityID: match.ID, ActionType: "match_confirmation_updated", Description: message, MetadataJSON: xjson.MustMarshal(map[string]string{"team_status": teamStatus})})
	return s.notifyManagers(ctx, match.TournamentID, notificationType, message, message)
}

func (s *MatchService) findCaptainsTeamInMatch(ctx context.Context, userID string, match *entity.Match) (string, error) {
	captains, err := s.teams.FindCaptainMembership(ctx, userID, match.TournamentID)
	if err != nil {
		return "", err
	}
	for _, captain := range captains {
		if (match.Team1ID != nil && captain.TeamID == *match.Team1ID) || (match.Team2ID != nil && captain.TeamID == *match.Team2ID) {
			return captain.TeamID, nil
		}
	}
	return "", nil
}

func (s *MatchService) userCaptainsTeamInMatch(ctx context.Context, userID string, match *entity.Match) (bool, error) {
	teamID, err := s.findCaptainsTeamInMatch(ctx, userID, match)
	if err != nil {
		return false, err
	}
	return teamID != "", nil
}

func (s *MatchService) notifyManagers(ctx context.Context, tournamentID, typ, title, message string) error {
	userIDs, err := s.tournaments.tournaments.ListUserIDsByRoles(ctx, tournamentID, []string{entity.TournamentRoleOwner, entity.TournamentRoleManager})
	if err != nil {
		return err
	}
	for _, userID := range userIDs {
		notification := &entity.Notification{
			ID:          uuid.NewString(),
			UserID:      userID,
			Type:        typ,
			Title:       title,
			Message:     message,
			PayloadJSON: xjson.MustMarshal(map[string]string{"tournament_id": tournamentID}),
		}
		if err := s.notifications.Create(ctx, notification); err != nil {
			return err
		}
	}
	return nil
}

func (s *MatchService) notifyMatchTeams(ctx context.Context, match *entity.Match, typ, title, message string) error {
	for _, teamID := range []*string{match.Team1ID, match.Team2ID} {
		if teamID == nil {
			continue
		}
		members, err := s.teams.ListMembersByTeamID(ctx, *teamID)
		if err != nil {
			return err
		}
		for _, member := range members {
			if member.UserID == nil {
				continue
			}
			notification := &entity.Notification{ID: uuid.NewString(), UserID: *member.UserID, Type: typ, Title: title, Message: message, PayloadJSON: xjson.MustMarshal(map[string]string{"match_id": match.ID, "tournament_id": match.TournamentID})}
			if err := s.notifications.Create(ctx, notification); err != nil {
				return err
			}
		}
	}
	return nil
}
