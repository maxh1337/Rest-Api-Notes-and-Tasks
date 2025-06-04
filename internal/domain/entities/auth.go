package entities

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalidCredentials                  = errors.New("invalid credentials")
	ErrFailedToCreateAccessToken           = errors.New("failed to create access token")
	ErrFailedToCreateRefreshToken          = errors.New("failed to create refresh token")
	ErrFailedToLoginRefreshTokenNotProvide = errors.New("failed to login, please provide correct refreshToken")
)

type Session struct {
	SessionID    string    `json:"session_id"`
	UserID       uuid.UUID `json:"user_id"`
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	UserAgent    string    `json:"user_agent"`
	IP           string    `json:"ip"`
	ExpiresAt    int64     `json:"expires_at"`
}

type JWTClaims struct {
	UserID    uuid.UUID
	SessionID string
	ExpiresAt time.Time
	TokenType string
}

type UserRegisterReq struct {
	Username string `json:"username" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}
type UserAuthRes struct {
	ID       uuid.UUID `json:"id"`
	Username string    `json:"username"`
	Email    string    `json:"email"`
	Role     RoleType  `json:"role"`
}

type UserLoginReq struct {
	Identifier string `json:"identifier" validate:"required"`
	Password   string `json:"password" validate:"required"`
}

type UserLogoutReq struct {
	RefreshToken string `json:"refresh_token"`
	SessionID    string `json:"session_id"`
}
