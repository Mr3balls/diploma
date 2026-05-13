package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"time"

	"esports-backend/internal/apperror"
	"esports-backend/internal/bracket"
	"esports-backend/internal/entity"
	"esports-backend/internal/repository"
	ws "esports-backend/internal/transport/websocket"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// slugChars is the alphabet for auto-generated slugs.
const slugChars = "abcdefghijklmnopqrstuvwxyz0123456789"

// ChallongeService handles Challonge-style tournament lifecycle.
type ChallongeService struct {
	pool         *pgxpool.Pool
	tournaments  *repository.TournamentRepository
	bracketRepo  *repository.BracketRepository
	participants *repository.ParticipantRepository
	members      *repository.TournamentMemberRepository
	reports      *repository.ResultReportRepository
	matchLog     *repository.MatchLogRepository
	invites      *repository.CoOrganizerInviteRepository
	users        *repository.UserRepository
	hub          *ws.Hub
}

func NewChallongeService(
	pool *pgxpool.Pool,
	tournaments *repository.TournamentRepository,
	bracketRepo *repository.BracketRepository,
	participants *repository.ParticipantRepository,
	members *repository.TournamentMemberRepository,
	reports *repository.ResultReportRepository,
	matchLog *repository.MatchLogRepository,
	invites *repository.CoOrganizerInviteRepository,
	users *repository.UserRepository,
	hub *ws.Hub,
) *ChallongeService {
	return &ChallongeService{
		pool:         pool,
		tournaments:  tournaments,
		bracketRepo:  bracketRepo,
		participants: participants,
		members:      members,
		reports:      reports,
		matchLog:     matchLog,
		invites:      invites,
		users:        users,
		hub:          hub,
	}
}

// ── Slug helpers ──────────────────────────────────────────────────────────────

func generateSlug(length int) (string, error) {
	b := make([]byte, length)
	for i := range b {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(slugChars))))
		if err != nil {
			return "", err
		}
		b[i] = slugChars[n.Int64()]
	}
	return string(b), nil
}

func generateInviteToken() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// ── Tournament creation ───────────────────────────────────────────────────────

type CreateChallongeTournamentReq struct {
	Name            string     `json:"name"             validate:"required"`
	Description     string     `json:"description"`
	Game            string     `json:"game"`
	Format          string     `json:"format"           validate:"required,oneof=single_elimination double_elimination"`
	MaxParticipants int        `json:"max_participants"`
	StartTime       *time.Time `json:"start_time"`
	Privacy         string     `json:"privacy"          validate:"required,oneof=public unlisted private"`
	AllowSpectators bool       `json:"allow_spectators"`
	Slug            string     `json:"slug"` // optional; auto-generated if empty
}

// CreateTournament creates a Challonge-style tournament and assigns the creator as organizer.
func (s *ChallongeService) CreateTournament(ctx context.Context, creatorID string, req CreateChallongeTournamentReq) (*entity.Tournament, error) {
	slug, err := s.resolveSlug(ctx, req.Slug)
	if err != nil {
		return nil, err
	}

	privacy := req.Privacy
	if privacy == "unlisted" {
		// Map unlisted → private in visibility column; slug distinguishes it.
		privacy = "private"
	}

	t := &entity.Tournament{
		ID:              uuid.New().String(),
		Title:           req.Name,
		Description:     strPtr(req.Description),
		Discipline:      req.Game,
		Format:          req.Format,
		MaxParticipants: req.MaxParticipants,
		Status:          entity.TournamentStatusDraft,
		Visibility:      privacy,
		Slug:            &slug,
		OwnerUserID:     creatorID,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
	if req.StartTime != nil {
		t.StartAt = req.StartTime
	}

	if err := s.tournaments.Create(ctx, t); err != nil {
		return nil, apperror.Internal("failed to create tournament")
	}

	// Assign creator as organizer in tournament_members.
	m := &entity.TournamentMember{
		ID:           uuid.New().String(),
		TournamentID: t.ID,
		UserID:       creatorID,
		Role:         entity.TournamentRoleOrganizer,
		JoinedAt:     time.Now(),
	}
	if err := s.members.Upsert(ctx, m); err != nil {
		return nil, apperror.Internal("failed to assign organizer role")
	}

	return t, nil
}

func (s *ChallongeService) resolveSlug(ctx context.Context, requested string) (string, error) {
	if requested != "" {
		_, err := s.tournaments.GetBySlug(ctx, requested)
		if err == nil {
			return "", apperror.Conflict("slug already taken")
		}
		if !errors.Is(err, pgx.ErrNoRows) {
			return "", apperror.Internal("slug lookup failed")
		}
		return requested, nil
	}
	// Auto-generate: try lengths 8, 9, 10, 11.
	for l := 8; l <= 11; l++ {
		slug, err := generateSlug(l)
		if err != nil {
			return "", apperror.Internal("slug generation failed")
		}
		_, err = s.tournaments.GetBySlug(ctx, slug)
		if errors.Is(err, pgx.ErrNoRows) {
			return slug, nil
		}
	}
	return "", apperror.Internal("could not generate unique slug")
}

// GetBySlug retrieves a tournament by its URL slug.
func (s *ChallongeService) GetBySlug(ctx context.Context, slug string) (*entity.Tournament, error) {
	t, err := s.tournaments.GetBySlug(ctx, slug)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apperror.NotFound("tournament not found")
	}
	return t, err
}

// ── Participants ──────────────────────────────────────────────────────────────

type AddParticipantReq struct {
	Name   string  `json:"name"    validate:"required"`
	UserID *string `json:"user_id"`
}

