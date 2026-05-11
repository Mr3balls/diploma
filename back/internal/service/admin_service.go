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

type PagedResult[T any] struct {
	Items []T
	Total int
}

func (s *AdminService) ListUsers(ctx context.Context, limit, offset int) (*PagedResult[interface{}], error) {
	items, err := s.users.ListUsers(ctx, limit, offset)
	if err != nil {
		return nil, err
	}
	total, err := s.users.CountUsers(ctx)
	if err != nil {
		return nil, err
	}
	converted := make([]interface{}, len(items))
	for i, v := range items {
		converted[i] = v
	}
	return &PagedResult[interface{}]{Items: converted, Total: total}, nil
}

func (s *AdminService) BlockUser(ctx context.Context, userID string) error {
	return s.users.SetBlocked(ctx, userID, true)
}

func (s *AdminService) UnblockUser(ctx context.Context, userID string) error {
	return s.users.SetBlocked(ctx, userID, false)
}

func (s *AdminService) ListTournaments(ctx context.Context, limit, offset int) (*PagedResult[interface{}], error) {
	items, err := s.tournaments.ListAll(ctx, limit, offset)
	if err != nil {
		return nil, err
	}
	total, err := s.tournaments.CountAll(ctx)
	if err != nil {
		return nil, err
	}
	converted := make([]interface{}, len(items))
	for i, v := range items {
		converted[i] = v
	}
	return &PagedResult[interface{}]{Items: converted, Total: total}, nil
}
