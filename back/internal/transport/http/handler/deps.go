package handler

import (
	"esports-backend/internal/service"
	ws "esports-backend/internal/transport/websocket"

	"github.com/go-playground/validator/v10"
)

type Deps struct {
	Validate      *validator.Validate
	Auth          *service.AuthService
	Users         *service.UserService
	Tournaments   *service.TournamentService
	Imports       *service.ImportService
	Teams         *service.TeamService
	Brackets      *service.BracketService
	Matches       *service.MatchService
	Notifications *service.NotificationService
	Audits        *service.AuditService
	Admin         *service.AdminService
	Challonge     *service.ChallongeService
	Chat          *service.ChatService
	Hub           *ws.Hub
	JWTSecret     string
}