func (s *ChallongeService) AddParticipant(ctx context.Context, tournamentID, actorID string, req AddParticipantReq) (*entity.Participant, error) {
	t, err := s.getTournamentGuard(ctx, tournamentID, entity.TournamentStatusDraft, entity.TournamentStatusReady, entity.TournamentStatusRegistrationOpen, entity.TournamentStatusRegistrationClosed, entity.TournamentStatusBracketGenerated)
	if err != nil {
		return nil, err
	}
	if t.MaxParticipants > 0 {
		count, err := s.participants.Count(ctx, tournamentID)
		if err != nil {
			return nil, apperror.Internal("participant count failed")
		}
		if count >= t.MaxParticipants {
			return nil, apperror.BadRequest("max_participants_reached", fmt.Sprintf("max %d participants", t.MaxParticipants), nil)
		}
	}
	count, _ := s.participants.Count(ctx, tournamentID)
	p := &entity.Participant{
		ID:           uuid.New().String(),
		TournamentID: tournamentID,
		UserID:       req.UserID,
		Name:         req.Name,
		Seed:         count + 1,
		Status:       entity.ParticipantStatusActive,
		CreatedAt:    time.Now(),
	}
	if err := s.participants.Create(ctx, p); err != nil {
		return nil, apperror.Internal("failed to add participant")
	}
	if req.UserID != nil {
		_ = s.members.Upsert(ctx, &entity.TournamentMember{
			ID: uuid.New().String(), TournamentID: tournamentID,
			UserID: *req.UserID, Role: entity.TournamentRoleParticipant, JoinedAt: time.Now(),
		})
	}
	s.maybeSetReady(ctx, tournamentID)
	s.hub.Broadcast(s.slugOf(t), ws.Event{Type: "participant_added", Payload: p})
	return p, nil
}

func (s *ChallongeService) BulkAddParticipants(ctx context.Context, tournamentID, actorID string, names []string) ([]*entity.Participant, error) {
	var out []*entity.Participant
	for _, name := range names {
		p, err := s.AddParticipant(ctx, tournamentID, actorID, AddParticipantReq{Name: name})
		if err != nil {
			return out, err
		}
		out = append(out, p)
	}
	return out, nil
}

func (s *ChallongeService) RemoveParticipant(ctx context.Context, tournamentID, participantID, actorID string) error {
	t, err := s.getTournamentGuard(ctx, tournamentID, entity.TournamentStatusDraft, entity.TournamentStatusReady, entity.TournamentStatusRegistrationOpen, entity.TournamentStatusRegistrationClosed, entity.TournamentStatusBracketGenerated)
	if err != nil {
		return err
	}
	if err := s.participants.Delete(ctx, participantID); err != nil {
		return apperror.Internal("failed to remove participant")
	}
	s.normalizeSeedOrder(ctx, tournamentID)
	s.hub.Broadcast(s.slugOf(t), ws.Event{Type: "participant_removed", Payload: map[string]string{"id": participantID}})
	return nil
}

func (s *ChallongeService) ShuffleParticipants(ctx context.Context, tournamentID, actorID string) error {
	t, err := s.getTournamentGuard(ctx, tournamentID, entity.TournamentStatusDraft, entity.TournamentStatusReady, entity.TournamentStatusRegistrationOpen, entity.TournamentStatusRegistrationClosed, entity.TournamentStatusBracketGenerated)
	if err != nil {
		return err
	}
	ps, err := s.participants.ListByTournament(ctx, tournamentID)
	if err != nil {
		return apperror.Internal("list participants failed")
	}
	// Fisher-Yates shuffle
	for i := len(ps) - 1; i > 0; i-- {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(i+1)))
		j := int(n.Int64())
		ps[i], ps[j] = ps[j], ps[i]
	}
	for i, p := range ps {
		_ = s.participants.UpdateSeed(ctx, p.ID, i+1)
	}
	s.hub.Broadcast(s.slugOf(t), ws.Event{Type: "seeds_updated", Payload: map[string]string{"tournament_id": tournamentID}})
	return nil
}

type ParticipantSeedItem struct {
	ParticipantID string `json:"participant_id"`
	Seed          int    `json:"seed"`
}

func (s *ChallongeService) ReorderParticipants(ctx context.Context, tournamentID, actorID string, items []ParticipantSeedItem) error {
	t, err := s.getTournamentGuard(ctx, tournamentID, entity.TournamentStatusDraft, entity.TournamentStatusReady, entity.TournamentStatusRegistrationOpen, entity.TournamentStatusRegistrationClosed, entity.TournamentStatusBracketGenerated)
	if err != nil {
		return err
	}
	// Validate: seeds must be 1..N contiguous.
	seedSet := make(map[int]bool, len(items))
	for _, item := range items {
		seedSet[item.Seed] = true
	}
	for i := 1; i <= len(items); i++ {
		if !seedSet[i] {
			return apperror.BadRequest("invalid_seeds", "seeds must be a contiguous 1..N range", nil)
		}
	}
	for _, item := range items {
		if err := s.participants.UpdateSeed(ctx, item.ParticipantID, item.Seed); err != nil {
			return apperror.Internal("update seed failed")
		}
	}
	s.hub.Broadcast(s.slugOf(t), ws.Event{Type: "seeds_updated", Payload: map[string]string{"tournament_id": tournamentID}})
	return nil
}

// ── Tournament lifecycle ──────────────────────────────────────────────────────

