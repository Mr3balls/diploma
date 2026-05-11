package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"time"

	"esports-backend/internal/config"
	"esports-backend/internal/database"
	googleclient "esports-backend/internal/integration/google"
	"esports-backend/internal/repository"
	"esports-backend/internal/service"
	httptransport "esports-backend/internal/transport/http"
	"esports-backend/internal/transport/http/handler"
	ws "esports-backend/internal/transport/websocket"

	"github.com/go-playground/validator/v10"
)

var phoneRuRegex = regexp.MustCompile(`^(\+7|8)\d{10}$`)

type disabledSheetsClient struct{}

func (disabledSheetsClient) ReadRows(_ context.Context, _ string, _ string) ([][]string, error) {
	return nil, fmt.Errorf("google sheets client is not configured")
}

func main() {
	ctx := context.Background()
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config load failed: %v", err)
	}
	pg, err := database.NewPostgres(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("postgres connection failed: %v", err)
	}
	defer pg.Close()

	redisClient, err := database.NewRedis(ctx, cfg.RedisURL)
	if err != nil {
		log.Fatalf("redis connection failed: %v", err)
	}
	defer redisClient.Close()

	var sheetsReader service.SheetsReader = disabledSheetsClient{}
	sheetsClient, err := googleclient.NewSheetsClient(ctx, cfg.GoogleServiceAccountFile)
	if err != nil {
		log.Printf("google sheets client disabled: %v", err)
	} else {
		sheetsReader = sheetsClient
	}

	userRepo := repository.NewUserRepository(pg)
	sessionRepo := repository.NewSessionRepository(pg)
	tournamentRepo := repository.NewTournamentRepository(pg)
	importRepo := repository.NewImportRepository(pg)
	teamRepo := repository.NewTeamRepository(pg)
	bracketRepo := repository.NewBracketRepository(pg)
	notificationRepo := repository.NewNotificationRepository(pg)
	auditRepo := repository.NewAuditRepository(pg)

	// Challonge-style repositories
	participantRepo := repository.NewParticipantRepository(pg)
	membersRepo := repository.NewTournamentMemberRepository(pg)
	reportsRepo := repository.NewResultReportRepository(pg)
	matchLogRepo := repository.NewMatchLogRepository(pg)
	invitesRepo := repository.NewCoOrganizerInviteRepository(pg)

	tournamentService := service.NewTournamentService(tournamentRepo, teamRepo, bracketRepo, auditRepo)
	authService := service.NewAuthService(cfg, userRepo, sessionRepo, auditRepo)
	userService := service.NewUserService(userRepo)
	importService := service.NewImportService(tournamentService, importRepo, teamRepo, userRepo, notificationRepo, auditRepo, sheetsReader)
	notificationService := service.NewNotificationService(notificationRepo)
	bracketService := service.NewBracketService(pg, tournamentService, bracketRepo, teamRepo, notificationRepo, auditRepo)
	matchService := service.NewMatchService(tournamentService, bracketRepo, teamRepo, notificationRepo, auditRepo, bracketService)
	teamService := service.NewTeamService(tournamentService, teamRepo, userRepo, notificationRepo, auditRepo)
	auditService := service.NewAuditService(tournamentService, auditRepo)
	adminService := service.NewAdminService(userRepo, tournamentRepo)

	// Challonge SSE hub + service
	hub := ws.NewHub()
	challongeService := service.NewChallongeService(pg, tournamentRepo, bracketRepo, participantRepo, membersRepo, reportsRepo, matchLogRepo, invitesRepo, userRepo, hub)

	validate := validator.New()
	_ = validate.RegisterValidation("phone_ru", func(fl validator.FieldLevel) bool {
		return phoneRuRegex.MatchString(fl.Field().String())
	})

	deps := handler.Deps{
		Validate:      validate,
		Auth:          authService,
		Users:         userService,
		Tournaments:   tournamentService,
		Imports:       importService,
		Teams:         teamService,
		Brackets:      bracketService,
		Matches:       matchService,
		Notifications: notificationService,
		Audits:        auditService,
		Admin:         adminService,
		Challonge:     challongeService,
		Hub:           hub,
	}

	srv := &http.Server{
		Addr:              ":" + cfg.HTTPPort,
		Handler:           httptransport.NewRouter(cfg, deps),
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       20 * time.Second,
		WriteTimeout:      20 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	log.Printf("server listening on :%s", cfg.HTTPPort)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
