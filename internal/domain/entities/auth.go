package entities

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalidCredentials         = errors.New("invalid credentials")
	ErrFailedToCreateAccessToken  = errors.New("failed to create accessToken")
	ErrFailedToCreateRefreshToken = errors.New("failed to create refreshToken")
	ErrRefreshTokenNotProvided    = errors.New("refreshToken not provided")
	ErrSessionIDTokenNotProvided  = errors.New("session_id token not provided")
	Err2FARequired                = errors.New("2fa_required")
	ErrTokensMismatch             = errors.New("mismatch refreshToken from request and refreshToken from session")
)

type JWTClaims struct {
	UserID    uuid.UUID
	SessionID string
	ExpiresAt time.Time
	TokenType string
}

type TwoFactorJWTClaims struct {
	UserID    uuid.UUID
	UserAgent string
	UserIP    string
	ExpiresAt time.Time
}

type UserRegisterReq struct {
	Username string `json:"username" validate:"required,min=3,max=50,alphanum"`
	Email    string `json:"email" validate:"required,email,max=255"`
	Password string `json:"password" validate:"required,min=8,max=128"`
}

type UserAuthRes struct {
	ID             uuid.UUID `json:"id"`
	Username       string    `json:"username"`
	Email          string    `json:"email"`
	Role           RoleType  `json:"role"`
	TwoFactorToken string    `json:"two_factor_token"`
}

type UserLoginReq struct {
	Identifier string `json:"identifier" validate:"required,max=255"`
	Password   string `json:"password" validate:"required"`
}

type UserLogoutReq struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
	SessionID    string `json:"session_id" validate:"required,uuid"`
}

type UserGetNewTokensReq struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
	SessionID    string `json:"session_id" validate:"required,uuid"`
}