func (s *ChallongeService) Start(ctx context.Context, tournamentID, actorID string) error {
	t, err := s.getTournamentGuard(ctx, tournamentID,
		entity.TournamentStatusDraft,
		entity.TournamentStatusReady,
		entity.TournamentStatusRegistrationOpen,
		entity.TournamentStatusRegistrationClosed,
		entity.TournamentStatusBracketGenerated,
	)
	if err != nil {
		return err
	}
	ps, err := s.participants.ListByTournament(ctx, tournamentID)
	if err != nil || len(ps) < 2 {
		return apperror.BadRequest("insufficient_participants", "at least 2 participants required", nil)
	}
	if err := s.generateBracket(ctx, t, ps, actorID); err != nil {
		return err
	}
	if err := s.tournaments.UpdateStatus(ctx, tournamentID, entity.TournamentStatusInProgress); err != nil {
		return apperror.Internal("status update failed")
	}
	s.hub.Broadcast(s.slugOf(t), ws.Event{Type: "tournament_started", Payload: map[string]interface{}{
		"format": t.Format, "participants": len(ps),
	}})
	return nil
}

func (s *ChallongeService) Reset(ctx context.Context, tournamentID, actorID string) error {
	t, err := s.getTournament(ctx, tournamentID)
	if err != nil {
		return err
	}
	if err := s.bracketRepo.DeleteByTournamentID(ctx, tournamentID); err != nil {
		return apperror.Internal("bracket delete failed")
	}
	s.normalizeSeedOrder(ctx, tournamentID)
	_ = s.tournaments.UpdateStatus(ctx, tournamentID, entity.TournamentStatusDraft)
	s.maybeSetReady(ctx, tournamentID)
	s.hub.Broadcast(s.slugOf(t), ws.Event{Type: "tournament_reset", Payload: nil})
	return nil
}

func (s *ChallongeService) Unfinalize(ctx context.Context, tournamentID, actorID string) error {
	t, err := s.getTournamentGuard(ctx, tournamentID, entity.TournamentStatusCompleted)
	if err != nil {
		return err
	}
	if err := s.tournaments.UpdateStatus(ctx, tournamentID, entity.TournamentStatusInProgress); err != nil {
		return apperror.Internal("status update failed")
	}
	s.hub.Broadcast(s.slugOf(t), ws.Event{Type: "tournament_unfinalized", Payload: map[string]string{"by_user_id": actorID}})
	return nil
}

// ── Match results ─────────────────────────────────────────────────────────────

type SubmitResultReq struct {
	WinnerID string `json:"winner_id" validate:"required"` // participant ID
	Score1   int    `json:"score1"`
	Score2   int    `json:"score2"`
}

func (s *ChallongeService) SubmitResult(ctx context.Context, tournamentID, matchID, actorID string, req SubmitResultReq) error {
	if req.Score1 == req.Score2 {
		return apperror.BadRequest("draw_not_allowed", "draws are not allowed; winner_id is required", nil)
	}
	m, err := s.bracketRepo.GetMatchByID(ctx, matchID)
	if err != nil {
		return apperror.NotFound("match not found")
	}
	if m.TournamentID != tournamentID {
		return apperror.Forbidden("match does not belong to this tournament")
	}
	if m.WinnerParticipantID != nil {
		return apperror.BadRequest("already_decided", "match already has a result; reset first", nil)
	}
	winner, err := s.participants.GetByID(ctx, req.WinnerID)
	if err != nil || winner.TournamentID != tournamentID {
		return apperror.BadRequest("invalid_winner", "winner_id must be a participant of this tournament", nil)
	}
	// Determine loser.
	var loserID *string
	switch req.WinnerID {
	case ptrVal(m.Participant1ID):
		loserID = m.Participant2ID
	case ptrVal(m.Participant2ID):
		loserID = m.Participant1ID
	default:
		return apperror.BadRequest("invalid_winner", "winner_id must be one of the match participants", nil)
	}

	score := fmt.Sprintf("%d:%d", req.Score1, req.Score2)
	m.WinnerParticipantID = &req.WinnerID
	m.ScoreText = &score
	m.Status = entity.MatchStatusFinished
	if err := s.bracketRepo.UpdateMatchState(ctx, m); err != nil {
		return apperror.Internal("match update failed")
	}
	s.logMatchAction(ctx, tournamentID, matchID, actorID, "result_submitted", map[string]interface{}{
		"winner_id": req.WinnerID, "score1": req.Score1, "score2": req.Score2, "via": "organizer",
	})
	if err := s.advanceParticipant(ctx, m, req.WinnerID, ptrVal(loserID)); err != nil {
		return err
	}
	t, _ := s.getTournament(ctx, tournamentID)
	s.hub.Broadcast(s.slugOf(t), ws.Event{Type: "match_result", Payload: m})
	return nil
}

// AdminSetParticipantResult lets an organizer directly set the result of an individual-mode match.
// Permission check is expected to be performed by the caller.
func (s *ChallongeService) AdminSetParticipantResult(ctx context.Context, matchID, actorID, winnerParticipantID string) error {
	m, err := s.bracketRepo.GetMatchByID(ctx, matchID)
	if err != nil {
		return apperror.NotFound("match not found")
	}
	return s.SubmitResult(ctx, m.TournamentID, matchID, actorID, SubmitResultReq{
		WinnerID: winnerParticipantID,
		Score1:   1,
		Score2:   0,
	})
}

