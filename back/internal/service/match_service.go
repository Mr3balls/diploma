package service

import (
	"context"
	"time"

	"esports-backend/internal/apperror"
	"esports-backend/internal/entity"
	"esports-backend/internal/pkg/notif"
	"esports-backend/internal/pkg/xjson"
	"esports-backend/internal/repository"

	"github.com/google/uuid"
)

type MatchService struct {
	tournaments   *TournamentService
	brackets      *repository.BracketRepository
	groups        *repository.GroupRepository
	teams         *repository.TeamRepository
	users         repository.UserStore
	notifications *repository.NotificationRepository
	audits        *repository.AuditRepository
	bracketFlow   *BracketService
	email         *EmailService
}

func NewMatchService(tournaments *TournamentService, brackets *repository.BracketRepository, groups *repository.GroupRepository, teams *repository.TeamRepository, users repository.UserStore, notifications *repository.NotificationRepository, audits *repository.AuditRepository, bracketFlow *BracketService, email *EmailService) *MatchService {
	return &MatchService{tournaments: tournaments, brackets: brackets, groups: groups, teams: teams, users: users, notifications: notifications, audits: audits, bracketFlow: bracketFlow, email: email}
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
	if err := s.notifyMatchTeams(ctx, match, entity.NotificationMatchTimeChanged, notif.MatchTimeChanged); err != nil {
		return err
	}
	if in.ScheduledAt != nil {
		tournament, _ := s.tournaments.GetTournament(ctx, match.TournamentID)
		go s.emailMatchTeams(match, func(email, _ string) {
			loc := ""
			if in.LocationOrServer != nil {
				loc = *in.LocationOrServer
			}
			title := ""
			if tournament != nil {
				title = tournament.Title
			}
			s.email.SendMatchScheduled(email, title, *in.ScheduledAt, loc)
		})
	}
	return nil
}

func (s *MatchService) ConfirmReady(ctx context.Context, actorUserID, matchID string) error {
	return s.updateTeamConfirmation(ctx, actorUserID, matchID, entity.MatchTeamConfirmationReadyConfirmed, entity.MatchStatusConfirmed, entity.NotificationMatchAssigned, notif.MatchAssigned)
}

func (s *MatchService) RequestReschedule(ctx context.Context, actorUserID, matchID string) error {
	return s.updateTeamConfirmation(ctx, actorUserID, matchID, entity.MatchTeamConfirmationRescheduleRequested, entity.MatchStatusRescheduleRequested, entity.NotificationMatchRescheduled, notif.MatchRescheduled)
}

func (s *MatchService) ReportIssue(ctx context.Context, actorUserID, matchID string) error {
	return s.updateTeamConfirmation(ctx, actorUserID, matchID, entity.MatchTeamConfirmationIssueReported, entity.MatchStatusIssueReported, entity.NotificationMatchCancelled, notif.MatchCancelled)
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
	return s.notifyManagers(ctx, match.TournamentID, entity.NotificationResultSubmitted, notif.ResultSubmitted)
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
	if err := s.propagateOrUpdateStats(ctx, actorUserID, match); err != nil {
		return err
	}
	_ = s.audits.Create(ctx, &entity.AuditLog{ID: uuid.NewString(), ActorUserID: &actorUserID, TournamentID: &match.TournamentID, EntityType: "match", EntityID: match.ID, ActionType: "admin_set_result", Description: "Admin directly set match result", MetadataJSON: xjson.MustMarshal(map[string]string{"winner_team_id": in.WinnerTeamID})})
	if err := s.notifyMatchTeams(ctx, match, entity.NotificationResultConfirmed, notif.ResultConfirmed); err != nil {
		return err
	}
	s.sendResultConfirmedEmail(ctx, match)
	return nil
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
	if err := s.propagateOrUpdateStats(ctx, actorUserID, match); err != nil {
		return err
	}
	_ = s.audits.Create(ctx, &entity.AuditLog{ID: uuid.NewString(), ActorUserID: &actorUserID, TournamentID: &match.TournamentID, EntityType: "match", EntityID: match.ID, ActionType: "result_approved", Description: "Match result approved", MetadataJSON: xjson.MustMarshal(map[string]string{"winner_team_id": *match.WinnerTeamID})})
	if err := s.notifyMatchTeams(ctx, match, entity.NotificationResultConfirmed, notif.ResultConfirmed); err != nil {
		return err
	}
	s.sendResultConfirmedEmail(ctx, match)
	return nil
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

func (s *MatchService) updateTeamConfirmation(ctx context.Context, actorUserID, matchID, teamStatus, matchStatus, notificationType string, textsFunc func(string) notif.Texts) error {
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
	desc := textsFunc("ru").Title
	_ = s.audits.Create(ctx, &entity.AuditLog{ID: uuid.NewString(), ActorUserID: &actorUserID, TournamentID: &match.TournamentID, EntityType: "match", EntityID: match.ID, ActionType: "match_confirmation_updated", Description: desc, MetadataJSON: xjson.MustMarshal(map[string]string{"team_status": teamStatus})})
	return s.notifyManagers(ctx, match.TournamentID, notificationType, textsFunc)
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

func (s *MatchService) notifyManagers(ctx context.Context, tournamentID, typ string, textsFunc func(string) notif.Texts) error {
	userIDs, err := s.tournaments.tournaments.ListUserIDsByRoles(ctx, tournamentID, []string{entity.TournamentRoleOwner, entity.TournamentRoleManager})
	if err != nil {
		return err
	}
	langs := s.users.GetLangsByIDs(ctx, userIDs)
	for _, userID := range userIDs {
		lang := langs[userID]
		texts := textsFunc(lang)
		notification := &entity.Notification{
			ID:          uuid.NewString(),
			UserID:      userID,
			Type:        typ,
			Title:       texts.Title,
			Message:     texts.Message,
			PayloadJSON: xjson.MustMarshal(map[string]string{"tournament_id": tournamentID}),
		}
		if err := s.notifications.Create(ctx, notification); err != nil {
			return err
		}
	}
	return nil
}

func (s *MatchService) notifyMatchTeams(ctx context.Context, match *entity.Match, typ string, textsFunc func(string) notif.Texts) error {
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
			texts := textsFunc(member.UserLang)
			notification := &entity.Notification{ID: uuid.NewString(), UserID: *member.UserID, Type: typ, Title: texts.Title, Message: texts.Message, PayloadJSON: xjson.MustMarshal(map[string]string{"match_id": match.ID, "tournament_id": match.TournamentID})}
			if err := s.notifications.Create(ctx, notification); err != nil {
				return err
			}
		}
	}
	return nil
}

