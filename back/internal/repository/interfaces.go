package repository

import (
	"context"

	"esports-backend/internal/entity"
)

type TournamentStore interface {
	Create(ctx context.Context, t *entity.Tournament) error
	Update(ctx context.Context, t *entity.Tournament) error
	SetStatus(ctx context.Context, id, status string) error
	SetWinner(ctx context.Context, tournamentID string, winnerTeamID, winnerParticipantID *string) error
	SoftDelete(ctx context.Context, id string) error
	GetByID(ctx context.Context, id string) (*entity.Tournament, error)
	GetBySlug(ctx context.Context, slug string) (*entity.Tournament, error)
	ListPublic(ctx context.Context, limit, offset int, f TournamentFilter) ([]entity.Tournament, error)
	CountPublic(ctx context.Context, f TournamentFilter) (int, error)
	ListAll(ctx context.Context, limit, offset int) ([]entity.Tournament, error)
	CountAll(ctx context.Context) (int, error)
	AddRole(ctx context.Context, role *entity.TournamentUserRole) error
	RemoveRole(ctx context.Context, tournamentID, userID, role string) error
	HasRole(ctx context.Context, tournamentID, userID, role string) (bool, error)
	ListRoles(ctx context.Context, tournamentID, userID string) ([]string, error)
	ListUserIDsByRoles(ctx context.Context, tournamentID string, roles []string) ([]string, error)
}

type TeamStore interface {
	ListByTournament(ctx context.Context, tournamentID string, admin bool) ([]entity.Team, error)
}

type BracketStore interface {
	ListMatchesByTournament(ctx context.Context, tournamentID string) ([]entity.Match, error)
}

type AuditStore interface {
	Create(ctx context.Context, log *entity.AuditLog) error
}

type UserStore interface {
	Create(ctx context.Context, u *entity.User) error
	GetByEmail(ctx context.Context, email string) (*entity.User, error)
	GetByNickname(ctx context.Context, nickname string) (*entity.User, error)
	GetByID(ctx context.Context, id string) (*entity.User, error)
	UpdateProfile(ctx context.Context, u *entity.User) error
	SoftDelete(ctx context.Context, id string) error
	AssignPlatformRole(ctx context.Context, userID, roleCode string) error
	GetPlatformRoles(ctx context.Context, userID string) ([]string, error)
	// GetLangsByIDs returns a map of userID → lang for the given IDs (default "ru" for unknown).
	GetLangsByIDs(ctx context.Context, ids []string) map[string]string
}

type SessionStore interface {
	Create(ctx context.Context, s *entity.AuthSession) error
	GetActiveByHash(ctx context.Context, hash string) (*entity.AuthSession, error)
	RevokeByHash(ctx context.Context, hash string) error
}