func (s *ChallongeService) ResetMatch(ctx context.Context, tournamentID, matchID, actorID string) error {
	t, err := s.getTournament(ctx, tournamentID)
	if err != nil {
		return err
	}
	if t.Status == entity.TournamentStatusCompleted {
		return apperror.BadRequest("tournament_completed", "un-finalize the tournament before resetting a match", nil)
	}
	m, err := s.bracketRepo.GetMatchByID(ctx, matchID)
	if err != nil {
		return apperror.NotFound("match not found")
	}
	if m.TournamentID != tournamentID {
		return apperror.Forbidden("match does not belong to this tournament")
	}
	// Load all matches for cascade computation.
	allMatches, err := s.bracketRepo.ListMatchesByTournament(ctx, tournamentID)
	if err != nil {
		return apperror.Internal("list matches failed")
	}
	// Build bracket node list from matches for CascadeReset.
	nodes := matchesToNodes(allMatches)
	affected := bracket.CascadeReset(nodes, matchIndexOf(nodes, matchID))

	// Clear current match result.
	m.WinnerParticipantID = nil
	m.ScoreText = nil
	m.Status = entity.MatchStatusScheduled
	_ = s.bracketRepo.UpdateMatchState(ctx, m)
	s.logMatchAction(ctx, tournamentID, matchID, actorID, "result_reset", nil)

	// Clear downstream matches.
	for _, nodeIdx := range affected {
		for i := range allMatches {
			if allMatches[i].GlobalNumber == nodeIdx+1 {
				am := &allMatches[i]
				am.WinnerParticipantID = nil
				am.Participant1ID = nil
				am.Participant2ID = nil
				am.ScoreText = nil
				am.Status = entity.MatchStatusScheduled
				_ = s.bracketRepo.UpdateMatchState(ctx, am)
			}
		}
	}
	s.hub.Broadcast(s.slugOf(t), ws.Event{Type: "match_reset", Payload: map[string]string{"match_id": matchID}})
	return nil
}

// ── Participant result reporting ──────────────────────────────────────────────

func (s *ChallongeService) ReportResult(ctx context.Context, tournamentID, matchID, actorID string, req SubmitResultReq) (*entity.ResultReport, error) {
	if req.Score1 == req.Score2 {
		return nil, apperror.BadRequest("draw_not_allowed", "draws are not allowed", nil)
	}
	// Verify the actor is a participant in this specific match.
	m, err := s.bracketRepo.GetMatchByID(ctx, matchID)
	if err != nil {
		return nil, apperror.NotFound("match not found")
	}
	actorParticipant, err := s.participants.GetByUserAndTournament(ctx, tournamentID, actorID)
	if err != nil {
		return nil, apperror.Forbidden("you are not a participant in this tournament")
	}
	if ptrVal(m.Participant1ID) != actorParticipant.ID && ptrVal(m.Participant2ID) != actorParticipant.ID {
		return nil, apperror.Forbidden("you are not playing in this match")
	}
	rr := &entity.ResultReport{
		ID:           uuid.New().String(),
		MatchID:      matchID,
		ReportedByID: actorID,
		WinnerID:     req.WinnerID,
		Score1:       req.Score1,
		Score2:       req.Score2,
		Status:       entity.ResultReportPending,
		CreatedAt:    time.Now(),
	}
	if err := s.reports.Create(ctx, rr); err != nil {
		return nil, apperror.Internal("create report failed")
	}
	t, _ := s.getTournament(ctx, tournamentID)
	s.hub.Broadcast(s.slugOf(t), ws.Event{Type: "result_report_submitted", Payload: rr})
	return rr, nil
}

func (s *ChallongeService) ApproveReport(ctx context.Context, tournamentID, matchID, reportID, actorID string) error {
	rr, err := s.reports.GetByID(ctx, reportID)
	if err != nil {
		return apperror.NotFound("report not found")
	}
	if rr.MatchID != matchID {
		return apperror.BadRequest("mismatch", "report does not belong to this match", nil)
	}
	if rr.Status != entity.ResultReportPending {
		return apperror.BadRequest("report_not_pending", "report is not pending", nil)
	}
	if err := s.reports.UpdateStatus(ctx, reportID, entity.ResultReportApproved); err != nil {
		return apperror.Internal("report update failed")
	}
	// Apply the result.
	if err := s.SubmitResult(ctx, tournamentID, matchID, actorID, SubmitResultReq{
		WinnerID: rr.WinnerID, Score1: rr.Score1, Score2: rr.Score2,
	}); err != nil {
		return err
	}
	// Override last log entry's via field.
	s.logMatchAction(ctx, tournamentID, matchID, actorID, "result_submitted", map[string]interface{}{
		"winner_id": rr.WinnerID, "score1": rr.Score1, "score2": rr.Score2, "via": "approved_report",
	})
	t, _ := s.getTournament(ctx, tournamentID)
	s.hub.Broadcast(s.slugOf(t), ws.Event{Type: "result_report_approved", Payload: rr})
	return nil
}

func (s *ChallongeService) RejectReport(ctx context.Context, tournamentID, matchID, reportID, actorID string) error {
	rr, err := s.reports.GetByID(ctx, reportID)
	if err != nil {
		return apperror.NotFound("report not found")
	}
	if rr.Status != entity.ResultReportPending {
		return apperror.BadRequest("report_not_pending", "report is not pending", nil)
	}
	if err := s.reports.UpdateStatus(ctx, reportID, entity.ResultReportRejected); err != nil {
		return apperror.Internal("report update failed")
	}
	t, _ := s.getTournament(ctx, tournamentID)
	s.hub.Broadcast(s.slugOf(t), ws.Event{Type: "result_report_rejected", Payload: rr})
	return nil
}

// ── Data retrieval ────────────────────────────────────────────────────────────

type BracketResponse struct {
	Tournament      *entity.Tournament    `json:"tournament"`
	Matches         []entity.Match        `json:"matches"`
	Participants    []*entity.Participant `json:"participants"`
	CurrentUserRole string                `json:"current_user_role,omitempty"`
}

func (s *ChallongeService) GetBracket(ctx context.Context, tournamentID, currentUserID string) (*BracketResponse, error) {
	t, err := s.getTournament(ctx, tournamentID)
	if err != nil {
		return nil, err
	}
	matches, err := s.bracketRepo.ListMatchesByTournament(ctx, tournamentID)
	if err != nil {
		return nil, apperror.Internal("list matches failed")
	}
	ps, err := s.participants.ListByTournament(ctx, tournamentID)
	if err != nil {
		return nil, apperror.Internal("list participants failed")
	}
	role := ""
	if currentUserID != "" {
		role, _ = s.members.GetRole(ctx, tournamentID, currentUserID)
	}
	return &BracketResponse{
		Tournament:      t,
		Matches:         matches,
		Participants:    ps,
		CurrentUserRole: role,
	}, nil
}

