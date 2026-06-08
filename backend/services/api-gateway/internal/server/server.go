package server

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httprate"

	identityv1 "github.com/venturez/backend/gen/go/identity/v1"
	"github.com/venturez/backend/services/api-gateway/internal/auth"
	"github.com/venturez/backend/services/api-gateway/internal/config"
	"github.com/venturez/backend/services/api-gateway/internal/handler"
	mw "github.com/venturez/backend/services/api-gateway/internal/middleware"
)

func New(cfg config.Config, verifier *auth.Verifier, identity identityv1.IdentityClient) http.Handler {
	authH := handler.NewAuth(identity, cfg.CookieSecure, cfg.RefreshCookieMaxAge)

	r := chi.NewRouter()
	r.Use(chimw.RequestID)
	r.Use(chimw.RealIP)
	r.Use(chimw.Logger)
	r.Use(chimw.Recoverer)
	r.Use(chimw.Timeout(30 * time.Second))
	r.Use(httprate.LimitByIP(cfg.RateLimitPerMin, time.Minute))

	r.Get("/healthz", handler.Health)

	r.Route("/auth", func(r chi.Router) {
		r.Post("/register", authH.Register)
		r.Post("/login", authH.Login)
		r.Post("/refresh", authH.Refresh)
		r.Post("/logout", authH.Logout)
		r.Post("/password-reset/request", authH.RequestPasswordReset)
		r.Post("/password-reset/confirm", authH.ResetPassword)
		r.Post("/verify-email", authH.VerifyEmail)
	})

	r.Route("/api/v1", func(r chi.Router) {
		r.Use(mw.Authenticator(verifier))
		r.Get("/me", handler.Me)
	})

	return r
}
