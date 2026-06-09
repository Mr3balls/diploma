package service

import (
	"context"
	"math/rand"
	"time"

	"esports-backend/internal/apperror"
	"esports-backend/internal/bracket"
	"esports-backend/internal/entity"
	"esports-backend/internal/pkg/notif"
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
	groups        *repository.GroupRepository
	teams         *repository.TeamRepository
	participants  *repository.ParticipantRepository
	users         repository.UserStore
	notifications *repository.NotificationRepository
	audits        *repository.AuditRepository
}

func NewBracketService(
	pool *pgxpool.Pool,
	tournaments *TournamentService,
	brackets *repository.BracketRepository,
	groups *repository.GroupRepository,
	teams *repository.TeamRepository,
	participants *repository.ParticipantRepository,
	users repository.UserStore,
	notifications *repository.NotificationRepository,
	audits *repository.AuditRepository,
) *BracketService {
	return &BracketService{
		pool:          pool,
		tournaments:   tournaments,
		brackets:      brackets,
		groups:        groups,
		teams:         teams,
		participants:  participants,
		users:         users,
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
	txGroupRepo := repository.NewGroupRepository(tx)
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

	if format == "group_stage" || format == "group_de" {
		var createFn func(context.Context, *entity.Tournament, *entity.Bracket, []string, *repository.GroupRepository, *repository.BracketRepository) error
		if format == "group_stage" {
			createFn = s.createGroupStage
		} else {
			createFn = s.createGroupDE
		}
		if err := createFn(ctx, tournament, br, teamIDs, txGroupRepo, txBracketRepo); err != nil {
			return nil, nil, err
		}
		if err := s.tournaments.tournaments.SetStatus(ctx, tournamentID,
			entity.TournamentStatusBracketGenerated); err != nil {
			return nil, nil, err
		}
		if err := tx.Commit(ctx); err != nil {
			return nil, nil, err
		}
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

	go s.notifyMatchAssigned(tournamentID, storedMatches)

	return storedBracket, storedMatches, nil
}

// notifyMatchAssigned sends match_assigned notifications after bracket generation.
// Handles both team-based and individual (participant-based) tournaments.
func (s *BracketService) notifyMatchAssigned(tournamentID string, matches []entity.Match) {
	ctx := context.Background()

	// Team-based: notify all members of teams that have a match.
	seenTeams := make(map[string]bool)
	for _, m := range matches {
		if m.Team1ID == nil || m.Team2ID == nil {
			continue
		}
		for _, teamID := range []string{*m.Team1ID, *m.Team2ID} {
			if seenTeams[teamID] {
				continue
			}
			seenTeams[teamID] = true
			members, err := s.teams.ListMembersByTeamID(ctx, teamID)
			if err != nil {
				continue
			}
			payload := xjson.MustMarshal(map[string]string{"match_id": m.ID, "tournament_id": tournamentID})
			for _, member := range members {
				if member.UserID == nil {
					continue
				}
				texts := notif.MatchAssigned(member.UserLang)
				_ = s.notifications.Create(ctx, &entity.Notification{
					ID:          uuid.NewString(),
					UserID:      *member.UserID,
					Type:        entity.NotificationMatchAssigned,
					Title:       texts.Title,
					Message:     texts.Message,
					PayloadJSON: payload,
				})
			}
		}
	}

	// Individual: notify participants that have a match (both sides set).
	type participantNotif struct {
		matchID string
		userID  string
	}
	seenParticipants := make(map[string]bool)
	var toNotify []participantNotif
	var userIDs []string
	for _, m := range matches {
		if m.Participant1ID == nil || m.Participant2ID == nil {
			continue
		}
		for _, pid := range []string{*m.Participant1ID, *m.Participant2ID} {
			if seenParticipants[pid] {
				continue
			}
			seenParticipants[pid] = true
			p, err := s.participants.GetByID(ctx, pid)
			if err != nil || p.UserID == nil {
				continue
			}
			toNotify = append(toNotify, participantNotif{matchID: m.ID, userID: *p.UserID})
			userIDs = append(userIDs, *p.UserID)
		}
	}
	if len(toNotify) == 0 {
		return
	}
	langs := s.users.GetLangsByIDs(ctx, userIDs)
	for _, n := range toNotify {
		texts := notif.MatchAssigned(langs[n.userID])
		payload := xjson.MustMarshal(map[string]string{"match_id": n.matchID, "tournament_id": tournamentID})
		_ = s.notifications.Create(ctx, &entity.Notification{
			ID:          uuid.NewString(),
			UserID:      n.userID,
			Type:        entity.NotificationMatchAssigned,
			Title:       texts.Title,
			Message:     texts.Message,
			PayloadJSON: payload,
		})
	}
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

type TeamBracketResponse struct {
	Bracket *entity.Bracket       `json:"bracket"`
	Groups  []entity.BracketGroup `json:"groups,omitempty"`
	Matches []entity.Match        `json:"matches"`
}

func (s *BracketService) GetBracket(ctx context.Context, tournamentID string) (*TeamBracketResponse, error) {
	br, err := s.brackets.GetByTournamentID(ctx, tournamentID)
	if err != nil {
		return nil, apperror.NotFound("bracket not found")
	}
	matches, err := s.brackets.ListMatchesByTournament(ctx, tournamentID)
	if err != nil {
		return nil, err
	}
	resp := &TeamBracketResponse{Bracket: br, Matches: matches}
	if br.Format == "group_stage" || br.Format == "group_de" {
		groups, err := s.groups.ListByBracketID(ctx, br.ID)
		if err != nil {
			return nil, err
		}
		resp.Groups = groups
	}
	return resp, nil
}

// createGroupStage generates groups and round-robin matches inside a transaction.
func (s *BracketService) createGroupStage(ctx context.Context, tournament *entity.Tournament, br *entity.Bracket, teamIDs []string, txGroupRepo *repository.GroupRepository, txBracketRepo *repository.BracketRepository) error {
	groupCount := 2
	if tournament.GroupCount != nil && *tournament.GroupCount >= 2 {
		groupCount = *tournament.GroupCount
	}

	groupNames := []string{"A", "B", "C", "D"}
	// Distribute teams evenly across groups
	groups := make([]*entity.BracketGroup, groupCount)
	for i := 0; i < groupCount; i++ {
		g := &entity.BracketGroup{
			ID:           uuid.NewString(),
			BracketID:    br.ID,
			TournamentID: tournament.ID,
			Name:         "Группа " + groupNames[i],
			Position:     i,
		}
		if err := txGroupRepo.CreateGroup(ctx, g); err != nil {
			return err
		}
		groups[i] = g
	}

	// Assign teams to groups (snake draft: 0,1,2,3,3,2,1,0,0,...)
	groupMembers := make([][]string, groupCount)
	for i, teamID := range teamIDs {
		gIdx := i % groupCount
		member := &entity.BracketGroupMember{
			ID:      uuid.NewString(),
			GroupID: groups[gIdx].ID,
			TeamID:  teamID,
		}
		if err := txGroupRepo.CreateMember(ctx, member); err != nil {
			return err
		}
		groupMembers[gIdx] = append(groupMembers[gIdx], teamID)
	}

	// Generate round-robin matches within each group
	globalNum := 1
	for gIdx, members := range groupMembers {
		groupID := groups[gIdx].ID
		round := 1
		for i := 0; i < len(members); i++ {
			for j := i + 1; j < len(members); j++ {
				t1 := members[i]
				t2 := members[j]
				m := &entity.Match{
					ID:                      uuid.NewString(),
					TournamentID:            tournament.ID,
					BracketID:               br.ID,
					GroupID:                 &groupID,
					BracketSection:          "WB",
					RoundNumber:             round,
					SlotIndex:               j - i,
					GlobalNumber:            globalNum,
					Team1ID:                 &t1,
					Team2ID:                 &t2,
					Status:                  entity.MatchStatusScheduled,
					Team1ConfirmationStatus: entity.MatchTeamConfirmationPending,
					Team2ConfirmationStatus: entity.MatchTeamConfirmationPending,
				}
				if err := txBracketRepo.CreateMatch(ctx, m); err != nil {
					return err
				}
				globalNum++
				round++
			}
		}
	}
	return nil
}

// createGroupDE generates a double-elimination sub-bracket (WB + LB, no GF)
// for each group and inserts all matches tagged with the group's ID.
func (s *BracketService) createGroupDE(ctx context.Context, tournament *entity.Tournament, br *entity.Bracket, teamIDs []string, txGroupRepo *repository.GroupRepository, txBracketRepo *repository.BracketRepository) error {
	groupCount := 2
	if tournament.GroupCount != nil && *tournament.GroupCount >= 2 {
		groupCount = *tournament.GroupCount
	}

	groupNames := []string{"A", "B", "C", "D"}
	groups := make([]*entity.BracketGroup, groupCount)
	for i := 0; i < groupCount; i++ {
		g := &entity.BracketGroup{
			ID:           uuid.NewString(),
			BracketID:    br.ID,
			TournamentID: tournament.ID,
			Name:         "Группа " + groupNames[i],
			Position:     i,
		}
		if err := txGroupRepo.CreateGroup(ctx, g); err != nil {
			return err
		}
		groups[i] = g
	}

	// Distribute teams using simple modulo (0,1,2,...,g-1,0,1,...).
	groupMembers := make([][]string, groupCount)
	for i, teamID := range teamIDs {
		gIdx := i % groupCount
		member := &entity.BracketGroupMember{
			ID:      uuid.NewString(),
			GroupID: groups[gIdx].ID,
			TeamID:  teamID,
		}
		if err := txGroupRepo.CreateMember(ctx, member); err != nil {
			return err
		}
		groupMembers[gIdx] = append(groupMembers[gIdx], teamID)
	}

	// Generate a DE sub-bracket for each group.
	for gIdx, members := range groupMembers {
		groupID := groups[gIdx].ID
		nodes, _, _, err := bracket.BuildGroupDE(len(members))
		if err != nil {
			return err
		}

		uuidByIdx := make(map[int]string, len(nodes))
		for _, n := range nodes {
			uuidByIdx[n.Index] = uuid.NewString()
		}

		matches := make([]*entity.Match, 0, len(nodes))
		for _, n := range nodes {
			m := &entity.Match{
				ID:                      uuidByIdx[n.Index],
				TournamentID:            tournament.ID,
				BracketID:               br.ID,
				GroupID:                 &groupID,
				BracketSection:          string(n.Section),
				RoundNumber:             n.Round,
				SlotIndex:               n.Slot,
				IsBye:                   n.IsBye,
				Status:                  entity.MatchStatusScheduled,
				Team1ConfirmationStatus: entity.MatchTeamConfirmationPending,
				Team2ConfirmationStatus: entity.MatchTeamConfirmationPending,
			}
			if n.Seed1 > 0 {
				id := members[n.Seed1-1]
				m.Team1ID = &id
			}
			if n.Seed2 > 0 {
				id := members[n.Seed2-1]
				m.Team2ID = &id
			}
			if n.Src1 >= 0 {
				id := uuidByIdx[n.Src1]
				m.SourceMatch1ID = &id
			}
			if n.Src2 >= 0 {
				id := uuidByIdx[n.Src2]
				m.SourceMatch2ID = &id
			}
			if n.WinNext >= 0 {
				id := uuidByIdx[n.WinNext]
				m.NextMatchID = &id
			}
			if n.LoseNext >= 0 {
				id := uuidByIdx[n.LoseNext]
				m.LoserNextMatchID = &id
				m.LoserNextSlot = n.LoseSlot
			}
			matches = append(matches, m)
		}

		// Phase 1: insert without self-referential FK columns.
		for _, m := range matches {
			nextID, src1ID, src2ID, loseNextID, loseSlot := m.NextMatchID, m.SourceMatch1ID, m.SourceMatch2ID, m.LoserNextMatchID, m.LoserNextSlot
			m.NextMatchID, m.SourceMatch1ID, m.SourceMatch2ID, m.LoserNextMatchID = nil, nil, nil, nil
			if err := txBracketRepo.CreateMatch(ctx, m); err != nil {
				return err
			}
			m.NextMatchID, m.SourceMatch1ID, m.SourceMatch2ID, m.LoserNextMatchID, m.LoserNextSlot = nextID, src1ID, src2ID, loseNextID, loseSlot
		}

		// Phase 2: propagate byes + write FK references.
		propagateByes(matches)
		for _, m := range matches {
			if err := txBracketRepo.UpdateMatchState(ctx, m); err != nil {
				return err
			}
		}
	}
	return nil
}

// AdvanceToPlayoff takes qualified teams per group and generates an SE playoff.
// For group_stage: top 2 per group → standard SE.
// For group_de: 3 seeds per group (QP=1→SF bye, QP=2/3→QF) → 8-slot SE.
func (s *BracketService) AdvanceToPlayoff(ctx context.Context, actorUserID, tournamentID string) (*TeamBracketResponse, error) {
	ok, err := s.tournaments.CanManageTournament(ctx, tournamentID, actorUserID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, apperror.Forbidden("insufficient tournament permissions")
	}

	br, err := s.brackets.GetByTournamentID(ctx, tournamentID)
	if err != nil {
		return nil, apperror.NotFound("bracket not found")
	}
	if br.Format != "group_stage" && br.Format != "group_de" {
		return nil, apperror.BadRequest("wrong_format", "tournament is not group stage or group DE", nil)
	}
	if br.Status == "playoff" {
		return nil, apperror.BadRequest("already_advanced", "already advanced to playoff", nil)
	}

	groups, err := s.groups.ListByBracketID(ctx, br.ID)
	if err != nil {
		return nil, err
	}

	var playoffTeamIDs []string

	if br.Format == "group_de" {
		// Collect 3 seeds per group ordered by qualified_position.
		// Seeding order: [G0_S1, G1_S1, ..., G0_S2, G1_S2, ..., G0_S3, G1_S3, ...]
		// With 2 groups → 6 teams in an 8-slot bracket:
		//   seeds 1,2 (S1s) get R1 byes → advance directly to SF.
		//   seeds 3-6 play QF (S2 vs opposite-group S3).
		seed1s := make([]string, 0, len(groups))
		seed2s := make([]string, 0, len(groups))
		seed3s := make([]string, 0, len(groups))
		for _, g := range groups {
			for _, m := range g.Members {
				if m.QualifiedPosition == nil {
					continue
				}
				switch *m.QualifiedPosition {
				case 1:
					seed1s = append(seed1s, m.TeamID)
				case 2:
					seed2s = append(seed2s, m.TeamID)
				case 3:
					seed3s = append(seed3s, m.TeamID)
				}
			}
		}
		if len(seed1s) == 0 {
			return nil, apperror.BadRequest("groups_not_complete", "group DE stages have not finished yet", nil)
		}
		playoffTeamIDs = append(playoffTeamIDs, seed1s...)
		playoffTeamIDs = append(playoffTeamIDs, seed2s...)
		playoffTeamIDs = append(playoffTeamIDs, seed3s...)
	} else {
		// group_stage: top 2 per group (sorted by points desc).
		for _, g := range groups {
			for i, m := range g.Members {
				if i >= 2 {
					break
				}
				playoffTeamIDs = append(playoffTeamIDs, m.TeamID)
			}
		}
	}

	if len(playoffTeamIDs) < 2 {
		return nil, apperror.BadRequest("insufficient_teams", "need at least 2 teams to advance", nil)
	}

	matches, err := buildEntityMatches(tournamentID, br.ID, "single_elimination", playoffTeamIDs)
	if err != nil {
		return nil, err
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	txBracketRepo := repository.NewBracketRepository(tx)

	if _, err := tx.Exec(ctx,
		`DELETE FROM matches WHERE tournament_id=$1 AND group_id IS NOT NULL`,
		tournamentID,
	); err != nil {
		return nil, err
	}

	for _, m := range matches {
		nextID, src1ID, src2ID := m.NextMatchID, m.SourceMatch1ID, m.SourceMatch2ID
		m.NextMatchID, m.SourceMatch1ID, m.SourceMatch2ID = nil, nil, nil
		if err := txBracketRepo.CreateMatch(ctx, m); err != nil {
			return nil, err
		}
		m.NextMatchID, m.SourceMatch1ID, m.SourceMatch2ID = nextID, src1ID, src2ID
	}
	propagateByes(matches)
	for _, m := range matches {
		if err := txBracketRepo.UpdateMatchState(ctx, m); err != nil {
			return nil, err
		}
	}

	if _, err := tx.Exec(ctx,
		`UPDATE brackets SET status='playoff' WHERE id=$1`,
		br.ID,
	); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	_ = s.audits.Create(ctx, &entity.AuditLog{
		ID:           uuid.NewString(),
		ActorUserID:  &actorUserID,
		TournamentID: &tournamentID,
		EntityType:   "bracket",
		EntityID:     br.ID,
		ActionType:   "advanced_to_playoff",
		Description:  "Advanced to playoff stage",
		MetadataJSON: xjson.MustMarshal(map[string]interface{}{"team_ids": playoffTeamIDs}),
	})

	return s.GetBracket(ctx, tournamentID)
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

	// ── Group DE: record qualification for terminal group matches ───────────
	// Group matches never trigger tournament completion; propagation within the
	// group bracket is handled normally above (winner/loser paths).
	if current.GroupID != nil {
		s.recordGroupDEQualification(ctx, current, loserID)
		return nil
	}

	// ── Tournament completion / GF reset ────────────────────────────────────
	switch {
	case current.BracketSection == entity.BracketSectionGF && current.RoundNumber == 1:
		return s.handleGFResult(ctx, actorUserID, current, byID)

	case current.BracketSection == entity.BracketSectionGF && current.RoundNumber == 2:
		// GF reset complete → champion decided.
		return s.finishTournament(ctx, current.TournamentID, current.WinnerTeamID, current.WinnerParticipantID)

	case current.NextMatchID == nil &&
		current.BracketSection == entity.BracketSectionWB:
		// Single elimination final (no GF section created).
		return s.finishTournament(ctx, current.TournamentID, current.WinnerTeamID, current.WinnerParticipantID)
	}

	return nil
}

// recordGroupDEQualification sets qualified_position in group members when a
// group WB Final or LB Final finishes.
//
//	WB Final winner → position 1 (SF seed); loser → position 2 (QF seed)
//	LB Final winner → position 3 (QF seed)
func (s *BracketService) recordGroupDEQualification(ctx context.Context, m *entity.Match, loserID *string) {
	isWBFinal := m.BracketSection == entity.BracketSectionWB && m.NextMatchID == nil
	isLBFinal := m.BracketSection == entity.BracketSectionLB && m.NextMatchID == nil
	if isWBFinal {
		if m.WinnerTeamID != nil {
			_ = s.groups.SetQualifiedPosition(ctx, *m.GroupID, *m.WinnerTeamID, 1)
		}
		if loserID != nil {
			_ = s.groups.SetQualifiedPosition(ctx, *m.GroupID, *loserID, 2)
		}
	} else if isLBFinal {
		if m.WinnerTeamID != nil {
			_ = s.groups.SetQualifiedPosition(ctx, *m.GroupID, *m.WinnerTeamID, 3)
		}
	}
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
		return s.finishTournament(ctx, gf.TournamentID, gf.WinnerTeamID, gf.WinnerParticipantID)
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

func (s *BracketService) finishTournament(ctx context.Context, tournamentID string, winnerTeamID, winnerParticipantID *string) error {
	_ = s.tournaments.tournaments.SetStatus(ctx, tournamentID, entity.TournamentStatusFinished)
	_ = s.tournaments.tournaments.SetWinner(ctx, tournamentID, winnerTeamID, winnerParticipantID)

	teams, _ := s.teams.ListByTournament(ctx, tournamentID, true)
	for _, team := range teams {
		members, _ := s.teams.ListMembersByTeamID(ctx, team.ID)
		for _, member := range members {
			if member.UserID != nil {
				texts := notif.TournamentFinished(member.UserLang)
				_ = s.notifications.Create(ctx, &entity.Notification{
					ID:      uuid.NewString(),
					UserID:  *member.UserID,
					Type:    entity.NotificationTournamentFinished,
					Title:   texts.Title,
					Message: texts.Message,
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

	// LB minor rounds (e.g. LB R1) receive both teams via LoserNext from WB
	// matches, so SourceMatch1/2ID are nil. Check byID for any undecided match
	// that will drop a loser into this match's empty slot before auto-advancing.
	if !slot1Pending && m.Team1ID == nil {
		for _, other := range byID {
			if other.LoserNextMatchID != nil && *other.LoserNextMatchID == m.ID &&
				other.LoserNextSlot == 1 && other.WinnerTeamID == nil {
				slot1Pending = true
				break
			}
		}
	}
	if !slot2Pending && m.Team2ID == nil {
		for _, other := range byID {
			if other.LoserNextMatchID != nil && *other.LoserNextMatchID == m.ID &&
				other.LoserNextSlot == 2 && other.WinnerTeamID == nil {
				slot2Pending = true
				break
			}
		}
	}

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

				// Also check LoserNext: LB minor rounds receive teams via WB LoserNext,
				// not SourceMatchID, so the pending check above misses them.
				if !absentSlot1IsPending && m.Team1ID == nil {
					for _, other := range byID {
						if other.LoserNextMatchID != nil && *other.LoserNextMatchID == m.ID &&
							other.LoserNextSlot == 1 && other.WinnerTeamID == nil {
							absentSlot1IsPending = true
							break
						}
					}
				}
				if !absentSlot2IsPending && m.Team2ID == nil {
					for _, other := range byID {
						if other.LoserNextMatchID != nil && *other.LoserNextMatchID == m.ID &&
							other.LoserNextSlot == 2 && other.WinnerTeamID == nil {
							absentSlot2IsPending = true
							break
						}
					}
				}

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
