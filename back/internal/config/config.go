package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv                   string
	HTTPPort                 string
	FrontendURL              string
	AccessTokenSecret        string
	AccessTokenTTL           time.Duration
	RefreshTokenTTL          time.Duration
	DatabaseURL              string
	RedisURL                 string
	AllowedOrigins           string
	GoogleServiceAccountFile string
	GoogleDefaultWorksheet   string
	AuthRateLimitPerMinute   int
	BrevoAPIKey              string
	EmailFromAddress         string
	EmailFromName            string
	VAPIDPublicKey           string
	VAPIDPrivateKey          string
	VAPIDEmail               string
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		AppEnv:                   getEnv("APP_ENV", "development"),
		HTTPPort:                 getEnv("HTTP_PORT", "8080"),
		AccessTokenSecret:        getEnv("ACCESS_TOKEN_SECRET", "change-me-super-secret"),
		AccessTokenTTL:           getDuration("ACCESS_TOKEN_TTL", "15m"),
		RefreshTokenTTL:          getDuration("REFRESH_TOKEN_TTL", "720h"),
		FrontendURL:              getEnv("FRONTEND_URL", "http://localhost:5173"),
		DatabaseURL:              getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/esports?sslmode=disable"),
		RedisURL:                 getEnv("REDIS_URL", "redis://localhost:6379/0"),
		AllowedOrigins:           getEnv("ALLOWED_ORIGINS", "http://localhost:3000,http://localhost:5173"),
		GoogleServiceAccountFile: getEnvAllowEmpty("GOOGLE_SERVICE_ACCOUNT_FILE", "./credentials/google-service-account.json"),
		GoogleDefaultWorksheet:   getEnv("GOOGLE_DEFAULT_WORKSHEET", "Sheet1"),
		AuthRateLimitPerMinute:   getInt("AUTH_RATE_LIMIT_PER_MINUTE", 30),
		BrevoAPIKey:              getEnvAllowEmpty("BREVO_API_KEY", ""),
		EmailFromAddress:         getEnvAllowEmpty("EMAIL_FROM_ADDRESS", "gulyachinaalan@gmail.com"),
		EmailFromName:            getEnvAllowEmpty("EMAIL_FROM_NAME", "ACE Tournament"),
		VAPIDPublicKey:           getEnvAllowEmpty("VAPID_PUBLIC_KEY", ""),
		VAPIDPrivateKey:          getEnvAllowEmpty("VAPID_PRIVATE_KEY", ""),
		VAPIDEmail:               getEnvAllowEmpty("VAPID_EMAIL", ""),
	}

	if cfg.AccessTokenSecret == "" {
		return nil, fmt.Errorf("ACCESS_TOKEN_SECRET is required")
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func getEnvAllowEmpty(key, fallback string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		return fallback
	}
	return value
}

func getDuration(key, fallback string) time.Duration {
	raw := getEnv(key, fallback)
	d, err := time.ParseDuration(raw)
	if err != nil {
		d, _ = time.ParseDuration(fallback)
	}
	return d
}

func getInt(key string, fallback int) int {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback
	}
	v, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}
	return v
}
