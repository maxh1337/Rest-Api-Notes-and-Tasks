package services

import (
	"context"
	"fmt"
	"log"
	"rest-api-notes/internal/domain/entities"
	"rest-api-notes/internal/infrastructure/cache"
	"time"

	"github.com/google/uuid"
)

type sessionService struct {
	redisClient cache.RedisClient
	ttl         time.Duration
}

type SessionService interface {
	CreateSession(ctx context.Context, userID uuid.UUID, sessionID, accessToken, refreshToken, userAgent, ip string) (*entities.Session, error)
	GetSession(ctx context.Context, userID uuid.UUID, sessionID string) (*entities.Session, error)
	DeleteSession(ctx context.Context, userID uuid.UUID, sessionID string) error
	UpdateSession(ctx context.Context, userID uuid.UUID, sessionID string, session *entities.Session) error
	IsSessionValid(ctx context.Context, userID uuid.UUID, sessionID, refreshToken string) (bool, error)
}

func NewSessionService(redisClient cache.RedisClient, ttl time.Duration) SessionService {
	return &sessionService{
		redisClient: redisClient,
		ttl:         ttl,
	}
}

func (s *sessionService) CreateSession(ctx context.Context, userID uuid.UUID, sessionID, accessToken, refreshToken, userAgent, ip string) (*entities.Session, error) {
	expiresAt := time.Now().Add(s.ttl).Unix()

	session := &entities.Session{
		SessionID:    sessionID,
		UserID:       userID,
		RefreshToken: refreshToken,
		AccessToken:  accessToken,
		UserAgent:    userAgent,
		IP:           ip,
		ExpiresAt:    expiresAt,
	}

	key := fmt.Sprintf("session:%s:%s", userID, sessionID)
	log.Printf("Key at create session created")
	if err := s.redisClient.SetStruct(ctx, key, session, s.ttl); err != nil {
		return nil, err
	}

	log.Printf("Set Struct executed without errors")

	return session, nil
}

func (s *sessionService) GetSession(ctx context.Context, userID uuid.UUID, sessionID string) (*entities.Session, error) {
	key := fmt.Sprintf("session:%s:%s", userID, sessionID)
	var session entities.Session
	if err := s.redisClient.GetStruct(ctx, key, &session); err != nil {
		if err.Error() == "redis: nil" {
			return nil, nil
		}
		return nil, err
	}
	return &session, nil
}

func (s *sessionService) IsSessionValid(ctx context.Context, userID uuid.UUID, sessionID, refreshToken string) (bool, error) {
	session, err := s.GetSession(ctx, userID, sessionID)
	if err != nil || session == nil {
		return false, err
	}

	if session.RefreshToken != refreshToken {
		return false, nil
	}

	if time.Now().Unix() > session.ExpiresAt {
		_ = s.DeleteSession(ctx, userID, sessionID)
		return false, nil
	}

	return true, nil
}

func (s *sessionService) DeleteSession(ctx context.Context, userID uuid.UUID, sessionID string) error {
	key := fmt.Sprintf("session:%s:%s", userID, sessionID)
	return s.redisClient.Delete(ctx, key)
}

func (s *sessionService) UpdateSession(ctx context.Context, userID uuid.UUID, sessionID string, session *entities.Session) error {
	key := fmt.Sprintf("session:%s:%s", userID, sessionID)
	return s.redisClient.SetStruct(ctx, key, session, s.ttl)
}
