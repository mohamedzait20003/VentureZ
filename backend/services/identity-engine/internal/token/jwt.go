package token

import (
	"time"
	"errors"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	Roles []string `json:"roles"`
	jwt.RegisteredClaims
}

type Signer struct {
	priv *rsa.PrivateKey
	ttl  time.Duration
	issuer string
}

func NewSigner(priv *rsa.PrivateKey, ttl time.Duration, issuer string) *Signer {
	return &Signer{priv: priv, ttl: ttl, issuer: issuer}
}

func (s *Signer) Issue(userID, role string) (string, time.Duration, error) {
	now := time.Now()
	claims := Claims{
		Roles: []string{role},
		RegisteredClaims: jwt.RegisteredClaims{
			Subject: userID,
			Issuer: s.issuer,
			IssuedAt: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.ttl)),
		},
	}

	signed, err := jwt.NewWithClaims(jwt.SigningMethodRS256, claims).SignedString(s.priv)
	return signed, s.ttl, err
}

func LoadPrivateKey(pemData []byte) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(pemData)

	if block == nil {
		return nil, errors.New("invalid PEM data")
	}

	if k, err := x509.ParsePKCS8PrivateKey(block.Bytes); err == nil {
		if rk, ok := k.(*rsa.PrivateKey); ok {
			return rk, nil
		}

		return nil, errors.New("not an RSA private key")
	}

	return x509.ParsePKCS1PrivateKey(block.Bytes)
}

