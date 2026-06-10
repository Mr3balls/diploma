package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"syscall"
	"time"

	"esports-backend/internal/config"
	"esports-backend/internal/database"
	googleclient "esports-backend/internal/integration/google"
	"esports-backend/internal/pkg/email"
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
	groupRepo := repository.NewGroupRepository(pg)
	notificationRepo := repository.NewNotificationRepository(pg)
	pushSubRepo := repository.NewPushSubscriptionRepository(pg)
	auditRepo := repository.NewAuditRepository(pg)
	chatRepo := repository.NewChatRepository(pg)

	// Challonge-style repositories
	participantRepo := repository.NewParticipantRepository(pg)
	membersRepo := repository.NewTournamentMemberRepository(pg)
	reportsRepo := repository.NewResultReportRepository(pg)
	matchLogRepo := repository.NewMatchLogRepository(pg)
	invitesRepo := repository.NewCoOrganizerInviteRepository(pg)

	var emailSender *email.Sender
	if cfg.BrevoAPIKey != "" {
		emailSender = email.NewSender(cfg.BrevoAPIKey, cfg.EmailFromAddress, cfg.EmailFromName)
		log.Printf("email notifications enabled (Brevo, from: %s <%s>)", cfg.EmailFromName, cfg.EmailFromAddress)
	} else {
		log.Println("email notifications disabled (BREVO_API_KEY not configured)")
	}
	emailService := service.NewEmailService(emailSender)

	passwordResetRepo := repository.NewPasswordResetRepository(pg)

	tournamentService := service.NewTournamentService(tournamentRepo, teamRepo, bracketRepo, auditRepo)
	authService := service.NewAuthService(cfg, userRepo, sessionRepo, auditRepo, emailService, teamRepo, notificationRepo, passwordResetRepo)
	userService := service.NewUserService(userRepo)
	importService := service.NewImportService(tournamentService, importRepo, teamRepo, userRepo, notificationRepo, auditRepo, sheetsReader, emailService)
	pushService := service.NewPushService(pushSubRepo, cfg.VAPIDPrivateKey, cfg.VAPIDPublicKey, cfg.VAPIDEmail)
	notificationService := service.NewNotificationService(notificationRepo, userRepo, pushService)
	bracketService := service.NewBracketService(pg, tournamentService, bracketRepo, groupRepo, teamRepo, participantRepo, userRepo, notificationRepo, auditRepo)
	matchService := service.NewMatchService(tournamentService, bracketRepo, groupRepo, teamRepo, participantRepo, userRepo, notificationRepo, auditRepo, bracketService, emailService)
	teamService := service.NewTeamService(tournamentService, teamRepo, userRepo, notificationRepo, auditRepo, emailService)
	auditService := service.NewAuditService(tournamentService, auditRepo)
	adminService := service.NewAdminService(userRepo, tournamentRepo)

	// SSE hub (shared for tournament brackets, user notifications, and chat)
	hub := ws.NewHub()
	notificationRepo.WithBroadcaster(hub)
	notificationRepo.WithPushSender(pushService)

	chatService := service.NewChatService(chatRepo, tournamentRepo)
	chatService.WithBroadcaster(hub)
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
		Push:          pushService,
		Audits:        auditService,
		Admin:         adminService,
		Challonge:     challongeService,
		Chat:          chatService,
		Hub:           hub,
		JWTSecret:     cfg.AccessTokenSecret,
	}

	srv := &http.Server{
		Addr:              ":" + cfg.HTTPPort,
		Handler:           httptransport.NewRouter(cfg, deps, redisClient),
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       20 * time.Second,
		WriteTimeout:      20 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("server listening on :%s", cfg.HTTPPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	<-quit
	log.Println("shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("forced shutdown: %v", err)
	}
	log.Println("server stopped")
}
