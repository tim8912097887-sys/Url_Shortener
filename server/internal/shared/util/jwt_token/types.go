package jwttoken

import "github.com/golang-jwt/jwt/v5"

type AccessTokenClaims struct {
	UserID       string `json:"user_id"`
	TokenVersion int    `json:"token_version"`
	Type         string `json:"type"`
	jwt.RegisteredClaims
}

type RefreshTokenClaims struct {
	UserID       string `json:"user_id"`
	TokenVersion int    `json:"token_version"`
	Type         string `json:"type"`
	jwt.RegisteredClaims
}