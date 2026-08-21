package jwttoken

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	usererror "github.com/tim8912097887-sys/url-shortener/internal/shared/error/user_error"
)

// tokenManager implements the TokenManager interface declared in service.go.
// Access and refresh tokens are signed with different secrets so a leaked
// access-token secret can't be used to forge long-lived refresh tokens.
type TokenManager struct {
	accessSecret  []byte
	refreshSecret []byte
}

func NewTokenManager(accessSecret, refreshSecret string) *TokenManager {
	return &TokenManager{
		accessSecret:  []byte(accessSecret),
		refreshSecret: []byte(refreshSecret),
	}
}

func (m *TokenManager) GenerateAccessToken(userID string, tokenVersion int) (string, error) {
	claims := AccessTokenClaims{
		UserID:       userID,
		TokenVersion: tokenVersion,
		Type:         "access",
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(AccessTokenTTL)),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(m.accessSecret)
}

func (m *TokenManager) GenerateRefreshToken(userID string, tokenVersion int) (string, error) {
	claims := RefreshTokenClaims{
		UserID:       userID,
		TokenVersion: tokenVersion,
		Type:         "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(RefreshTokenTTL)),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(m.refreshSecret)
}

func (m *TokenManager) ParseAccessToken(tokenString string) (*AccessTokenClaims, error) {
	claims := &AccessTokenClaims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, usererror.ErrInvalidToken
		}
		return m.accessSecret, nil
	})
	if err != nil || !token.Valid {
		return nil, usererror.ErrInvalidToken
	}

	return claims, nil
}

func (m *TokenManager) ParseRefreshToken(tokenString string) (*RefreshTokenClaims, error) {
	claims := &RefreshTokenClaims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, usererror.ErrInvalidToken
		}
		return m.refreshSecret, nil
	})
	if err != nil || !token.Valid {
		return nil, usererror.ErrInvalidToken
	}

	return claims, nil
}