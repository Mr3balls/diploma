package middleware

import (
	"context"
	"net/http"
	"strings"

	"esports-backend/internal/apperror"
	tok "esports-backend/internal/pkg/tokens"
)

type AuthUser struct {
	UserID string
	Roles  []string
}

type ctxKey string

const userCtxKey ctxKey = "auth_user"

func AuthRequired(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				writeError(w, apperror.Unauthorized("missing bearer token"))
				return
			}
			token := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer"))
			if token == authHeader {
				writeError(w, apperror.Unauthorized("invalid authorization header"))
				return
			}
			claims, err := tok.ParseAccessToken(secret, strings.TrimSpace(token))
			if err != nil {
				writeError(w, apperror.Unauthorized("invalid or expired access token"))
				return
			}
			ctx := context.WithValue(r.Context(), userCtxKey, &AuthUser{UserID: claims.UserID, Roles: claims.Roles})
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func OptionalAuth(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				next.ServeHTTP(w, r)
				return
			}
			token := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer"))
			if token == authHeader {
				next.ServeHTTP(w, r)
				return
			}
			claims, err := tok.ParseAccessToken(secret, strings.TrimSpace(token))
			if err == nil {
				ctx := context.WithValue(r.Context(), userCtxKey, &AuthUser{UserID: claims.UserID, Roles: claims.Roles})
				r = r.WithContext(ctx)
			}
			next.ServeHTTP(w, r)
		})
	}
}

func CurrentUser(ctx context.Context) *AuthUser {
	user, _ := ctx.Value(userCtxKey).(*AuthUser)
	return user
}

func HasPlatformRole(user *AuthUser, role string) bool {
	if user == nil {
		return false
	}
	for _, item := range user.Roles {
		if item == role {
			return true
		}
	}
	return false
}
