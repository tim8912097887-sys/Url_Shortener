package userschema

import "github.com/google/uuid"

type UserInsert struct {
	ID           uuid.UUID `json:"id"`
	Username     string `json:"username"`
	Email        string `json:"email"`
	PasswordHash string `json:"password_hash"`
}

type CreateUserRepositoryResponse struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

type GetUserbyEmailRepositoryResponse struct {
	ID           string `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"password_hash"`
	TokenVersion int       `json:"token_version"`
}

type GetUserByIDRepositoryResponse struct {
	ID           string `json:"id"`
	TokenVersion int    `json:"token_version"`
}
