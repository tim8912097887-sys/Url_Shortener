package oauth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	oautherror "github.com/tim8912097887-sys/url-shortener/internal/shared/error/oauth_error"
	usererror "github.com/tim8912097887-sys/url-shortener/internal/shared/error/user_error"
	oauthschema "github.com/tim8912097887-sys/url-shortener/internal/shared/schema/oauth_schema"
	userschema "github.com/tim8912097887-sys/url-shortener/internal/shared/schema/user_schema"
	jwttoken "github.com/tim8912097887-sys/url-shortener/internal/shared/util/jwt_token"
	"golang.org/x/oauth2"
)

type OAuthRepository interface {
	CreateOAuthAccount(
		ctx context.Context,
		userInsert userschema.UserInsert,
	    oauthInsert oauthschema.OAuthAccountInsert,
	) (*oauthschema.CreateOAuthAccountRepositoryResponse, error)

	GetUserByOAuthAccount(
		ctx context.Context,
		provider oauthschema.Provider,
		providerAccountID string,
	) (*oauthschema.GetUserByOAuthAccountRepositoryResponse, error)
}

type OAuthCache interface {
	Save(ctx context.Context, state string) error
	Consume(ctx context.Context, state string) error
}

type TokenManager interface {
	GenerateAccessToken(userID string, tokenVersion int) (string, error)
	GenerateRefreshToken(userID string, tokenVersion int) (string, error)
	ParseAccessToken(tokenString string) (*jwttoken.AccessTokenClaims, error)
	ParseRefreshToken(tokenString string) (*jwttoken.RefreshTokenClaims, error)
}

type ServiceConfig struct {
	Cache OAuthCache
	OAuthConfig *OAuth
	TokenManager TokenManager
	Repository OAuthRepository
}

type Service struct {
	cache OAuthCache
	oauthConfig *OAuth
	tokenManager TokenManager
	repository OAuthRepository
}

func NewService(serviceConfig *ServiceConfig) *Service {
	return &Service{
		cache: serviceConfig.Cache,
		oauthConfig: serviceConfig.OAuthConfig,
		tokenManager: serviceConfig.TokenManager,
		repository: serviceConfig.Repository,
	}
}

func (s *Service) GetAuthorizationURL(
	ctx context.Context,
	provider oauthschema.Provider,
) (string, error) {
	config, ok := s.oauthConfig.GetProvider(provider)
	if !ok {
		return "", oautherror.ErrInvalidProvider
	}

	state, err := generateSecureState()
	if err != nil {
		return "", err
	}

	if err := s.cache.Save(ctx, state); err != nil {
		return "", err
	}

	return config.AuthCodeURL(
		state,
		oauth2.AccessTypeOffline,
	), nil
}

func (s *Service) Authenticate(
	ctx context.Context,
	provider oauthschema.Provider,
	state string,
	code string,
) (*oauthschema.TokenResponse, error) {
	config, ok := s.oauthConfig.GetProvider(provider)
	if !ok {
		return nil, oautherror.ErrInvalidProvider
	}

	// check state
	if err := s.cache.Consume(ctx, state); err != nil {
		return nil, err
	}
    
	token, err := config.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf(
			"exchange oauth code: %w",
			err,
		)
	}
    
	var oauthUser *oauthschema.OAuthUser
	switch provider {
	case ProviderGoogle:
		oauthUser, err = s.getGoogleUser(ctx, token)
        if err != nil {
			return nil, err
		}
	default:
		return nil, oautherror.ErrInvalidProvider
	}
    
	var existUser *oauthschema.GetUserByOAuthAccountRepositoryResponse
	existUser, err = s.repository.GetUserByOAuthAccount(ctx, oauthUser.Provider, oauthUser.ProviderID)

	if err != nil {
		if errors.Is(err, usererror.ErrUserNotFound) {
			createdUser, err := s.repository.CreateOAuthAccount(ctx, userschema.UserInsert{
				ID: uuid.New(),
				Username: oauthUser.Name,
				Email: oauthUser.Email,
			}, oauthschema.OAuthAccountInsert{
				ID: uuid.New(),
				Provider: oauthUser.Provider,
				ProviderAccountID: oauthUser.ProviderID,
				ProviderEmail: oauthUser.Email,
			})

			if err != nil {
				return nil, err
			}

			existUser = &oauthschema.GetUserByOAuthAccountRepositoryResponse{
				UserID: createdUser.UserID,
				TokenVersion: 0,
			}
		} else {
			return nil, err
		}
	}
   
	accessToken, err := s.tokenManager.GenerateAccessToken(existUser.UserID, existUser.TokenVersion)
	if err != nil {
		return &oauthschema.TokenResponse{
			AccessToken:  "",
			RefreshToken: "",
		}, err
	}

	refreshToken, err := s.tokenManager.GenerateRefreshToken(existUser.UserID, existUser.TokenVersion)
	if err != nil {
		return &oauthschema.TokenResponse{
			AccessToken:  "",
			RefreshToken: "",
		}, err
	}
    
	return &oauthschema.TokenResponse{
		AccessToken: accessToken, 
		RefreshToken: refreshToken,
	}, nil
}

func (s *Service) getGoogleUser(
    ctx context.Context,
    token *oauth2.Token,
) (*oauthschema.OAuthUser, error) {
    client := oauth2.NewClient(
        ctx,
        oauth2.StaticTokenSource(token),
    )

    req, err := http.NewRequestWithContext(
        ctx,
        http.MethodGet,
        "https://openidconnect.googleapis.com/v1/userinfo",
        nil,
    )
    if err != nil {
        return nil, fmt.Errorf(
            "create google user request: %w",
            err,
        )
    }

    resp, err := client.Do(req)
    if err != nil {
        return nil, fmt.Errorf(
            "get google user: %w",
            err,
        )
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf(
            "google userinfo returned status %d",
            resp.StatusCode,
        )
    }

    var googleUser oauthschema.GoogleUserResponse

    if err := json.NewDecoder(resp.Body).Decode(
        &googleUser,
    ); err != nil {
        return nil, fmt.Errorf(
            "decode google user: %w",
            err,
        )
    }

    return &oauthschema.OAuthUser{
        Provider:       ProviderGoogle,
        ProviderID: googleUser.ID,
        Email:          googleUser.Email,
        Name:           googleUser.Name,
        AvatarURL:      googleUser.Picture,
    }, nil
}

func generateSecureState() (string, error) {
	b := make([]byte, 32)

	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(b), nil
}