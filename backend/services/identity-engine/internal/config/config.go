package config

import (
	"os"
	"time"
)

type Config struct {
	GRPCAddr string
	DatabaseURL string
	JWTPrivateKeyPath string
	Issuer string
	AccessTTL time.Duration
	RefreshTTL time.Duration

	// Email (SMTP). If SMTPHost is empty, identity-engine falls back to a
	// dev "log" mailer instead of sending real email.
	SMTPHost string
	SMTPPort string
	SMTPUser string
	SMTPPass string
	MailFrom string
	AppBaseURL string // base URL for links in emails, e.g. https://app.venturez.dev
}

func Load() Config {
	return Config{
		GRPCAddr:    env("GRPC_ADDR", ":50053"),
		DatabaseURL: env("DATABASE_URL", "postgres://venturez:venturez@localhost:5432/venturez?sslmode=disable"),
		JWTPrivateKeyPath: env("JWT_PRIVATE_KEY_PATH", "./keys/jwt_private.pem"),
		Issuer: env("JWT_ISSUER", "venturez-identity"),
		AccessTTL: dur("ACCESS_TTL", 15*time.Minute),
		RefreshTTL: dur("REFRESH_TTL", 7*24*time.Hour),

		SMTPHost: env("SMTP_HOST", ""),
		SMTPPort: env("SMTP_PORT", "587"),
		SMTPUser: env("SMTP_USER", ""),
		SMTPPass: env("SMTP_PASS", ""),
		MailFrom: env("MAIL_FROM", "VentureZ <no-reply@venturez.dev>"),
		AppBaseURL: env("APP_BASE_URL", "http://localhost:3000"),
	}
}

func env(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}

	return def
}

func dur(key string, def time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if p, err := time.ParseDuration(v); err == nil {
			return p
		}
	}
	
	return def
}
