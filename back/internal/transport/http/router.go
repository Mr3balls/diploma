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
		AllowedMethods:   []string{"GET", "POST", "PATCH", "PUT", "DELETE", "OPTIONS"},
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
	r.Get("/tournaments/{id}/participants", tournamentHandler.GetParticipants)
	r.Get("/teams/{id}", teamHandler.GetTeam)

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

		pr.Post("/tournaments/{id}/register-team", tournamentHandler.RegisterTeam)

		// Individual-mode participant management
		pr.Post("/tournaments/{id}/participants", tournamentHandler.AddParticipant)
		pr.Post("/tournaments/{id}/participants/bulk", tournamentHandler.BulkAddParticipants)
		pr.Delete("/tournaments/{id}/participants/{participantId}", tournamentHandler.RemoveParticipant)
		pr.Post("/tournaments/{id}/participants/shuffle", tournamentHandler.ShuffleParticipants)
		pr.Post("/tournaments/{id}/start-bracket", tournamentHandler.StartIndividualBracket)
		pr.Post("/tournaments/{id}/join", tournamentHandler.JoinIndividual)

		pr.Post("/tournaments/{id}/google-sheet/connect", importHandler.ConnectSheet)
		pr.Post("/tournaments/{id}/google-sheet/validate", importHandler.ValidateSheet)
		pr.Post("/tournaments/{id}/imports/preview", importHandler.PreviewImport)
		pr.Post("/tournaments/{id}/imports/confirm", importHandler.ConfirmImport)
		pr.Get("/tournaments/{id}/imports", importHandler.ListImports)
		pr.Get("/imports/{batchId}", importHandler.GetImport)

		pr.Get("/tournaments/{id}/admin/teams", teamHandler.GetAdminTeams)
		pr.Post("/tournaments/{id}/admin/teams", teamHandler.AdminCreateTeam)
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
		pr.Post("/matches/{id}/reset", bracketHandler.ResetMatch)

		pr.Get("/tournaments/{id}/admin/matches", matchHandler.GetAdminMatches)
		pr.Patch("/matches/{id}/schedule", matchHandler.Schedule)
		pr.Post("/matches/{id}/confirm-ready", matchHandler.ConfirmReady)
		pr.Post("/matches/{id}/request-reschedule", matchHandler.RequestReschedule)
		pr.Post("/matches/{id}/report-issue", matchHandler.ReportIssue)
		pr.Post("/matches/{id}/submit-result", matchHandler.SubmitResult)
		pr.Post("/matches/{id}/approve-result", matchHandler.ApproveResult)
		pr.Post("/matches/{id}/reject-result", matchHandler.RejectResult)
		pr.Post("/matches/{id}/admin-set-result", matchHandler.AdminSetResult)

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

	// ── Challonge-style individual-participant tournaments ────────────────────
	challongeHandler := handler.NewChallongeHandler(deps)

	// Public (no auth needed)
	r.Get("/c/{slug}", challongeHandler.GetBracket)
	r.Get("/c/{slug}/standings", challongeHandler.GetStandings)
	r.Get("/c/{slug}/events", challongeHandler.ServeEvents)

	// Auth-required
	r.Group(func(cr chi.Router) {
		cr.Use(mw.AuthRequired(cfg.AccessTokenSecret))

		cr.Post("/c", challongeHandler.CreateTournament)
		cr.Get("/c/my-matches", challongeHandler.GetMyMatches)
		cr.Post("/c/invites/{token}/accept", challongeHandler.AcceptInvite)

		cr.Post("/c/{slug}/join", challongeHandler.Join)

		// Participant management (organizer / co-organizer)
		cr.Post("/c/{slug}/participants", challongeHandler.AddParticipant)
		cr.Post("/c/{slug}/participants/bulk", challongeHandler.BulkAddParticipants)
		cr.Delete("/c/{slug}/participants/{participantID}", challongeHandler.RemoveParticipant)
		cr.Post("/c/{slug}/participants/shuffle", challongeHandler.ShuffleParticipants)
		cr.Put("/c/{slug}/participants/reorder", challongeHandler.ReorderParticipants)

		// Lifecycle (organizer / co-organizer)
		cr.Post("/c/{slug}/start", challongeHandler.Start)
		cr.Post("/c/{slug}/reset", challongeHandler.Reset)
		cr.Post("/c/{slug}/unfinalize", challongeHandler.Unfinalize)

		// Match results (organizer / co-organizer)
		cr.Post("/c/{slug}/matches/{matchID}/result", challongeHandler.SubmitResult)
		cr.Post("/c/{slug}/matches/{matchID}/reset", challongeHandler.ResetMatch)

		// Match result reporting (participant self-report)
		cr.Post("/c/{slug}/matches/{matchID}/report", challongeHandler.ReportResult)
		cr.Post("/c/{slug}/matches/{matchID}/reports/{reportID}/approve", challongeHandler.ApproveReport)
		cr.Post("/c/{slug}/matches/{matchID}/reports/{reportID}/reject", challongeHandler.RejectReport)

		// Audit log (organizer / co-organizer)
		cr.Get("/c/{slug}/log", challongeHandler.GetLog)

		// Co-organizer management (organizer only)
		cr.Post("/c/{slug}/co-organizers/invite", challongeHandler.InviteCoOrganizer)
		cr.Delete("/c/{slug}/co-organizers/{userID}", challongeHandler.RemoveCoOrganizer)
	})

	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	return r
}
