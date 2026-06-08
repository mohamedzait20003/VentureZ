package identity_test

import (
	"testing"

	"github.com/venturez/backend/services/identity-engine/internal/password"
)

func TestHashAndVerify(t *testing.T) {
	hash, err := password.HashPassword("Admin@1234")
	if err != nil {
		t.Fatalf("hash: %v", err)
	}
	if hash == "Admin@1234" {
		t.Fatal("password was not hashed")
	}
	if !password.VerifyPassword(hash, "Admin@1234") {
		t.Fatal("correct password should verify")
	}
	if password.VerifyPassword(hash, "wrong-password") {
		t.Fatal("wrong password must not verify")
	}
}