func (s *ChallongeService) GetStandings(ctx context.Context, tournamentID string) ([]bracket.Standing, error) {
	t, err := s.getTournament(ctx, tournamentID)
	if err != nil {
		return nil, err
	}
	if t.Status != entity.TournamentStatusCompleted {
		return nil, apperror.BadRequest("not_completed", "standings only available when tournament is completed", nil)
	}
	matches, err := s.bracketRepo.ListMatchesByTournament(ctx, tournamentID)
	if err != nil {
		return nil, apperror.Internal("list matches failed")
	}
	ps, err := s.participants.ListByTournament(ctx, tournamentID)
	if err != nil {
		return nil, apperror.Internal("list participants failed")
	}
	nodes := matchesToNodes(matches)
	outcomes := buildOutcomes(matches, ps)
	return bracket.ComputeStandings(nodes, outcomes, len(ps), t.Format), nil
}

func (s *ChallongeService) GetLog(ctx context.Context, tournamentID string) ([]*entity.MatchLogEntry, error) {
	return s.matchLog.ListByTournament(ctx, tournamentID)
}

type MatchWithTournament struct {
	entity.Match
	TournamentSlug  string `json:"tournament_slug"`
	TournamentTitle string `json:"tournament_title"`
}

func (s *ChallongeService) GetMyMatches(ctx context.Context, userID string) ([]*MatchWithTournament, error) {
	p, err := s.participants.GetByUserAndTournament(ctx, "", userID) // list all
	_ = p
	_ = err
	// Simplified: query matches where participant user_id matches.
	rows, err := s.pool.Query(ctx, `
		SELECT m.id, m.tournament_id, m.bracket_id, m.bracket_section, m.round_number, m.slot_index,
		       COALESCE(m.global_number,0),
		       m.team1_id, m.team2_id, m.participant1_id, m.participant2_id,
		       m.scheduled_at, m.location_or_server,
		       m.status, m.team1_confirmation_status, m.team2_confirmation_status,
		       m.winner_team_id, m.winner_participant_id, m.score_text, m.manager_comment,
		       m.next_match_id, m.source_match1_id, m.source_match2_id,
		       m.loser_next_match_id, m.loser_next_slot, m.is_bye,
		       m.created_at, m.updated_at, m.deleted_at,
		       t.slug, t.title
		FROM matches m
		JOIN participants p1 ON (m.participant1_id = p1.id OR m.participant2_id = p1.id)
		JOIN tournaments t ON t.id = m.tournament_id
		WHERE p1.user_id = $1 AND m.deleted_at IS NULL
		ORDER BY m.created_at DESC
	`, userID)
	if err != nil {
		return nil, apperror.Internal("query failed")
	}
	defer rows.Close()
	var out []*MatchWithTournament
	for rows.Next() {
		var mwt MatchWithTournament
		var slug *string
		m := &mwt.Match
		if err := rows.Scan(
			&m.ID, &m.TournamentID, &m.BracketID, &m.BracketSection, &m.RoundNumber, &m.SlotIndex,
			&m.GlobalNumber,
			&m.Team1ID, &m.Team2ID, &m.Participant1ID, &m.Participant2ID,
			&m.ScheduledAt, &m.LocationOrServer,
			&m.Status, &m.Team1ConfirmationStatus, &m.Team2ConfirmationStatus,
			&m.WinnerTeamID, &m.WinnerParticipantID, &m.ScoreText, &m.ManagerComment,
			&m.NextMatchID, &m.SourceMatch1ID, &m.SourceMatch2ID,
			&m.LoserNextMatchID, &m.LoserNextSlot, &m.IsBye,
			&m.CreatedAt, &m.UpdatedAt, &m.DeletedAt,
			&slug, &mwt.TournamentTitle,
		); err != nil {
			continue
		}
		if slug != nil {
			mwt.TournamentSlug = *slug
		}
		out = append(out, &mwt)
	}
	return out, rows.Err()
}

// ── Co-organizer management ───────────────────────────────────────────────────

type InviteCoOrgReq struct {
	Username string `json:"username"`
	Email    string `json:"email"`
}

func (s *ChallongeService) InviteCoOrganizer(ctx context.Context, tournamentID, actorID string, req InviteCoOrgReq) (*entity.CoOrganizerInvite, error) {
	var invitee *entity.User
	var err error
	if req.Username != "" {
		invitee, err = s.users.GetByNickname(ctx, req.Username)
	} else if req.Email != "" {
		invitee, err = s.users.GetByEmail(ctx, req.Email)
	} else {
		return nil, apperror.BadRequest("missing_field", "username or email required", nil)
	}
	if err != nil {
		return nil, apperror.NotFound("user not found")
	}
	token, err := generateInviteToken()
	if err != nil {
		return nil, apperror.Internal("token generation failed")
	}
	inv := &entity.CoOrganizerInvite{
		ID:           uuid.New().String(),
		TournamentID: tournamentID,
		InviteeID:    invitee.ID,
		InvitedByID:  actorID,
		Token:        token,
		Status:       entity.InviteStatusPending,
		ExpiresAt:    time.Now().Add(7 * 24 * time.Hour),
		CreatedAt:    time.Now(),
	}
	if err := s.invites.Create(ctx, inv); err != nil {
		return nil, apperror.Internal("invite creation failed")
	}
	t, _ := s.getTournament(ctx, tournamentID)
	s.hub.Broadcast(s.slugOf(t), ws.Event{Type: "co_organizer_invited", Payload: map[string]string{"invitee_id": invitee.ID}})
	return inv, nil
}