// emailMatchTeams looks up user emails for all team members in a match and calls fn for each.
func (s *MatchService) emailMatchTeams(match *entity.Match, fn func(email, userID string)) {
	ctx := context.Background()
	for _, teamID := range []*string{match.Team1ID, match.Team2ID} {
		if teamID == nil {
			continue
		}
		members, err := s.teams.ListMembersByTeamID(ctx, *teamID)
		if err != nil {
			continue
		}
		for _, member := range members {
			if member.UserID == nil {
				continue
			}
			user, err := s.users.GetByID(ctx, *member.UserID)
			if err != nil || user.Email == "" {
				continue
			}
			fn(user.Email, user.ID)
		}
	}
}

func (s *MatchService) sendResultConfirmedEmail(ctx context.Context, match *entity.Match) {
	if match.WinnerTeamID == nil {
		return
	}
	tournament, err := s.tournaments.GetTournament(ctx, match.TournamentID)
	if err != nil {
		return
	}
	winnerName := *match.WinnerTeamID
	if team, err := s.teams.GetTeamByID(ctx, *match.WinnerTeamID); err == nil {
		winnerName = team.Name
	}
	go s.emailMatchTeams(match, func(email, _ string) {
		s.email.SendResultConfirmed(email, tournament.Title, winnerName)
	})
}

// propagateOrUpdateStats decides how to handle a finished match result:
// - group_de group matches: propagate within the group DE bracket
// - group_stage group matches: update round-robin standings only
// - all other matches: propagate through the main bracket
func (s *MatchService) propagateOrUpdateStats(ctx context.Context, actorUserID string, match *entity.Match) error {
	if match.GroupID != nil {
		br, _ := s.brackets.GetByTournamentID(ctx, match.TournamentID)
		if br != nil && br.Format == "group_de" {
			return s.bracketFlow.PropagateWinner(ctx, actorUserID, match.ID)
		}
		s.updateGroupStats(ctx, match)
		return nil
	}
	return s.bracketFlow.PropagateWinner(ctx, actorUserID, match.ID)
}

func (s *MatchService) updateGroupStats(ctx context.Context, match *entity.Match) {
	if match.GroupID == nil || match.WinnerTeamID == nil || match.Team1ID == nil || match.Team2ID == nil {
		return
	}
	loserTeamID := *match.Team1ID
	if *match.Team1ID == *match.WinnerTeamID {
		loserTeamID = *match.Team2ID
	}
	_ = s.groups.RecordWin(ctx, *match.GroupID, *match.WinnerTeamID, loserTeamID)
}
