package oauthschema

import (
	"github.com/google/uuid"
)

type OAuthAccountInsert struct {
	ID                uuid.UUID
	Provider          Provider
	ProviderAccountID string
	ProviderEmail     string
}

type CreateOAuthAccountRepositoryResponse struct {
	ID                string
	UserID            string
	Provider          Provider
	ProviderAccountID string
}

type GetUserByOAuthAccountRepositoryResponse struct {
	UserID       string
	TokenVersion int
}