func (s *ChallongeService) AcceptInvite(ctx context.Context, token, userID string) error {
	_ = s.invites.MarkExpired(ctx) // sweep expired
	inv, err := s.invites.GetByToken(ctx, token)
	if err != nil {
		return apperror.NotFound("invite not found")
	}
	if inv.InviteeID != userID {
		return apperror.Forbidden("this invite is for a different user")
	}
	if inv.Status != entity.InviteStatusPending {
		return apperror.BadRequest("invite_not_pending", "invite is no longer pending", nil)
	}
	if time.Now().After(inv.ExpiresAt) {
		_ = s.invites.UpdateStatus(ctx, inv.ID, entity.InviteStatusExpired)
		return apperror.BadRequest("invite_expired", "invite has expired", nil)
	}
	_ = s.invites.UpdateStatus(ctx, inv.ID, entity.InviteStatusAccepted)
	_ = s.members.Upsert(ctx, &entity.TournamentMember{
		ID: uuid.New().String(), TournamentID: inv.TournamentID,
		UserID: userID, Role: entity.TournamentRoleCoOrganizer, JoinedAt: time.Now(),
	})
	t, _ := s.getTournament(ctx, inv.TournamentID)
	s.hub.Broadcast(s.slugOf(t), ws.Event{Type: "co_organizer_added", Payload: map[string]string{"user_id": userID}})
	return nil
}

func (s *ChallongeService) RemoveCoOrganizer(ctx context.Context, tournamentID, targetUserID, actorID string) error {
	return s.members.Delete(ctx, tournamentID, targetUserID)
}

// ── Join ──────────────────────────────────────────────────────────────────────

func (s *ChallongeService) Join(ctx context.Context, tournamentID, userID string) error {
	t, err := s.getTournamentGuard(ctx, tournamentID, entity.TournamentStatusDraft, entity.TournamentStatusReady, entity.TournamentStatusRegistrationOpen, entity.TournamentStatusRegistrationClosed)
	if err != nil {
		return err
	}
	if t.Visibility == "private" {
		return apperror.Forbidden("tournament is private")
	}
	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return apperror.NotFound("user not found")
	}
	name := user.Nickname
	if name == "" {
		name = user.Email
	}
	_, err = s.AddParticipant(ctx, tournamentID, userID, AddParticipantReq{Name: name, UserID: &userID})
	return err
}

// JoinPool adds a user to the participant pool for team-mode tournaments (registration_open).
func (s *ChallongeService) JoinPool(ctx context.Context, tournamentID, userID string) error {
	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return apperror.NotFound("user not found")
	}
	name := user.Nickname
	if name == "" {
		name = user.Email
	}
	existing, _ := s.participants.ListByTournament(ctx, tournamentID)
	for _, p := range existing {
		if p.UserID != nil && *p.UserID == userID {
			return apperror.BadRequest("already_joined", "вы уже в списке участников", nil)
		}
	}
	p := &entity.Participant{
		ID:           uuid.New().String(),
		TournamentID: tournamentID,
		UserID:       &userID,
		Name:         name,
		Seed:         len(existing) + 1,
		Status:       entity.ParticipantStatusActive,
		CreatedAt:    time.Now(),
	}
	if err := s.participants.Create(ctx, p); err != nil {
		return apperror.Internal("failed to join pool")
	}
	_ = s.members.Upsert(ctx, &entity.TournamentMember{
		ID: uuid.New().String(), TournamentID: tournamentID,
		UserID: userID, Role: entity.TournamentRoleParticipant, JoinedAt: time.Now(),
	})
	return nil
}

// ── Bracket generation ────────────────────────────────────────────────────────

func (s *ChallongeService) generateBracket(ctx context.Context, t *entity.Tournament, ps []*entity.Participant, actorID string) error {
	n := len(ps)
	var nodes []bracket.Node
	var err error
	switch t.Format {
	case "double_elimination":
		nodes, err = bracket.BuildDouble(n)
	default:
		nodes, err = bracket.BuildSingle(n)
	}
	if err != nil {
		return apperror.Internal("bracket build failed")
	}
	nums := bracket.GlobalNumbers(nodes)

	b := &entity.Bracket{
		ID:            uuid.New().String(),
		TournamentID:  t.ID,
		Format:        t.Format,
		SeedingMethod: "manual",
		Status:        "active",
		GeneratedBy:   actorID,
		GeneratedAt:   time.Now(),
		MetadataJSON:  json.RawMessage(`{}`),
	}
	// Delete old bracket if exists.
	_ = s.bracketRepo.DeleteByTournamentID(ctx, t.ID)
	if err := s.bracketRepo.CreateBracket(ctx, b); err != nil {
		return apperror.Internal("bracket create failed")
	}

	// Seed map: seed (1-based) → participant ID.
	seedToParticipant := make(map[int]string, len(ps))
	for _, p := range ps {
		seedToParticipant[p.Seed] = p.ID
	}

	matchIDs := make([]string, len(nodes))
	for i := range nodes {
		matchIDs[i] = uuid.New().String()
	}

	allMatches := make([]*entity.Match, len(nodes))
	for i, nd := range nodes {
		m := &entity.Match{
			ID:                      matchIDs[i],
			TournamentID:            t.ID,
			BracketID:               b.ID,
			BracketSection:          string(nd.Section),
			RoundNumber:             nd.Round,
			SlotIndex:               nd.Slot,
			GlobalNumber:            nums[nd.Index],
			Status:                  entity.MatchStatusScheduled,
			Team1ConfirmationStatus: entity.MatchTeamConfirmationPending,
			Team2ConfirmationStatus: entity.MatchTeamConfirmationPending,
			IsBye:                   nd.IsBye,
		}
		if nd.Seed1 > 0 {
			if pid, ok := seedToParticipant[nd.Seed1]; ok {
				m.Participant1ID = &pid
			}
		}
		if nd.Seed2 > 0 {
			if pid, ok := seedToParticipant[nd.Seed2]; ok {
				m.Participant2ID = &pid
			}
		}
		if nd.WinNext >= 0 {
			next := matchIDs[nd.WinNext]
			m.NextMatchID = &next
		}
		if nd.LoseNext >= 0 {
			next := matchIDs[nd.LoseNext]
			m.LoserNextMatchID = &next
			m.LoserNextSlot = nd.LoseSlot
		}
		allMatches[i] = m
	}

	// Phase 1: insert all matches without FK references (avoid self-referential FK violations).
	for _, m := range allMatches {
		nextID, loseNextID, loseSlot := m.NextMatchID, m.LoserNextMatchID, m.LoserNextSlot
		m.NextMatchID, m.LoserNextMatchID = nil, nil
		if err := s.bracketRepo.CreateMatch(ctx, m); err != nil {
			return apperror.Internal("match create failed: " + err.Error())
		}
		m.NextMatchID, m.LoserNextMatchID, m.LoserNextSlot = nextID, loseNextID, loseSlot
	}

	// Phase 2: update all matches with FK references.
	for _, m := range allMatches {
		if err := s.bracketRepo.UpdateMatchState(ctx, m); err != nil {
			return apperror.Internal("match update failed")
		}
	}

	// Auto-advance BYE matches.
	s.propagateByes(ctx, t.ID, nodes, matchIDs, seedToParticipant)
	return nil
}

