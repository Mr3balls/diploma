package service

import (
	"context"
	"fmt"
	"time"

	"esports-backend/internal/apperror"
	"esports-backend/internal/config"
	"esports-backend/internal/entity"
	"esports-backend/internal/pkg/password"
	tok "esports-backend/internal/pkg/tokens"
	"esports-backend/internal/pkg/xjson"
	"esports-backend/internal/repository"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type AuthService struct {
	cfg           *config.Config
	users         repository.UserStore
	sessions      repository.SessionStore
	audit         repository.AuditStore
	email         *EmailService
	teams         *repository.TeamRepository
	notifications *repository.NotificationRepository
	resetTokens   *repository.PasswordResetRepository
}

func NewAuthService(cfg *config.Config, users repository.UserStore, sessions repository.SessionStore, audit repository.AuditStore, email *EmailService, teams *repository.TeamRepository, notifications *repository.NotificationRepository, resetTokens *repository.PasswordResetRepository) *AuthService {
	return &AuthService{cfg: cfg, users: users, sessions: sessions, audit: audit, email: email, teams: teams, notifications: notifications, resetTokens: resetTokens}
}

type RegisterInput struct {
	FirstName string
	LastName  string
	Email     string
	Phone     string
	Nickname  string
	Password  string
}

type LoginInput struct {
	Email     string
	Password  string
	UserAgent *string
	IPAddress *string
}

func (s *AuthService) Register(ctx context.Context, in RegisterInput, userAgent, ipAddress *string) (*entity.User, *entity.TokenPair, error) {
	if _, err := s.users.GetByEmail(ctx, in.Email); err == nil {
		return nil, nil, apperror.Conflict("email already exists")
	} else if err != pgx.ErrNoRows {
		return nil, nil, err
	}

	if _, err := s.users.GetByNickname(ctx, in.Nickname); err == nil {
		return nil, nil, apperror.Conflict("nickname already exists")
	} else if err != pgx.ErrNoRows {
		return nil, nil, err
	}

	hash, err := password.Hash(in.Password)
	if err != nil {
		return nil, nil, err
	}

	user := &entity.User{
		ID:           uuid.NewString(),
		FirstName:    in.FirstName,
		LastName:     in.LastName,
		Email:        in.Email,
		Phone:        in.Phone,
		Nickname:     in.Nickname,
		PasswordHash: hash,
		IsBlocked:    false,
	}
	if err := s.users.Create(ctx, user); err != nil {
		return nil, nil, err
	}
	if err := s.users.AssignPlatformRole(ctx, user.ID, entity.PlatformRolePlayer); err != nil {
		return nil, nil, err
	}

	// Link any pending team memberships created before this user registered
	if s.teams != nil && s.notifications != nil {
		if pending, err := s.teams.FindPendingByEmail(ctx, user.Email); err == nil {
			for _, m := range pending {
				_ = s.teams.SetMemberUserID(ctx, m.ID, user.ID)
				payload := map[string]string{"team_id": m.TeamID, "team_member_id": m.ID}
				_ = s.notifications.Create(ctx, &entity.Notification{
					ID:                uuid.NewString(),
					UserID:            user.ID,
					Type:              entity.NotificationAddedToTeam,
					Title:             "Вас добавили в команду",
					Message:           fmt.Sprintf("Вас пригласили в команду. Подтвердите участие."),
					PayloadJSON:       xjson.MustMarshal(payload),
					ActionPayloadJSON: xjson.MustMarshal(payload),
				})
			}
		}
	}
	roles, err := s.users.GetPlatformRoles(ctx, user.ID)
	if err != nil {
		return nil, nil, err
	}
	user.Roles = roles

	pair, err := s.issueTokens(ctx, user, userAgent, ipAddress)
	if err != nil {
		return nil, nil, err
	}
	return user, pair, nil
}

func (s *AuthService) Login(ctx context.Context, in LoginInput) (*entity.User, *entity.TokenPair, error) {
	user, err := s.users.GetByEmail(ctx, in.Email)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil, apperror.Unauthorized("invalid email or password")
		}
		return nil, nil, err
	}
	if user.IsBlocked {
		return nil, nil, apperror.Forbidden("user is blocked")
	}
	if err := password.Compare(user.PasswordHash, in.Password); err != nil {
		return nil, nil, apperror.Unauthorized("invalid email or password")
	}
	roles, err := s.users.GetPlatformRoles(ctx, user.ID)
	if err != nil {
		return nil, nil, err
	}
	user.Roles = roles
	pair, err := s.issueTokens(ctx, user, in.UserAgent, in.IPAddress)
	if err != nil {
		return nil, nil, err
	}
	return user, pair, nil
}

