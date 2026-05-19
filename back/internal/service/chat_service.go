package service

import (
	"context"
	"strings"
	"time"

	"esports-backend/internal/apperror"
	"esports-backend/internal/entity"
	"esports-backend/internal/repository"

	"github.com/google/uuid"
)

type ChatBroadcaster interface {
	BroadcastToTournament(tournamentID string, msg *entity.TournamentMessage)
}

type ChatService struct {
	chat        *repository.ChatRepository
	tournaments repository.TournamentStore
	broadcaster ChatBroadcaster
}

func NewChatService(chat *repository.ChatRepository, tournaments repository.TournamentStore) *ChatService {
	return &ChatService{chat: chat, tournaments: tournaments}
}

func (s *ChatService) WithBroadcaster(b ChatBroadcaster) {
	s.broadcaster = b
}

// canAccess checks that the user can see the tournament (public or has a role).
func (s *ChatService) canAccess(ctx context.Context, tournamentID, userID string) error {
	t, err := s.tournaments.GetByID(ctx, tournamentID)
	if err != nil {
		return apperror.NotFound("tournament not found")
	}
	if t.Visibility == entity.TournamentVisibilityPrivate {
		ok, err := s.tournaments.HasRole(ctx, tournamentID, userID, entity.TournamentRoleOwner)
		if err != nil {
			return err
		}
		if !ok {
			ok2, _ := s.tournaments.HasRole(ctx, tournamentID, userID, entity.TournamentRoleManager)
			if !ok2 {
				return apperror.Forbidden("tournament is private")
			}
		}
	}
	return nil
}

func (s *ChatService) GetMessages(ctx context.Context, tournamentID, userID string, limit int, before time.Time) ([]*entity.TournamentMessage, error) {
	if err := s.canAccess(ctx, tournamentID, userID); err != nil {
		return nil, err
	}
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	msgs, err := s.chat.List(ctx, tournamentID, limit, before)
	if err != nil {
		return nil, err
	}
	// Reverse so client gets oldest-first order
	for i, j := 0, len(msgs)-1; i < j; i, j = i+1, j-1 {
		msgs[i], msgs[j] = msgs[j], msgs[i]
	}
	return msgs, nil
}

func (s *ChatService) SendMessage(ctx context.Context, tournamentID, userID, content string) (*entity.TournamentMessage, error) {
	if err := s.canAccess(ctx, tournamentID, userID); err != nil {
		return nil, err
	}
	content = strings.TrimSpace(content)
	if content == "" {
		return nil, apperror.BadRequest("empty_content", "message cannot be empty", nil)
	}
	if len([]rune(content)) > 1000 {
		return nil, apperror.BadRequest("content_too_long", "message must be 1000 characters or fewer", nil)
	}
	msg := &entity.TournamentMessage{
		ID:           uuid.NewString(),
		TournamentID: tournamentID,
		UserID:       userID,
		Content:      content,
		CreatedAt:    time.Now().UTC(),
	}
	if err := s.chat.Create(ctx, msg); err != nil {
		return nil, err
	}
	// Re-fetch to populate user_nickname via JOIN
	if full, err := s.chat.GetByID(ctx, msg.ID); err == nil {
		msg = full
	}
	if s.broadcaster != nil {
		s.broadcaster.BroadcastToTournament(tournamentID, msg)
	}
	return msg, nil
}