func (s *ChallongeService) propagateByes(ctx context.Context, tournamentID string, nodes []bracket.Node, matchIDs []string, seedMap map[int]string) {
	for _, nd := range nodes {
		if !nd.IsBye {
			continue
		}
		m, err := s.bracketRepo.GetMatchByID(ctx, matchIDs[nd.Index])
		if err != nil {
			continue
		}
		// Exactly one side is empty (BYE).
		var winnerID *string
		if m.Participant1ID != nil {
			winnerID = m.Participant1ID
		} else {
			winnerID = m.Participant2ID
		}
		if winnerID == nil {
			continue
		}
		m.WinnerParticipantID = winnerID
		m.Status = entity.MatchStatusFinished
		_ = s.bracketRepo.UpdateMatchState(ctx, m)
		_ = s.advanceParticipant(ctx, m, *winnerID, "")
	}
}

// advanceParticipant places winner into the next match and loser into LB (or eliminates).
func (s *ChallongeService) advanceParticipant(ctx context.Context, m *entity.Match, winnerID, loserID string) error {
	// Advance winner.
	if m.NextMatchID != nil {
		next, err := s.bracketRepo.GetMatchByID(ctx, *m.NextMatchID)
		if err != nil {
			return nil
		}
		wid := winnerID
		if next.Participant1ID == nil {
			next.Participant1ID = &wid
		} else {
			next.Participant2ID = &wid
		}
		if next.Participant1ID != nil && next.Participant2ID != nil {
			next.Status = entity.MatchStatusScheduled
		}
		_ = s.bracketRepo.UpdateMatchState(ctx, next)
	} else {
		// Winner has no next match → champion.
		_ = s.participants.UpdateStatus(ctx, winnerID, entity.ParticipantStatusChampion, intPtr(1))
		_ = s.finishTournament(ctx, m.TournamentID, winnerID)
	}

	// Advance loser to LB or eliminate.
	if loserID == "" {
		return nil
	}
	if m.LoserNextMatchID != nil {
		lnext, err := s.bracketRepo.GetMatchByID(ctx, *m.LoserNextMatchID)
		if err != nil {
			return nil
		}
		lid := loserID
		if m.LoserNextSlot == 1 {
			lnext.Participant1ID = &lid
		} else {
			lnext.Participant2ID = &lid
		}
		if lnext.Participant1ID != nil && lnext.Participant2ID != nil {
			lnext.Status = entity.MatchStatusScheduled
		}
		_ = s.bracketRepo.UpdateMatchState(ctx, lnext)
	} else {
		// No LB route → eliminated.
		_ = s.participants.UpdateStatus(ctx, loserID, entity.ParticipantStatusEliminated, nil)
	}
	// Handle GF Reset: if this is GF and loser (LB champ) won (i.e. winnerID is LB champ).
	if m.BracketSection == entity.BracketSectionGF {
		s.handleGFResult(ctx, m, winnerID, loserID)
	}
	return nil
}

func (s *ChallongeService) handleGFResult(ctx context.Context, gf *entity.Match, winnerID, loserID string) {
	// If WB champion (participant1) lost → LB champion won → need GF Reset.
	if winnerID == ptrVal(gf.Participant2ID) {
		// Create GF Reset match.
		reset := &entity.Match{
			ID:                      uuid.New().String(),
			TournamentID:            gf.TournamentID,
			BracketID:               gf.BracketID,
			BracketSection:          entity.BracketSectionGF,
			RoundNumber:             2, // GF Reset = round 2
			SlotIndex:               1,
			Status:                  entity.MatchStatusScheduled,
			Team1ConfirmationStatus: entity.MatchTeamConfirmationPending,
			Team2ConfirmationStatus: entity.MatchTeamConfirmationPending,
			Participant1ID:          gf.Participant2ID, // LB champ (winner of GF1) → slot 1
			Participant2ID:          gf.Participant1ID, // WB champ (loser of GF1) → slot 2
		}
		_ = s.bracketRepo.CreateMatch(ctx, reset)
		// Reset participants out of champion status.
		_ = s.participants.UpdateStatus(ctx, winnerID, entity.ParticipantStatusActive, nil)
	}
}

