package service

import (
	"context"

	"esports-backend/internal/apperror"
	"esports-backend/internal/repository"
)

type AuditService struct {
	tournaments *TournamentService
	audits      *repository.AuditRepository
}

func NewAuditService(tournaments *TournamentService, audits *repository.AuditRepository) *AuditService {
	return &AuditService{tournaments: tournaments, audits: audits}
}

func (s *AuditService) ListByTournament(ctx context.Context, actorUserID, tournamentID string) (interface{}, error) {
	ok, err := s.tournaments.CanManageTournament(ctx, tournamentID, actorUserID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, apperror.Forbidden("insufficient tournament permissions")
	}
	return s.audits.ListByTournament(ctx, tournamentID)
}