func (s *AuthService) Refresh(ctx context.Context, rawRefreshToken string, userAgent, ipAddress *string) (*entity.TokenPair, error) {
	session, err := s.sessions.GetActiveByHash(ctx, tok.HashRefreshToken(rawRefreshToken))
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, apperror.Unauthorized("invalid refresh token")
		}
		return nil, err
	}
	user, err := s.users.GetByID(ctx, session.UserID)
	if err != nil {
		return nil, err
	}
	roles, err := s.users.GetPlatformRoles(ctx, user.ID)
	if err != nil {
		return nil, err
	}
	user.Roles = roles
	if err := s.sessions.RevokeByHash(ctx, session.RefreshTokenHash); err != nil {
		return nil, err
	}
	return s.issueTokens(ctx, user, userAgent, ipAddress)
}

func (s *AuthService) Logout(ctx context.Context, rawRefreshToken string) error {
	return s.sessions.RevokeByHash(ctx, tok.HashRefreshToken(rawRefreshToken))
}

func (s *AuthService) ForgotPassword(ctx context.Context, userEmail string) map[string]string {
	// Always return the same message to prevent email enumeration.
	resp := map[string]string{"message": "Если аккаунт с таким email существует, письмо со сбросом пароля отправлено."}

	user, err := s.users.GetByEmail(ctx, userEmail)
	if err != nil {
		return resp
	}

	rawToken, tokenHash, err := tok.GenerateOpaqueRefreshToken()
	if err != nil {
		return resp
	}

	resetToken := &entity.PasswordResetToken{
		ID:        uuid.NewString(),
		UserID:    user.ID,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().Add(30 * time.Minute),
	}
	if err := s.resetTokens.Create(ctx, resetToken); err != nil {
		return resp
	}

	resetURL := fmt.Sprintf("%s/reset-password?token=%s", s.cfg.FrontendURL, rawToken)
	go s.email.SendPasswordReset(userEmail, resetURL)
	return resp
}

func (s *AuthService) ResetPassword(ctx context.Context, rawToken, newPassword string) error {
	if len(newPassword) < 8 {
		return apperror.BadRequest("weak_password", "password must be at least 8 characters", nil)
	}

	tokenHash := tok.HashRefreshToken(rawToken)
	resetToken, err := s.resetTokens.GetActiveByHash(ctx, tokenHash)
	if err != nil {
		return apperror.BadRequest("invalid_token", "token is invalid or expired", nil)
	}

	hash, err := password.Hash(newPassword)
	if err != nil {
		return err
	}
	if err := s.users.UpdatePassword(ctx, resetToken.UserID, hash); err != nil {
		return err
	}
	if err := s.resetTokens.MarkUsed(ctx, resetToken.ID); err != nil {
		return err
	}
	// Revoke all active sessions so the old password can't be reused via refresh tokens.
	_ = s.sessions.RevokeByUserID(ctx, resetToken.UserID)
	return nil
}

func (s *AuthService) issueTokens(ctx context.Context, user *entity.User, userAgent, ipAddress *string) (*entity.TokenPair, error) {
	accessToken, expiresIn, err := tok.GenerateAccessToken(s.cfg.AccessTokenSecret, s.cfg.AccessTokenTTL, user.ID, user.Roles)
	if err != nil {
		return nil, err
	}
	rawRefresh, refreshHash, err := tok.GenerateOpaqueRefreshToken()
	if err != nil {
		return nil, err
	}
	session := &entity.AuthSession{
		ID:               uuid.NewString(),
		UserID:           user.ID,
		RefreshTokenHash: refreshHash,
		UserAgent:        userAgent,
		IPAddress:        ipAddress,
		ExpiresAt:        time.Now().Add(s.cfg.RefreshTokenTTL),
	}
	if err := s.sessions.Create(ctx, session); err != nil {
		return nil, err
	}
	return &entity.TokenPair{AccessToken: accessToken, RefreshToken: rawRefresh, ExpiresIn: expiresIn}, nil
}
