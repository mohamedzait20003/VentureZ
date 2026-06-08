package config

import (
	"os"
	"strconv"
)

type Config struct {
	HTTPAddr string
	JWTPublicKeyPath string

	IdentityAddr string
	AgentAddr string
	FinancialEngineAddr string

	RateLimitPerMin int

	CookieSecure bool
	RefreshCookieMaxAge int
}

func Load() Config {
	return Config{
		HTTPAddr: getenv("HTTP_ADDR", ":8080"),
		JWTPublicKeyPath: getenv("JWT_PUBLIC_KEY_PATH", "./keys/jwt_public.pem"),
		IdentityAddr: getenv("IDENTITY_ADDR", "localhost:50053"),
		AgentAddr: getenv("AGENT_ADDR", "localhost:50052"),
		FinancialEngineAddr: getenv("FINANCIAL_ENGINE_ADDR", "localhost:50051"),
		RateLimitPerMin: getenvInt("RATE_LIMIT_PER_MIN", 60),
		CookieSecure: getenvBool("COOKIE_SECURE", false),
		RefreshCookieMaxAge: getenvInt("REFRESH_COOKIE_MAX_AGE", 7*24*3600),
	}
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}

	return def
}

func getenvInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}

	return def
}

func getenvBool(key string, def bool) bool {
	if v := os.Getenv(key); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}

	return def
}
