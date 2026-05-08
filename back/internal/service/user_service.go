package service

import (
	"context"

	"esports-backend/internal/apperror"
	"esports-backend/internal/entity"
	"esports-backend/internal/repository"
)

type UserService struct {
	users *repository.UserRepository
}

func NewUserService(users *repository.UserRepository) *UserService {
	return &UserService{users: users}
}

type UpdateProfileInput struct {
	FirstName string
	LastName  string
	Phone     string
	AvatarURL *string
}

func (s *UserService) GetMe(ctx context.Context, userID string) (*entity.User, error) {
	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return nil, apperror.NotFound("user not found")
	}
	roles, err := s.users.GetPlatformRoles(ctx, userID)
	if err != nil {
		return nil, err
	}
	user.Roles = roles
	return user, nil
}

func (s *UserService) UpdateMe(ctx context.Context, userID string, in UpdateProfileInput) (*entity.User, error) {
	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return nil, apperror.NotFound("user not found")
	}
	user.FirstName = in.FirstName
	user.LastName = in.LastName
	user.Phone = in.Phone
	user.AvatarURL = in.AvatarURL
	if err := s.users.UpdateProfile(ctx, user); err != nil {
		return nil, err
	}
	return s.GetMe(ctx, userID)
}

func (s *UserService) DeleteMe(ctx context.Context, userID string) error {
	return s.users.SoftDelete(ctx, userID)
}
