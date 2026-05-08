package httptransport

import (
	"net/http"
	"strings"

	"esports-backend/internal/config"
	"esports-backend/internal/transport/http/handler"
	mw "esports-backend/internal/transport/http/middleware"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
)

func NewRouter(cfg *config.Config, deps handler.Deps) http.Handler {
	r := chi.NewRouter()

	r.Use(mw.RequestID)
	r.Use(mw.Recoverer)
	r.Use(mw.Logging)
	r.Use(mw.SecurityHeaders)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   strings.Split(cfg.AllowedOrigins, ","),
		AllowedMethods:   []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Request-ID"},
		ExposedHeaders:   []string{"X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           300,
	}))
	r.Use(mw.OptionalAuth(cfg.AccessTokenSecret))

	authHandler := handler.NewAuthHandler(deps)
	profileHandler := handler.NewProfileHandler(deps)
	tournamentHandler := handler.NewTournamentHandler(deps)
	importHandler := handler.NewImportHandler(deps)
	teamHandler := handler.NewTeamHandler(deps)
	bracketHandler := handler.NewBracketHandler(deps)
	matchHandler := handler.NewMatchHandler(deps)
	notificationHandler := handler.NewNotificationHandler(deps)
	auditHandler := handler.NewAuditHandler(deps)
	adminHandler := handler.NewAdminHandler(deps)

	authLimiter := mw.NewRateLimiter(cfg.AuthRateLimitPerMinute, cfg.AuthRateLimitPerMinute/2+1)
	r.Route("/auth", func(ar chi.Router) {
		ar.Use(authLimiter.Middleware)
		ar.Post("/register", authHandler.Register)
		ar.Post("/login", authHandler.Login)
		ar.Post("/refresh", authHandler.Refresh)
		ar.Post("/logout", authHandler.Logout)
		ar.Post("/forgot-password", authHandler.ForgotPassword)
		ar.Post("/reset-password", authHandler.ResetPassword)
	})

	r.Get("/tournaments", tournamentHandler.ListPublic)
	r.Get("/tournaments/{id}", tournamentHandler.GetPublic)
	r.Get("/tournaments/{id}/teams", tournamentHandler.GetPublicTeams)
	r.Get("/tournaments/{id}/bracket", tournamentHandler.GetBracket)
	r.Get("/tournaments/{id}/matches", tournamentHandler.GetPublicMatches)

	r.Group(func(pr chi.Router) {
		pr.Use(mw.AuthRequired(cfg.AccessTokenSecret))

		pr.Get("/me", profileHandler.GetMe)
		pr.Patch("/me", profileHandler.UpdateMe)
		pr.Delete("/me", profileHandler.DeleteMe)

		pr.Post("/tournaments", tournamentHandler.Create)
		pr.Patch("/tournaments/{id}", tournamentHandler.Update)
		pr.Delete("/tournaments/{id}", tournamentHandler.Delete)
		pr.Post("/tournaments/{id}/status", tournamentHandler.ChangeStatus)
		pr.Post("/tournaments/{id}/managers", tournamentHandler.AddManager)
		pr.Delete("/tournaments/{id}/managers/{userId}", tournamentHandler.RemoveManager)

		pr.Post("/tournaments/{id}/google-sheet/connect", importHandler.ConnectSheet)
		pr.Post("/tournaments/{id}/google-sheet/validate", importHandler.ValidateSheet)
		pr.Post("/tournaments/{id}/imports/preview", importHandler.PreviewImport)
		pr.Post("/tournaments/{id}/imports/confirm", importHandler.ConfirmImport)
		pr.Get("/tournaments/{id}/imports", importHandler.ListImports)
		pr.Get("/imports/{batchId}", importHandler.GetImport)

		pr.Get("/tournaments/{id}/admin/teams", teamHandler.GetAdminTeams)
		pr.Get("/teams/{id}", teamHandler.GetTeam)
		pr.Patch("/teams/{id}", teamHandler.PatchTeam)
		pr.Post("/teams/{id}/approve", teamHandler.ApproveTeam)
		pr.Post("/teams/{id}/reject", teamHandler.RejectTeam)
		pr.Post("/teams/{id}/members/{memberId}/remove", teamHandler.RemoveMember)
		pr.Post("/teams/{id}/members/{memberId}/replace", teamHandler.ReplaceMember)

		pr.Post("/team-members/{id}/accept", teamHandler.AcceptMembership)
		pr.Post("/team-members/{id}/decline", teamHandler.DeclineMembership)

		pr.Post("/tournaments/{id}/bracket/generate", bracketHandler.Generate)
		pr.Post("/tournaments/{id}/bracket/regenerate", bracketHandler.Regenerate)
		pr.Post("/tournaments/{id}/bracket/reseed", bracketHandler.Reseed)

		pr.Get("/tournaments/{id}/admin/matches", matchHandler.GetAdminMatches)
		pr.Patch("/matches/{id}/schedule", matchHandler.Schedule)
		pr.Post("/matches/{id}/confirm-ready", matchHandler.ConfirmReady)
		pr.Post("/matches/{id}/request-reschedule", matchHandler.RequestReschedule)
		pr.Post("/matches/{id}/report-issue", matchHandler.ReportIssue)
		pr.Post("/matches/{id}/submit-result", matchHandler.SubmitResult)
		pr.Post("/matches/{id}/approve-result", matchHandler.ApproveResult)
		pr.Post("/matches/{id}/reject-result", matchHandler.RejectResult)

		pr.Get("/notifications", notificationHandler.List)
		pr.Get("/notifications/unread-count", notificationHandler.UnreadCount)
		pr.Post("/notifications/{id}/read", notificationHandler.Read)
		pr.Post("/notifications/read-all", notificationHandler.ReadAll)
		pr.Post("/notifications/{id}/action", notificationHandler.Action)

		pr.Get("/tournaments/{id}/audit", auditHandler.ListTournamentAudit)

		pr.Get("/admin/users", adminHandler.ListUsers)
		pr.Post("/admin/users/{id}/block", adminHandler.BlockUser)
		pr.Post("/admin/users/{id}/unblock", adminHandler.UnblockUser)
		pr.Get("/admin/tournaments", adminHandler.ListTournaments)
	})

	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	return r
}
