package handler

import (
	"esports-backend/internal/service"

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
}
