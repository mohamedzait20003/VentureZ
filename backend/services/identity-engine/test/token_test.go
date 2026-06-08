package identity_test

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/venturez/backend/services/identity-engine/internal/token"
)

func TestNewAndHash(t *testing.T) {
	raw, h, err := token.New()
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	if raw == "" || h == "" {
		t.Fatal("token and hash must be non-empty")
	}

	if raw == h {
		t.Fatal("raw token must not equal its hash")
	}

	if token.Hash(raw) != h {
		t.Fatal("Hash(raw) must reproduce the stored hash")
	}

	raw2, _, _ := token.New()
	if raw2 == raw {
		t.Fatal("successive tokens must be unique")
	}
}

func TestSignerIssueAndVerify(t *testing.T) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("genkey: %v", err)
	}

	s := token.NewSigner(priv, 15*time.Minute, "venturez-identity")

	tok, ttl, err := s.Issue("user_1", "admin")
	if err != nil {
		t.Fatalf("issue: %v", err)
	}
	
	if ttl != 15*time.Minute {
		t.Fatalf("ttl = %v, want 15m", ttl)
	}

	claims := &token.Claims{}
	parsed, err := jwt.ParseWithClaims(tok, claims, func(*jwt.Token) (any, error) {
		return &priv.PublicKey, nil
	})
	if err != nil || !parsed.Valid {
		t.Fatalf("verify failed: %v", err)
	}
	if claims.Subject != "user_1" {
		t.Fatalf("subject = %q, want user_1", claims.Subject)
	}
	if len(claims.Roles) != 1 || claims.Roles[0] != "admin" {
		t.Fatalf("roles = %v, want [admin]", claims.Roles)
	}
	if claims.Issuer != "venturez-identity" {
		t.Fatalf("issuer = %q, want venturez-identity", claims.Issuer)
	}
}
