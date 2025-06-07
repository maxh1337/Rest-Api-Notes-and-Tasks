package entities

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrSessionBelongsToAnotherDevice = errors.New("session belongs to another device or user")
	ErrSessionExpired                = errors.New("session expired, please login")
	ErrSessionNotFound               = errors.New("session not found")
	ErrInvalidSessionID              = errors.New("invalid session ID")
)

type Session struct {
	SessionID    string    `json:"session_id" redis:"session_id"`
	UserID       uuid.UUID `json:"user_id" redis:"user_id"`
	AccessToken  string    `json:"access_token" redis:"access_token"`
	RefreshToken string    `json:"refresh_token" redis:"refresh_token"`
	UserAgent    string    `json:"user_agent" redis:"user_agent"`
	IP           string    `json:"ip" redis:"ip"`
	ExpiresAt    int64     `json:"expires_at" redis:"expires_at"`
}

func (s *Session) IsExpired() bool {
	return time.Now().Unix() > s.ExpiresAt
}

func (s *Session) IsValid(refreshToken string) bool {
	return !s.IsExpired() && s.RefreshToken == refreshToken
}
