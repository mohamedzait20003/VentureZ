package handler

import (
	"net/http"

	"github.com/venturez/backend/services/api-gateway/internal/auth"
	"github.com/venturez/backend/services/api-gateway/internal/respond"
)

func Health(w http.ResponseWriter, r *http.Request) {
	respond.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func Me(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.FromContext(r.Context())
	if !ok {
		respond.Error(w, http.StatusUnauthorized, "unauthenticated")
		return
	}

	respond.JSON(w, http.StatusOK, map[string]any{
		"user_id": claims.Subject,
		"roles":   claims.Roles,
	})
}
