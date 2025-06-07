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
	GetAllUserSessions(ctx context.Context, userID uuid.UUID) (*[]entities.Session, error)
	Save2FACode(ctx context.Context, userID uuid.UUID, code string, context entities.TwoFASessionContext, token ...string) error
	Delete2FACode(ctx context.Context, userID uuid.UUID, context entities.TwoFASessionContext) error
	Get2FAData(ctx context.Context, userID uuid.UUID, context entities.TwoFASessionContext) (*entities.TwoFASessionData, error)
	Verify2FACode(ctx context.Context, userID uuid.UUID, code string, context entities.TwoFASessionContext) (*entities.TwoFASessionData, error)
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

func (s *sessionService) GetAllUserSessions(ctx context.Context, userID uuid.UUID) (*[]entities.Session, error) {
	key := fmt.Sprintf("session:%s:*", userID.String())
	var sessions []entities.Session
	err := s.redisClient.GetAllByKey(ctx, key, &sessions)
	if err != nil {
		return nil, err
	}
	return &sessions, nil
}

func (s *sessionService) Save2FACode(ctx context.Context, userID uuid.UUID, code string, context entities.TwoFASessionContext, token ...string) error {
	key := fmt.Sprintf("2fa_code:%s:%s", userID, context)

	data := &entities.TwoFASessionData{
		Code:      code,
		Context:   context,
		UserID:    userID,
		ExpiresAt: time.Now().Add(5 * time.Minute).Unix(),
	}

	_, err := s.Get2FAData(ctx, userID, context)
	if err != nil {
		if len(token) > 0 {
			data.Token = token[0]
		}

		return s.redisClient.SetStruct(ctx, key, data, 5*time.Minute)
	}

	if err := s.Delete2FACode(ctx, userID, context); err != nil {
		return err
	}

	if len(token) > 0 {
		data.Token = token[0]
	}

	return s.redisClient.SetStruct(ctx, key, data, 5*time.Minute)
}

func (s *sessionService) Delete2FACode(ctx context.Context, userID uuid.UUID, context entities.TwoFASessionContext) error {
	key := fmt.Sprintf("2fa_code:%s:%s", userID, context)
	return s.redisClient.Delete(ctx, key)
}

func (s *sessionService) Get2FAData(ctx context.Context, userID uuid.UUID, context entities.TwoFASessionContext) (*entities.TwoFASessionData, error) {
	key := fmt.Sprintf("2fa_code:%s:%s", userID, context)
	var data entities.TwoFASessionData
	if err := s.redisClient.GetStruct(ctx, key, &data); err != nil {
		return nil, entities.Err2FACodeInvalidOrExpired
	}

	// Проверяем не истек ли код
	if time.Now().Unix() > data.ExpiresAt {
		s.redisClient.Delete(ctx, key)
		return nil, entities.Err2FACodeInvalidOrExpired
	}

	return &data, nil
}

func (s *sessionService) Verify2FACode(ctx context.Context, userID uuid.UUID, code string, context entities.TwoFASessionContext) (*entities.TwoFASessionData, error) {
	data, err := s.Get2FAData(ctx, userID, context)
	if err != nil {
		return nil, err
	}

	if data.Code != code {
		return nil, entities.Err2FACodeInvalid
	}

	// Удаляем код после успешной проверки
	if err := s.Delete2FACode(ctx, userID, context); err != nil {
		log.Printf("Failed to delete 2FA code: %v", err)
	}

	return data, nil
}
