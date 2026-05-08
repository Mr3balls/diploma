package service

import (
	"context"
	"math"
	"math/rand"
	"time"

	"esports-backend/internal/apperror"
	"esports-backend/internal/entity"
	"esports-backend/internal/pkg/xjson"
	"esports-backend/internal/repository"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type BracketService struct {
	pool          *pgxpool.Pool
	tournaments   *TournamentService
	brackets      *repository.BracketRepository
	teams         *repository.TeamRepository
	notifications *repository.NotificationRepository
	audits        *repository.AuditRepository
}

func NewBracketService(pool *pgxpool.Pool, tournaments *TournamentService, brackets *repository.BracketRepository, teams *repository.TeamRepository, notifications *repository.NotificationRepository, audits *repository.AuditRepository) *BracketService {
	return &BracketService{pool: pool, tournaments: tournaments, brackets: brackets, teams: teams, notifications: notifications, audits: audits}
}

type ReseedInput struct {
	OrderedTeamIDs []string
}

func (s *BracketService) Generate(ctx context.Context, actorUserID, tournamentID string, regenerate bool, orderedTeamIDs []string) (*entity.Bracket, []entity.Match, error) {
	ok, err := s.tournaments.CanManageTournament(ctx, tournamentID, actorUserID)
	if err != nil {
		return nil, nil, err
	}
	if !ok {
		return nil, nil, apperror.Forbidden("insufficient tournament permissions")
	}
	tournament, err := s.tournaments.tournaments.GetByID(ctx, tournamentID)
	if err != nil {
		return nil, nil, apperror.NotFound("tournament not found")
	}
	if tournament.Status == entity.TournamentStatusInProgress || tournament.Status == entity.TournamentStatusFinished {
		return nil, nil, apperror.BadRequest("invalid_tournament_status", "bracket changes are not allowed after tournament starts", nil)
	}
	_, err = s.brackets.GetByTournamentID(ctx, tournamentID)
	if err == nil && !regenerate {
		return nil, nil, apperror.Conflict("bracket already exists")
	}
	if err != nil && err != pgx.ErrNoRows {
		return nil, nil, err
	}

	teamIDs, err := s.teams.ListApprovedTeamIDs(ctx, tournamentID)
	if err != nil {
		return nil, nil, err
	}
	if len(teamIDs) < 2 {
		return nil, nil, apperror.BadRequest("insufficient_teams", "at least 2 approved teams are required", nil)
	}

	seedingMethod := "random"
	if len(orderedTeamIDs) > 0 {
		if !sameSet(teamIDs, orderedTeamIDs) {
			return nil, nil, apperror.BadRequest("invalid_seeding", "ordered team ids must match approved teams exactly", nil)
		}
		teamIDs = append([]string(nil), orderedTeamIDs...)
		seedingMethod = "manual"
	} else {
		rand.New(rand.NewSource(time.Now().UnixNano())).Shuffle(len(teamIDs), func(i, j int) {
			teamIDs[i], teamIDs[j] = teamIDs[j], teamIDs[i]
		})
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, nil, err
	}
	defer tx.Rollback(ctx)

	txBracketRepo := repository.NewBracketRepository(tx)
	if regenerate {
		if err := txBracketRepo.DeleteByTournamentID(ctx, tournamentID); err != nil {
			return nil, nil, err
		}
	}

	bracket := &entity.Bracket{
		ID:            uuid.NewString(),
		TournamentID:  tournamentID,
		Format:        "single_elimination",
		SeedingMethod: seedingMethod,
		Status:        "generated",
		GeneratedBy:   actorUserID,
		GeneratedAt:   time.Now(),
		MetadataJSON:  xjson.MustMarshal(map[string]interface{}{"teams_seeded": teamIDs}),
	}
	if err := txBracketRepo.CreateBracket(ctx, bracket); err != nil {
		return nil, nil, err
	}

	matches, err := buildBracketMatches(tournamentID, bracket.ID, teamIDs)
	if err != nil {
		return nil, nil, err
	}
	for _, match := range matches {
		if err := txBracketRepo.CreateMatch(ctx, match); err != nil {
			return nil, nil, err
		}
	}
	propagateByes(matches)
	for _, match := range matches {
		if err := txBracketRepo.UpdateMatchState(ctx, match); err != nil {
			return nil, nil, err
		}
	}
	if err := s.tournaments.tournaments.SetStatus(ctx, tournamentID, entity.TournamentStatusBracketGenerated); err != nil {
		return nil, nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, nil, err
	}
	_ = s.audits.Create(ctx, &entity.AuditLog{ID: uuid.NewString(), ActorUserID: &actorUserID, TournamentID: &tournamentID, EntityType: "bracket", EntityID: bracket.ID, ActionType: "bracket_generated", Description: "Bracket generated", MetadataJSON: xjson.MustMarshal(map[string]interface{}{"seeding_method": seedingMethod, "team_ids": teamIDs})})
	storedBracket, err := s.brackets.GetByTournamentID(ctx, tournamentID)
	if err != nil {
		return nil, nil, err
	}
	storedMatches, err := s.brackets.ListMatchesByTournament(ctx, tournamentID)
	if err != nil {
		return nil, nil, err
	}
	return storedBracket, storedMatches, nil
}

func (s *BracketService) GetBracket(ctx context.Context, tournamentID string) (*entity.Bracket, []entity.Match, error) {
	bracket, err := s.brackets.GetByTournamentID(ctx, tournamentID)
	if err != nil {
		return nil, nil, apperror.NotFound("bracket not found")
	}
	matches, err := s.brackets.ListMatchesByTournament(ctx, tournamentID)
	if err != nil {
		return nil, nil, err
	}
	return bracket, matches, nil
}

func (s *BracketService) PropagateWinner(ctx context.Context, actorUserID string, matchID string) error {
	match, err := s.brackets.GetMatchByID(ctx, matchID)
	if err != nil {
		return apperror.NotFound("match not found")
	}
	if match.WinnerTeamID == nil {
		return apperror.BadRequest("winner_not_set", "winner must be set before propagation", nil)
	}
	matches, err := s.brackets.ListMatchesByTournament(ctx, match.TournamentID)
	if err != nil {
		return err
	}
	byID := make(map[string]*entity.Match, len(matches))
	for i := range matches {
		item := matches[i]
		copied := item
		byID[item.ID] = &copied
	}
	current := byID[matchID]
	if current.NextMatchID != nil {
		nextMatch := byID[*current.NextMatchID]
		if nextMatch != nil {
			if nextMatch.SourceMatch1ID != nil && *nextMatch.SourceMatch1ID == current.ID {
				nextMatch.Team1ID = current.WinnerTeamID
			}
			if nextMatch.SourceMatch2ID != nil && *nextMatch.SourceMatch2ID == current.ID {
				nextMatch.Team2ID = current.WinnerTeamID
			}
			if nextMatch.Team1ID != nil && nextMatch.Team2ID == nil {
				nextMatch.WinnerTeamID = nextMatch.Team1ID
				nextMatch.IsBye = true
				nextMatch.Status = entity.MatchStatusFinished
			} else if nextMatch.Team2ID != nil && nextMatch.Team1ID == nil {
				nextMatch.WinnerTeamID = nextMatch.Team2ID
				nextMatch.IsBye = true
				nextMatch.Status = entity.MatchStatusFinished
			}
			if err := s.brackets.UpdateMatchState(ctx, nextMatch); err != nil {
				return err
			}
			if nextMatch.WinnerTeamID != nil && nextMatch.IsBye && nextMatch.NextMatchID != nil {
				if err := s.PropagateWinner(ctx, actorUserID, nextMatch.ID); err != nil {
					return err
				}
			}
		}
	} else {
		_ = s.tournaments.tournaments.SetStatus(ctx, match.TournamentID, entity.TournamentStatusFinished)
		teams, _ := s.teams.ListByTournament(ctx, match.TournamentID, true)
		for _, team := range teams {
			members, _ := s.teams.ListMembersByTeamID(ctx, team.ID)
			for _, member := range members {
				if member.UserID != nil {
					_ = s.notifications.Create(ctx, &entity.Notification{ID: uuid.NewString(), UserID: *member.UserID, Type: entity.NotificationTournamentFinished, Title: "Турнир завершен", Message: "Турнир завершен. Финальный результат зафиксирован.", PayloadJSON: xjson.MustMarshal(map[string]string{"tournament_id": match.TournamentID})})
				}
			}
		}
	}
	return nil
}

func buildBracketMatches(tournamentID, bracketID string, teamIDs []string) ([]*entity.Match, error) {
	bracketSize := nextPowerOfTwo(len(teamIDs))
	firstRoundMatches := bracketSize / 2
	rounds := int(math.Log2(float64(bracketSize)))
	seeds := make([]*string, bracketSize)
	for i := 0; i < bracketSize; i++ {
		if i < len(teamIDs) {
			teamID := teamIDs[i]
			seeds[i] = &teamID
		}
	}

	all := make([]*entity.Match, 0, bracketSize-1)
	prevRound := make([]*entity.Match, 0, firstRoundMatches)
	for i := 0; i < firstRoundMatches; i++ {
		match := &entity.Match{
			ID:                      uuid.NewString(),
			TournamentID:            tournamentID,
			BracketID:               bracketID,
			RoundNumber:             1,
			SlotIndex:               i + 1,
			Team1ID:                 seeds[i*2],
			Team2ID:                 seeds[i*2+1],
			Status:                  entity.MatchStatusScheduled,
			Team1ConfirmationStatus: entity.MatchTeamConfirmationPending,
			Team2ConfirmationStatus: entity.MatchTeamConfirmationPending,
		}
		prevRound = append(prevRound, match)
		all = append(all, match)
	}
	for round := 2; round <= rounds; round++ {
		currentCount := len(prevRound) / 2
		currentRound := make([]*entity.Match, 0, currentCount)
		for i := 0; i < currentCount; i++ {
			left := prevRound[i*2]
			right := prevRound[i*2+1]
			match := &entity.Match{
				ID:                      uuid.NewString(),
				TournamentID:            tournamentID,
				BracketID:               bracketID,
				RoundNumber:             round,
				SlotIndex:               i + 1,
				Status:                  entity.MatchStatusScheduled,
				Team1ConfirmationStatus: entity.MatchTeamConfirmationPending,
				Team2ConfirmationStatus: entity.MatchTeamConfirmationPending,
				SourceMatch1ID:          &left.ID,
				SourceMatch2ID:          &right.ID,
			}
			left.NextMatchID = &match.ID
			right.NextMatchID = &match.ID
			currentRound = append(currentRound, match)
			all = append(all, match)
		}
		prevRound = currentRound
	}
	return all, nil
}

func propagateByes(matches []*entity.Match) {
	byID := make(map[string]*entity.Match, len(matches))
	for _, match := range matches {
		byID[match.ID] = match
	}
	changed := true
	for changed {
		changed = false
		for _, match := range matches {
			if match.SourceMatch1ID != nil && match.Team1ID == nil {
				if source, ok := byID[*match.SourceMatch1ID]; ok && source.WinnerTeamID != nil {
					match.Team1ID = source.WinnerTeamID
					changed = true
				}
			}
			if match.SourceMatch2ID != nil && match.Team2ID == nil {
				if source, ok := byID[*match.SourceMatch2ID]; ok && source.WinnerTeamID != nil {
					match.Team2ID = source.WinnerTeamID
					changed = true
				}
			}
			if match.WinnerTeamID == nil {
				if match.Team1ID != nil && match.Team2ID == nil {
					match.WinnerTeamID = match.Team1ID
					match.IsBye = true
					match.Status = entity.MatchStatusFinished
					changed = true
				}
				if match.Team2ID != nil && match.Team1ID == nil {
					match.WinnerTeamID = match.Team2ID
					match.IsBye = true
					match.Status = entity.MatchStatusFinished
					changed = true
				}
			}
		}
	}
}

func nextPowerOfTwo(n int) int {
	if n <= 1 {
		return 1
	}
	p := 1
	for p < n {
		p <<= 1
	}
	return p
}

func sameSet(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	counts := make(map[string]int, len(a))
	for _, item := range a {
		counts[item]++
	}
	for _, item := range b {
		counts[item]--
	}
	for _, v := range counts {
		if v != 0 {
			return false
		}
	}
	return true
}
