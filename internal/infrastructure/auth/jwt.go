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
	ErrInvalidAccessToken     = errors.New("invalid access token")
	ErrInvalid2FAToken        = errors.New("invalid 2FA token")
	ErrInvalid2FATokenExpired = errors.New("2FA token expired")
	ErrInvalidRefreshToken    = errors.New("invalid refresh token")
	ErrSigningToken           = errors.New("token signing error")
	CookieTokenAccess         = "accessToken"
	CookieToken2Fa            = "2fa_token"
	CookieTokenRefresh        = "refreshToken"
	CookieTokenSessionID      = "session_id"
)

type jwtService struct {
	cfg config.JWTConfig
}

// JWTType represents the type of JWT token.
type JWTType string

const (
	JWTTypeAccess  JWTType = "access"
	JWTTypeRefresh JWTType = "refresh"
	JWTType2FA     JWTType = "2fa"
)

type JWTService interface {
	GenerateToken(userID uuid.UUID, sessionID string, jwtType JWTType) (string, error)
	ValidateToken(tokenString string, jwtType JWTType) (*entities.JWTClaims, error)
	Generate2FAToken(userID uuid.UUID, userAgent, userIp string) (string, error)
	Validate2FAToken(tokenString string) (*entities.TwoFactorJWTClaims, error)
}

func NewJWTService(config config.JWTConfig) JWTService {
	return &jwtService{cfg: config}
}

func (s *jwtService) GenerateToken(userID uuid.UUID, sessionID string, jwtType JWTType) (string, error) {
	var expiresAt time.Time
	if jwtType == "access" {
		expiresAt = time.Now().Add(time.Duration(s.cfg.JWT_ACCESS_EXPIRATION) * time.Hour)
	} else if jwtType == "refresh" {
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
	var currentTokenError error
	if jwtType == "refresh" {
		currentTokenError = ErrInvalidRefreshToken
	} else if jwtType == "access" {
		currentTokenError = ErrInvalidAccessToken
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(s.cfg.JWTSecret), nil
	})
	if err != nil {
		return nil, currentTokenError
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userID, _ := uuid.Parse(claims["user_id"].(string))
		sessionID, _ := claims["session_id"].(string)
		exp, _ := claims["exp"].(float64)
		tokenType, _ := claims["type"].(string)

		if jwtType == "refresh" {
			if tokenType != "refresh" {
				return nil, currentTokenError
			}

			if time.Now().After(time.Unix(int64(exp), 0)) {
				return nil, currentTokenError
			}
		}

		return &entities.JWTClaims{
			UserID:    userID,
			SessionID: sessionID,
			ExpiresAt: time.Unix(int64(exp), 0),
			TokenType: tokenType,
		}, nil
	}
	return nil, currentTokenError
}

func (s *jwtService) Generate2FAToken(userID uuid.UUID, userAgent, userIp string) (string, error) {
	expiresAt := time.Now().Add(time.Duration(s.cfg.JWT_2FA_EXPIRATION) * time.Minute)

	claims := jwt.MapClaims{
		"user_id":    userID,
		"user_agent": userAgent,
		"user_ip":    userIp,
		"exp":        expiresAt.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(s.cfg.JWTSecret))
	if err != nil {
		return "", ErrSigningToken
	}

	return signedToken, nil
}

func (s *jwtService) Validate2FAToken(tokenString string) (*entities.TwoFactorJWTClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(s.cfg.JWTSecret), nil
	})
	if err != nil {
		return nil, ErrInvalid2FAToken
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userID, _ := uuid.Parse(claims["user_id"].(string))
		userAgent, _ := claims["user_agent"].(string)
		userIP, _ := claims["user_ip"].(string)
		exp, _ := claims["exp"].(float64)

		if time.Now().After(time.Unix(int64(exp), 0)) {
			return nil, ErrInvalid2FATokenExpired
		}

		return &entities.TwoFactorJWTClaims{
			UserID:    userID,
			UserAgent: userAgent,
			UserIP:    userIP,
			ExpiresAt: time.Unix(int64(exp), 0),
		}, nil
	}
	return nil, ErrInvalid2FAToken
}
