package service

import (
	"context"

	"esports-backend/internal/repository"
)

type AdminService struct {
	users       *repository.UserRepository
	tournaments *repository.TournamentRepository
}

func NewAdminService(users *repository.UserRepository, tournaments *repository.TournamentRepository) *AdminService {
	return &AdminService{users: users, tournaments: tournaments}
}

func (s *AdminService) ListUsers(ctx context.Context, limit, offset int) (interface{}, error) {
	return s.users.ListUsers(ctx, limit, offset)
}

func (s *AdminService) BlockUser(ctx context.Context, userID string) error {
	return s.users.SetBlocked(ctx, userID, true)
}

func (s *AdminService) UnblockUser(ctx context.Context, userID string) error {
	return s.users.SetBlocked(ctx, userID, false)
}

func (s *AdminService) ListTournaments(ctx context.Context, limit, offset int) (interface{}, error) {
	return s.tournaments.ListAll(ctx, limit, offset)
}
