package user

import (
	"context"
	"errors"
	"log/slog"
	"strings"

	"github.com/google/uuid"
	usererror "github.com/tim8912097887-sys/url-shortener/internal/shared/error/user_error"
	userschema "github.com/tim8912097887-sys/url-shortener/internal/shared/schema/user_schema"
	jwttoken "github.com/tim8912097887-sys/url-shortener/internal/shared/util/jwt_token"
	"golang.org/x/crypto/bcrypt"
)

type UserRepository interface {
	CreateUser(ctx context.Context, userInsert userschema.UserInsert) (*userschema.CreateUserRepositoryResponse, error)
	GetUserByEmail(ctx context.Context, email string) (*userschema.GetUserbyEmailRepositoryResponse, error)
	GetUserByID(ctx context.Context, id string) (*userschema.GetUserByIDRepositoryResponse, error)
	IncrementTokenVersion(ctx context.Context, id string) (int, error)
}

type TokenManager interface {
	GenerateAccessToken(userID string, tokenVersion int) (string, error)
	GenerateRefreshToken(userID string, tokenVersion int) (string, error)
	ParseAccessToken(tokenString string) (*jwttoken.AccessTokenClaims, error)
	ParseRefreshToken(tokenString string) (*jwttoken.RefreshTokenClaims, error)
}

type ServiceConfig struct {
	Repository UserRepository
	Tokens     jwttoken.TokenManager
	Logger     *slog.Logger
}

type service struct {
	repository UserRepository
	tokens     jwttoken.TokenManager
	logger     *slog.Logger
}

func NewService(serviceConfig *ServiceConfig) *service {
	return &service{
		repository: serviceConfig.Repository,
		tokens:     serviceConfig.Tokens,
		logger:     serviceConfig.Logger,
	}
}

// Signup always looks like it succeeded to the caller. If the email is new,
// a user is created; if it's already taken, CreateUser is a silent no-op.
// Either way we return nil, so signup can't be used to enumerate accounts.
func (s *service) Signup(ctx context.Context, createUserInput userschema.CreateUserInput) error {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(createUserInput.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	if _, err := s.repository.CreateUser(ctx, userschema.UserInsert{
		ID: uuid.New(),
		Username: createUserInput.Username, 
		Email: normalizeEmail(createUserInput.Email), 
		PasswordHash: string(passwordHash)}); err != nil {
		return err
	}

	return nil
}

// Login returns ErrInvalidCredential for both an unknown email and a wrong
// password — never anything that distinguishes the two.
func (s *service) Login(ctx context.Context, email, password string) (accessToken, refreshToken string, err error) {
	u, err := s.repository.GetUserByEmail(ctx, normalizeEmail(email))
	if err != nil {
		if errors.Is(err, usererror.ErrUserNotFound) {
			return "", "", usererror.ErrInvalidCredential
		}
		return "", "", err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		return "", "", usererror.ErrInvalidCredential
	}

	accessToken, err = s.tokens.GenerateAccessToken(u.ID, u.TokenVersion)
	if err != nil {
		return "", "", err
	}

	refreshToken, err = s.tokens.GenerateRefreshToken(u.ID, u.TokenVersion)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

// Logout re-validates the access token's claims against the DB (user still
// exists, token_version still current) before letting the handler clear the
// refresh cookie. There's no per-session table here, so this can't revoke
// that one refresh token server-side — only LogoutAll (which bumps
// token_version) actually invalidates outstanding tokens. If you need
// selective single-device revocation later, add a sessions/refresh-token
// table keyed by a token id (jti) instead of relying on token_version alone.
func (s *service) Logout(ctx context.Context, userID string, tokenVersion int) error {
	u, err := s.repository.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, usererror.ErrUserNotFound) {
			return usererror.ErrInvalidToken
		}
		return err
	}

	if u.TokenVersion != tokenVersion {
		return usererror.ErrInvalidToken
	}

	return nil
}

// LogoutAll validates the same way as Logout, then bumps token_version,
// which invalidates every access and refresh token issued to this user
// across every device.
func (s *service) LogoutAll(ctx context.Context, userID string, tokenVersion int) error {
	u, err := s.repository.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, usererror.ErrUserNotFound) {
			return usererror.ErrInvalidToken
		}
		return err
	}

	if u.TokenVersion != tokenVersion {
		return usererror.ErrInvalidToken
	}

	if _, err := s.repository.IncrementTokenVersion(ctx, userID); err != nil {
		return err
	}

	return nil
}

// Refresh validates the refresh token's claims against the DB, then rotates
// both tokens. Any failure — bad signature, expired token, unknown user,
// stale token_version — collapses to ErrInvalidToken.
func (s *service) Refresh(ctx context.Context, userID string, tokenVersion int) (newAccessToken, newRefreshToken string, err error) {

	u, err := s.repository.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, usererror.ErrUserNotFound) {
			return "", "", usererror.ErrInvalidToken
		}
		return "", "", err
	}

	if u.TokenVersion != tokenVersion {
		return "", "", usererror.ErrInvalidToken
	}

	newAccessToken, err = s.tokens.GenerateAccessToken(u.ID, u.TokenVersion)
	if err != nil {
		return "", "", err
	}

	newRefreshToken, err = s.tokens.GenerateRefreshToken(u.ID, u.TokenVersion)
	if err != nil {
		return "", "", err
	}

	return newAccessToken, newRefreshToken, nil
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}