func (s *ChallongeService) finishTournament(ctx context.Context, tournamentID, championID string) error {
	_ = s.tournaments.UpdateStatus(ctx, tournamentID, entity.TournamentStatusCompleted)
	t, _ := s.getTournament(ctx, tournamentID)
	s.hub.Broadcast(s.slugOf(t), ws.Event{Type: "tournament_completed", Payload: map[string]string{"champion_id": championID}})
	return nil
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func (s *ChallongeService) getTournament(ctx context.Context, id string) (*entity.Tournament, error) {
	t, err := s.tournaments.GetByID(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apperror.NotFound("tournament not found")
	}
	return t, err
}

func (s *ChallongeService) getTournamentGuard(ctx context.Context, id string, allowed ...string) (*entity.Tournament, error) {
	t, err := s.getTournament(ctx, id)
	if err != nil {
		return nil, err
	}
	for _, s := range allowed {
		if t.Status == s {
			return t, nil
		}
	}
	return nil, apperror.BadRequest("invalid_status", fmt.Sprintf("action not allowed in status %s", t.Status), nil)
}

func (s *ChallongeService) maybeSetReady(ctx context.Context, tournamentID string) {
	count, err := s.participants.Count(ctx, tournamentID)
	if err != nil {
		return
	}
	t, err := s.getTournament(ctx, tournamentID)
	if err != nil {
		return
	}
	if count >= 2 && t.Status == entity.TournamentStatusDraft {
		_ = s.tournaments.UpdateStatus(ctx, tournamentID, entity.TournamentStatusReady)
	} else if count < 2 && t.Status == entity.TournamentStatusReady {
		_ = s.tournaments.UpdateStatus(ctx, tournamentID, entity.TournamentStatusDraft)
	}
}

func (s *ChallongeService) normalizeSeedOrder(ctx context.Context, tournamentID string) {
	ps, err := s.participants.ListByTournament(ctx, tournamentID)
	if err != nil {
		return
	}
	for i, p := range ps {
		_ = s.participants.UpdateSeed(ctx, p.ID, i+1)
	}
}

func (s *ChallongeService) logMatchAction(ctx context.Context, tournamentID, matchID, actorID, action string, detail interface{}) {
	var raw json.RawMessage
	if detail != nil {
		b, _ := json.Marshal(detail)
		raw = b
	}
	entry := &entity.MatchLogEntry{
		ID:           uuid.New().String(),
		TournamentID: tournamentID,
		MatchID:      matchID,
		Action:       action,
		ActorID:      actorID,
		Detail:       raw,
		CreatedAt:    time.Now(),
	}
	_ = s.matchLog.Create(ctx, entry)
}

// GetMyRole returns the role of userID in the given tournament (empty string if none).
func (s *ChallongeService) GetMyRole(ctx context.Context, tournamentID, userID string) string {
	role, _ := s.members.GetRole(ctx, tournamentID, userID)
	return role
}

func (s *ChallongeService) slugOf(t *entity.Tournament) string {
	if t != nil && t.Slug != nil {
		return *t.Slug
	}
	return t.ID
}

// matchesToNodes converts persisted matches back to bracket.Node for CascadeReset.
// Only WinNext/LoseNext links are needed; Index = GlobalNumber-1.
func matchesToNodes(matches []entity.Match) []bracket.Node {
	byID := make(map[string]int, len(matches))
	for _, m := range matches {
		byID[m.ID] = m.GlobalNumber - 1
	}
	nodes := make([]bracket.Node, len(matches))
	for i, m := range matches {
		nodes[i] = bracket.Node{
			Index:    m.GlobalNumber - 1,
			Section:  bracket.Section(m.BracketSection),
			Round:    m.RoundNumber,
			Slot:     m.SlotIndex,
			WinNext:  -1,
			LoseNext: -1,
		}
		if m.NextMatchID != nil {
			if idx, ok := byID[*m.NextMatchID]; ok {
				nodes[i].WinNext = idx
			}
		}
		if m.LoserNextMatchID != nil {
			if idx, ok := byID[*m.LoserNextMatchID]; ok {
				nodes[i].LoseNext = idx
				nodes[i].LoseSlot = m.LoserNextSlot
			}
		}
	}
	return nodes
}

func matchIndexOf(nodes []bracket.Node, matchID string) int {
	// matchID is not in Node; use GlobalNumber convention (Index = GlobalNumber-1).
	// Caller must map matchID to a GlobalNumber before calling. Here we return 0 as fallback.
	_ = matchID
	return 0
}

func buildOutcomes(matches []entity.Match, ps []*entity.Participant) []bracket.MatchOutcome {
	pidToSeed := make(map[string]int, len(ps))
	for _, p := range ps {
		pidToSeed[p.ID] = p.Seed
	}
	var outcomes []bracket.MatchOutcome
	for _, m := range matches {
		if m.WinnerParticipantID == nil {
			continue
		}
		var winSeed, loseSeed int
		if m.Participant1ID != nil {
			winSeed = pidToSeed[*m.Participant1ID]
		}
		if m.Participant2ID != nil {
			loseSeed = pidToSeed[*m.Participant2ID]
		}
		if ptrVal(m.WinnerParticipantID) == ptrVal(m.Participant2ID) {
			winSeed, loseSeed = loseSeed, winSeed
		}
		outcomes = append(outcomes, bracket.MatchOutcome{
			NodeIndex:  m.GlobalNumber - 1,
			WinnerSeed: winSeed,
			LoserSeed:  loseSeed,
		})
	}
	return outcomes
}

func ptrVal(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func intPtr(i int) *int { return &i }
