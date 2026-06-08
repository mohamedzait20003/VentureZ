package middleware

import (
	"net/http"
	"strings"

	"github.com/venturez/backend/services/api-gateway/internal/auth"
	"github.com/venturez/backend/services/api-gateway/internal/respond"
)

func Authenticator(v *auth.Verifier) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h := r.Header.Get("Authorization")
			if !strings.HasPrefix(h, "Bearer ") {
				respond.Error(w, http.StatusUnauthorized, "missing bearer token")
				return
			}
			
			claims, err := v.Parse(strings.TrimPrefix(h, "Bearer "))
			if err != nil {
				respond.Error(w, http.StatusUnauthorized, "invalid token")
				return
			}

			next.ServeHTTP(w, r.WithContext(auth.WithClaims(r.Context(), claims)))
		})
	}
}

func RequireRole(role string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := auth.FromContext(r.Context())
			if !ok {
				respond.Error(w, http.StatusUnauthorized, "unauthenticated")
				return
			}

			if !claims.HasRole(role) {
				respond.Error(w, http.StatusForbidden, "forbidden: requires role "+role)
				return
			}
			
			next.ServeHTTP(w, r)
		})
	}
}
