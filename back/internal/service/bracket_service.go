package service

import (
	"context"
	"math/rand"
	"time"

	"esports-backend/internal/apperror"
	"esports-backend/internal/bracket"
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

func NewBracketService(
	pool *pgxpool.Pool,
	tournaments *TournamentService,
	brackets *repository.BracketRepository,
	teams *repository.TeamRepository,
	notifications *repository.NotificationRepository,
	audits *repository.AuditRepository,
) *BracketService {
	return &BracketService{
		pool:          pool,
		tournaments:   tournaments,
		brackets:      brackets,
		teams:         teams,
		notifications: notifications,
		audits:        audits,
	}
}

type ReseedInput struct {
	OrderedTeamIDs []string
}

// Generate creates (or re-creates) the bracket for the given tournament.
// It reads tournament.Format to decide single vs double elimination.
func (s *BracketService) Generate(
	ctx context.Context,
	actorUserID, tournamentID string,
	regenerate bool,
	orderedTeamIDs []string,
) (*entity.Bracket, []entity.Match, error) {

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
	if tournament.Status == entity.TournamentStatusInProgress ||
		tournament.Status == entity.TournamentStatusFinished {
		return nil, nil, apperror.BadRequest("invalid_tournament_status",
			"bracket changes are not allowed after tournament starts", nil)
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
		return nil, nil, apperror.BadRequest("insufficient_teams",
			"at least 2 approved teams are required", nil)
	}

	seedingMethod := "random"
	if len(orderedTeamIDs) > 0 {
		if !sameSet(teamIDs, orderedTeamIDs) {
			return nil, nil, apperror.BadRequest("invalid_seeding",
				"ordered team ids must match approved teams exactly", nil)
		}
		teamIDs = append([]string(nil), orderedTeamIDs...)
		seedingMethod = "manual"
	} else {
		rand.New(rand.NewSource(time.Now().UnixNano())).
			Shuffle(len(teamIDs), func(i, j int) { teamIDs[i], teamIDs[j] = teamIDs[j], teamIDs[i] })
	}

	format := tournament.Format
	if format == "" {
		format = "single_elimination"
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

	br := &entity.Bracket{
		ID:            uuid.NewString(),
		TournamentID:  tournamentID,
		Format:        format,
		SeedingMethod: seedingMethod,
		Status:        "generated",
		GeneratedBy:   actorUserID,
		GeneratedAt:   time.Now(),
		MetadataJSON:  xjson.MustMarshal(map[string]interface{}{"teams_seeded": teamIDs}),
	}
	if err := txBracketRepo.CreateBracket(ctx, br); err != nil {
		return nil, nil, err
	}

	matches, err := buildEntityMatches(tournamentID, br.ID, format, teamIDs)
	if err != nil {
		return nil, nil, err
	}

	// Phase 1: insert without self-referential FK columns to avoid constraint violations.
	for _, m := range matches {
		nextID, src1ID, src2ID, loseNextID, loseSlot := m.NextMatchID, m.SourceMatch1ID, m.SourceMatch2ID, m.LoserNextMatchID, m.LoserNextSlot
		m.NextMatchID, m.SourceMatch1ID, m.SourceMatch2ID, m.LoserNextMatchID = nil, nil, nil, nil
		if err := txBracketRepo.CreateMatch(ctx, m); err != nil {
			return nil, nil, err
		}
		m.NextMatchID, m.SourceMatch1ID, m.SourceMatch2ID, m.LoserNextMatchID, m.LoserNextSlot = nextID, src1ID, src2ID, loseNextID, loseSlot
	}

	// Phase 2: update with FK refs and propagate byes.
	propagateByes(matches)
	for _, m := range matches {
		if err := txBracketRepo.UpdateMatchState(ctx, m); err != nil {
			return nil, nil, err
		}
	}

	if err := s.tournaments.tournaments.SetStatus(ctx, tournamentID,
		entity.TournamentStatusBracketGenerated); err != nil {
		return nil, nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, nil, err
	}

	_ = s.audits.Create(ctx, &entity.AuditLog{
		ID:           uuid.NewString(),
		ActorUserID:  &actorUserID,
		TournamentID: &tournamentID,
		EntityType:   "bracket",
		EntityID:     br.ID,
		ActionType:   "bracket_generated",
		Description:  "Bracket generated",
		MetadataJSON: xjson.MustMarshal(map[string]interface{}{
			"format":         format,
			"seeding_method": seedingMethod,
			"team_ids":       teamIDs,
		}),
	})

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

// ResetMatch clears the result of the given match and cascades the reset
// through all dependent matches (winner path and loser/LB path).
// If a Grand Final Reset match (GF round 2) exists, it is deleted.
// Tournament status is rolled back from "finished" to "in_progress" if needed.
func (s *BracketService) ResetMatch(ctx context.Context, actorUserID, matchID string) error {
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
	if match.WinnerTeamID == nil && !match.IsBye {
		return apperror.BadRequest("match_not_decided", "match has no result to reset", nil)
	}

	allMatches, err := s.brackets.ListMatchesByTournament(ctx, match.TournamentID)
	if err != nil {
		return err
	}
	byID := make(map[string]*entity.Match, len(allMatches))
	for i := range allMatches {
		cp := allMatches[i]
		byID[cp.ID] = &cp
	}

	toUpdate := make(map[string]*entity.Match)
	toDelete := make(map[string]bool)
	visited := make(map[string]bool)
	cascadeMatchReset(byID[matchID], byID, toUpdate, toDelete, visited)

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	txRepo := repository.NewBracketRepository(tx)
	for id, m := range toUpdate {
		if toDelete[id] {
			continue
		}
		if err := txRepo.UpdateMatchState(ctx, m); err != nil {
			return err
		}
	}
	for id := range toDelete {
		if err := txRepo.DeleteMatchByID(ctx, id); err != nil {
			return err
		}
	}

	tournament, _ := s.tournaments.tournaments.GetByID(ctx, match.TournamentID)
	if tournament != nil && tournament.Status == entity.TournamentStatusFinished {
		if err := s.tournaments.tournaments.SetStatus(ctx, match.TournamentID,
			entity.TournamentStatusInProgress); err != nil {
			return err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	_ = s.audits.Create(ctx, &entity.AuditLog{
		ID:           uuid.NewString(),
		ActorUserID:  &actorUserID,
		TournamentID: &match.TournamentID,
		EntityType:   "match",
		EntityID:     matchID,
		ActionType:   "match_reset",
		Description:  "Match result reset (cascade)",
		MetadataJSON: xjson.MustMarshal(map[string]string{"match_id": matchID}),
	})
	return nil
}

// cascadeMatchReset recursively clears a match's result and removes the
// placed winner/loser from all subsequent matches.
func cascadeMatchReset(
	m *entity.Match,
	byID map[string]*entity.Match,
	toUpdate map[string]*entity.Match,
	toDelete map[string]bool,
	visited map[string]bool,
) {
	if visited[m.ID] {
		return
	}
	visited[m.ID] = true

	if m.WinnerTeamID == nil && !m.IsBye {
		return // nothing to undo
	}

	prevWinnerID := m.WinnerTeamID
	prevLoserID := loserOf(m)

	// Reset this match.
	m.WinnerTeamID = nil
	m.ScoreText = nil
	m.ManagerComment = nil
	m.IsBye = false
	m.Status = entity.MatchStatusScheduled
	m.Team1ConfirmationStatus = entity.MatchTeamConfirmationPending
	m.Team2ConfirmationStatus = entity.MatchTeamConfirmationPending
	toUpdate[m.ID] = m

	// Cascade via winner path.
	if prevWinnerID != nil && m.NextMatchID != nil {
		next, ok := byID[*m.NextMatchID]
		if ok {
			if next.BracketSection == entity.BracketSectionGF && next.RoundNumber == 2 {
				// GF reset match: delete it entirely.
				toDelete[next.ID] = true
				visited[next.ID] = true
			} else {
				clearTeamFromMatch(next, *prevWinnerID)
				toUpdate[next.ID] = next
				cascadeMatchReset(next, byID, toUpdate, toDelete, visited)
			}
		}
	}

	// Cascade via loser path (WB → LB in double elimination).
	if prevLoserID != nil && m.LoserNextMatchID != nil {
		lb, ok := byID[*m.LoserNextMatchID]
		if ok {
			if m.LoserNextSlot == 1 {
				lb.Team1ID = nil
			} else {
				lb.Team2ID = nil
			}
			toUpdate[lb.ID] = lb
			cascadeMatchReset(lb, byID, toUpdate, toDelete, visited)
		}
	}
}

// clearTeamFromMatch removes teamID from whichever slot holds it.
func clearTeamFromMatch(m *entity.Match, teamID string) {
	if m.Team1ID != nil && *m.Team1ID == teamID {
		m.Team1ID = nil
	} else if m.Team2ID != nil && *m.Team2ID == teamID {
		m.Team2ID = nil
	}
}

func (s *BracketService) GetBracket(ctx context.Context, tournamentID string) (*entity.Bracket, []entity.Match, error) {
	br, err := s.brackets.GetByTournamentID(ctx, tournamentID)
	if err != nil {
		return nil, nil, apperror.NotFound("bracket not found")
	}
	matches, err := s.brackets.ListMatchesByTournament(ctx, tournamentID)
	if err != nil {
		return nil, nil, err
	}
	return br, matches, nil
}

// PropagateWinner places winner/loser into the next matches and handles GF
// reset creation and tournament completion.
func (s *BracketService) PropagateWinner(ctx context.Context, actorUserID, matchID string) error {
	match, err := s.brackets.GetMatchByID(ctx, matchID)
	if err != nil {
		return apperror.NotFound("match not found")
	}
	if match.WinnerTeamID == nil {
		return apperror.BadRequest("winner_not_set", "winner must be set before propagation", nil)
	}

	allMatches, err := s.brackets.ListMatchesByTournament(ctx, match.TournamentID)
	if err != nil {
		return err
	}
	byID := make(map[string]*entity.Match, len(allMatches))
	for i := range allMatches {
		cp := allMatches[i]
		byID[cp.ID] = &cp
	}
	current := byID[matchID]

	loserID := loserOf(current)

	// ── Winner path ────────────────────────────────────────────────────────
	if current.NextMatchID != nil {
		next := byID[*current.NextMatchID]
		if next != nil {
			fillSlot(next, current.ID, current.WinnerTeamID)
			smartAutoAdvance(next, byID)
			if err := s.brackets.UpdateMatchState(ctx, next); err != nil {
				return err
			}
			if next.IsBye && next.WinnerTeamID != nil {
				if err := s.PropagateWinner(ctx, actorUserID, next.ID); err != nil {
					return err
				}
			}
		}
	}

	// ── Loser path (WB → LB in double elimination) ─────────────────────────
	if current.LoserNextMatchID != nil && loserID != nil {
		lb := byID[*current.LoserNextMatchID]
		if lb != nil {
			if current.LoserNextSlot == 1 {
				lb.Team1ID = loserID
			} else {
				lb.Team2ID = loserID
			}
			smartAutoAdvance(lb, byID)
			if err := s.brackets.UpdateMatchState(ctx, lb); err != nil {
				return err
			}
			if lb.IsBye && lb.WinnerTeamID != nil {
				if err := s.PropagateWinner(ctx, actorUserID, lb.ID); err != nil {
					return err
				}
			}
		}
	}

	// ── Tournament completion / GF reset ────────────────────────────────────
	switch {
	case current.BracketSection == entity.BracketSectionGF && current.RoundNumber == 1:
		return s.handleGFResult(ctx, actorUserID, current, byID)

	case current.BracketSection == entity.BracketSectionGF && current.RoundNumber == 2:
		// GF reset complete → champion decided.
		return s.finishTournament(ctx, current.TournamentID)

	case current.NextMatchID == nil &&
		current.BracketSection == entity.BracketSectionWB:
		// Single elimination final (no GF section created).
		return s.finishTournament(ctx, current.TournamentID)
	}

	return nil
}

// handleGFResult is called after GF Round 1 finishes.
// If LB champion (Team2) won → create reset match.
// If WB champion (Team1) won → tournament over.
func (s *BracketService) handleGFResult(
	ctx context.Context,
	actorUserID string,
	gf *entity.Match,
	_ map[string]*entity.Match,
) error {
	wbChampionWon := gf.Team1ID != nil && gf.WinnerTeamID != nil &&
		*gf.WinnerTeamID == *gf.Team1ID

	if wbChampionWon {
		return s.finishTournament(ctx, gf.TournamentID)
	}

	// LB champion won → need a Grand Final Reset (GF round 2).
	br, err := s.brackets.GetByTournamentID(ctx, gf.TournamentID)
	if err != nil {
		return err
	}
	reset := &entity.Match{
		ID:                      uuid.NewString(),
		TournamentID:            gf.TournamentID,
		BracketID:               br.ID,
		BracketSection:          entity.BracketSectionGF,
		RoundNumber:             2,
		SlotIndex:               1,
		Team1ID:                 gf.Team1ID, // WB champion
		Team2ID:                 gf.Team2ID, // LB champion (now winner of GF1)
		Status:                  entity.MatchStatusScheduled,
		Team1ConfirmationStatus: entity.MatchTeamConfirmationPending,
		Team2ConfirmationStatus: entity.MatchTeamConfirmationPending,
	}
	return s.brackets.CreateMatch(ctx, reset)
}

func (s *BracketService) finishTournament(ctx context.Context, tournamentID string) error {
	_ = s.tournaments.tournaments.SetStatus(ctx, tournamentID, entity.TournamentStatusFinished)

	teams, _ := s.teams.ListByTournament(ctx, tournamentID, true)
	for _, team := range teams {
		members, _ := s.teams.ListMembersByTeamID(ctx, team.ID)
		for _, member := range members {
			if member.UserID != nil {
				_ = s.notifications.Create(ctx, &entity.Notification{
					ID:      uuid.NewString(),
					UserID:  *member.UserID,
					Type:    entity.NotificationTournamentFinished,
					Title:   "Турнир завершен",
					Message: "Турнир завершен. Финальный результат зафиксирован.",
					PayloadJSON: xjson.MustMarshal(map[string]string{
						"tournament_id": tournamentID,
					}),
				})
			}
		}
	}
	return nil
}

// ── Helpers ───────────────────────────────────────────────────────────────

// fillSlot puts teamID into the correct slot of dst based on which source match
// is feeding it (using SourceMatch1ID / SourceMatch2ID convention).
func fillSlot(dst *entity.Match, srcMatchID string, teamID *string) {
	if dst.SourceMatch1ID != nil && *dst.SourceMatch1ID == srcMatchID {
		dst.Team1ID = teamID
	} else if dst.SourceMatch2ID != nil && *dst.SourceMatch2ID == srcMatchID {
		dst.Team2ID = teamID
	}
}

// smartAutoAdvance auto-wins a match that has exactly one real team and is
// certain the other slot will never be filled (no pending source match).
// It prevents premature auto-advance when an opponent will arrive later from
// an unresolved source match (e.g. LB major rounds in double elimination).
func smartAutoAdvance(m *entity.Match, byID map[string]*entity.Match) {
	if m.WinnerTeamID != nil {
		return
	}
	// A slot is "definitively empty" (true BYE) only when either:
	//  - no source match feeds that slot (SourceMatchID == nil), OR
	//  - the source match has already been decided (WinnerTeamID != nil).
	slot1Pending := m.Team1ID == nil &&
		m.SourceMatch1ID != nil &&
		byID[*m.SourceMatch1ID] != nil &&
		byID[*m.SourceMatch1ID].WinnerTeamID == nil

	slot2Pending := m.Team2ID == nil &&
		m.SourceMatch2ID != nil &&
		byID[*m.SourceMatch2ID] != nil &&
		byID[*m.SourceMatch2ID].WinnerTeamID == nil

	if m.Team1ID != nil && m.Team2ID == nil && !slot2Pending {
		m.WinnerTeamID = m.Team1ID
		m.IsBye = true
		m.Status = entity.MatchStatusFinished
	} else if m.Team2ID != nil && m.Team1ID == nil && !slot1Pending {
		m.WinnerTeamID = m.Team2ID
		m.IsBye = true
		m.Status = entity.MatchStatusFinished
	}
}

// loserOf returns the non-winner team ID (nil if match has no winner or is a bye).
func loserOf(m *entity.Match) *string {
	if m.WinnerTeamID == nil || m.IsBye {
		return nil
	}
	if m.Team1ID != nil && *m.Team1ID != *m.WinnerTeamID {
		return m.Team1ID
	}
	if m.Team2ID != nil && *m.Team2ID != *m.WinnerTeamID {
		return m.Team2ID
	}
	return nil
}

// buildEntityMatches converts a pure bracket graph into entity.Match records
// ready for DB insertion. Teams from teamIDs are placed by seed position.
func buildEntityMatches(
	tournamentID, bracketID, format string,
	teamIDs []string,
) ([]*entity.Match, error) {
	var nodes []bracket.Node
	var err error

	switch format {
	case "double_elimination":
		nodes, err = bracket.BuildDouble(len(teamIDs))
	default:
		nodes, err = bracket.BuildSingle(len(teamIDs))
	}
	if err != nil {
		return nil, err
	}

	// Assign a UUID to every node index.
	uuidByIdx := make(map[int]string, len(nodes))
	for _, n := range nodes {
		uuidByIdx[n.Index] = uuid.NewString()
	}

	matches := make([]*entity.Match, 0, len(nodes))
	for _, n := range nodes {
		m := &entity.Match{
			ID:                      uuidByIdx[n.Index],
			TournamentID:            tournamentID,
			BracketID:               bracketID,
			BracketSection:          string(n.Section),
			RoundNumber:             n.Round,
			SlotIndex:               n.Slot,
			IsBye:                   n.IsBye,
			Status:                  entity.MatchStatusScheduled,
			Team1ConfirmationStatus: entity.MatchTeamConfirmationPending,
			Team2ConfirmationStatus: entity.MatchTeamConfirmationPending,
		}

		// Seed-based team placement (WB Round 1 only).
		if n.Seed1 > 0 {
			id := teamIDs[n.Seed1-1]
			m.Team1ID = &id
		}
		if n.Seed2 > 0 {
			id := teamIDs[n.Seed2-1]
			m.Team2ID = &id
		}

		// Source matches (winner of source → fills slot in this match).
		if n.Src1 >= 0 {
			id := uuidByIdx[n.Src1]
			m.SourceMatch1ID = &id
		}
		if n.Src2 >= 0 {
			id := uuidByIdx[n.Src2]
			m.SourceMatch2ID = &id
		}

		// Winner next.
		if n.WinNext >= 0 {
			id := uuidByIdx[n.WinNext]
			m.NextMatchID = &id
		}

		// Loser next (WB → LB drop in double elimination).
		if n.LoseNext >= 0 {
			id := uuidByIdx[n.LoseNext]
			m.LoserNextMatchID = &id
			m.LoserNextSlot = n.LoseSlot
		}

		matches = append(matches, m)
	}
	return matches, nil
}

// propagateByes auto-advances teams in Round-1 BYE matches and cascades
// upward, so higher rounds are pre-filled where possible.
func propagateByes(matches []*entity.Match) {
	byID := make(map[string]*entity.Match, len(matches))
	for _, m := range matches {
		byID[m.ID] = m
	}

	changed := true
	for changed {
		changed = false
		for _, m := range matches {
			// Pull winners from already-completed source matches.
			if m.SourceMatch1ID != nil && m.Team1ID == nil {
				if src, ok := byID[*m.SourceMatch1ID]; ok && src.WinnerTeamID != nil {
					m.Team1ID = src.WinnerTeamID
					changed = true
				}
			}
			if m.SourceMatch2ID != nil && m.Team2ID == nil {
				if src, ok := byID[*m.SourceMatch2ID]; ok && src.WinnerTeamID != nil {
					m.Team2ID = src.WinnerTeamID
					changed = true
				}
			}
			// Also pull LB slot from WB loser (LoserNextMatchID mechanism).
			// This is handled via the LoseSlot fields on WB source matches.
			// We scan all WB matches whose loser destination is this match.
			// (Done implicitly: if a WB bye match finishes, its loser would be
			//  nil since it's a bye. So LB slots from byes stay empty — correct.)

			// Auto-advance only when we're sure the absent slot has no pending source
			// match (i.e. the slot is truly a BYE, not an unplayed match).
			if m.WinnerTeamID == nil {
				absentSlot1IsPending := m.Team1ID == nil && m.SourceMatch1ID != nil &&
					byID[*m.SourceMatch1ID] != nil && byID[*m.SourceMatch1ID].WinnerTeamID == nil
				absentSlot2IsPending := m.Team2ID == nil && m.SourceMatch2ID != nil &&
					byID[*m.SourceMatch2ID] != nil && byID[*m.SourceMatch2ID].WinnerTeamID == nil

				if m.Team1ID != nil && m.Team2ID == nil && !absentSlot2IsPending {
					m.WinnerTeamID = m.Team1ID
					m.IsBye = true
					m.Status = entity.MatchStatusFinished
					changed = true
				} else if m.Team2ID != nil && m.Team1ID == nil && !absentSlot1IsPending {
					m.WinnerTeamID = m.Team2ID
					m.IsBye = true
					m.Status = entity.MatchStatusFinished
					changed = true
				}
			}
		}
	}
}

// nextPowerOfTwo returns the smallest power of two >= n.
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

// buildBracketMatches is a convenience wrapper used by unit tests.
// It generates a single-elimination bracket and returns the match slice.
func buildBracketMatches(tournamentID, bracketID string, teamIDs []string) ([]*entity.Match, error) {
	return buildEntityMatches(tournamentID, bracketID, "single_elimination", teamIDs)
}

func sameSet(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	counts := make(map[string]int, len(a))
	for _, v := range a {
		counts[v]++
	}
	for _, v := range b {
		counts[v]--
	}
	for _, v := range counts {
		if v != 0 {
			return false
		}
	}
	return true
}
