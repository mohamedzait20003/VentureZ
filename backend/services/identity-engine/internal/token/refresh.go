package token

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
)

func hashToken(rawToken string) string {
	sum := sha256.Sum256([]byte(rawToken))
	return base64.URLEncoding.EncodeToString(sum[:])
}

func New() (rawToken, hash string, err error) {
	b := make([]byte, 32)

	if _, err := rand.Read(b); err != nil {
		return "", "", err
	}

	rawToken = base64.URLEncoding.EncodeToString(b)
	return rawToken, hashToken(rawToken), nil
}

func Hash(rawToken string) string {
	return hashToken(rawToken)
}