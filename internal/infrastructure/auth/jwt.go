package auth

import (
	"errors"
	"rest-api-notes/internal/config"
	"rest-api-notes/internal/domain/entities"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var (
	ErrInvalidToken      = errors.New("invalid token")
	ErrSigningToken      = errors.New("token signing error")
	CookieTokenAccess    = "accessToken"
	CookieTokenRefresh   = "refreshToken"
	CookieTokenSessionID = "session_id"
)

type jwtService struct {
	cfg config.JWTConfig
}

// JWTType represents the type of JWT token.
type JWTType string

const (
	JWTTypeAccess  JWTType = "access"
	JWTTypeRefresh JWTType = "refresh"
)

type JWTService interface {
	GenerateToken(userID uuid.UUID, sessionID string, jwtType JWTType) (string, error)
	ValidateToken(tokenString string, jwtType JWTType) (*entities.JWTClaims, error)
	// ValidateAccessToken(tokenString string) (*entities.Auth, error)
	// ValidateRefreshToken(tokenString string) (*entities.Auth, error)
}

func NewJWTService(config config.JWTConfig) JWTService {
	return &jwtService{cfg: config}
}

func (s *jwtService) GenerateToken(userID uuid.UUID, sessionID string, jwtType JWTType) (string, error) {
	var expiresAt time.Time
	if jwtType == "access" {
		expiresAt = time.Now().Add(time.Duration(s.cfg.JWT_ACCESS_EXPIRATION) * time.Hour)
	} else {
		expiresAt = time.Now().Add(time.Duration(s.cfg.JWT_REFRESH_EXPIRATION) * time.Hour)
	}

	claims := jwt.MapClaims{
		"user_id":    userID,
		"session_id": sessionID,
		"exp":        expiresAt.Unix(),
		"type":       jwtType,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(s.cfg.JWTSecret))
	if err != nil {
		return "", ErrSigningToken
	}

	return signedToken, nil
}

func (s *jwtService) ValidateToken(tokenString string, jwtType JWTType) (*entities.JWTClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(s.cfg.JWTSecret), nil
	})
	if err != nil {
		return nil, ErrInvalidToken
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userID, _ := uuid.Parse(claims["user_id"].(string))
		sessionID, _ := claims["session_id"].(string)
		exp, _ := claims["exp"].(float64)
		tokenType, _ := claims["type"].(string)

		if jwtType == "refresh" {
			if tokenType != "refresh" {
				return nil, ErrInvalidToken
			}

			if time.Now().After(time.Unix(int64(exp), 0)) {
				return nil, ErrInvalidToken
			}
		}

		return &entities.JWTClaims{
			UserID:    userID,
			SessionID: sessionID,
			ExpiresAt: time.Unix(int64(exp), 0),
			TokenType: tokenType,
		}, nil
	}
	return nil, ErrInvalidToken
}
