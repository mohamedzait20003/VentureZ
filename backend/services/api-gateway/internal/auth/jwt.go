package auth

import (
	"crypto/rsa"
	"errors"
	"fmt"

	"github.com/golang-jwt/jwt/v5"
)

type Verifier struct {
	pub *rsa.PublicKey
}

func NewVerifier(publicKeyPEM []byte) (*Verifier, error) {
	pub, err := jwt.ParseRSAPublicKeyFromPEM(publicKeyPEM)

	if err != nil {
		return nil, err
	}

	return &Verifier{pub: pub}, nil
}

func (v *Verifier) Parse(tokenStr string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}

		return v.pub, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}
	
	return claims, nil
}
