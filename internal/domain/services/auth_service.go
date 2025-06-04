package services

import (
	"context"
	"log"
	"rest-api-notes/internal/domain/entities"
	"rest-api-notes/internal/domain/repositories"

	"rest-api-notes/internal/infrastructure/auth"

	"github.com/google/uuid"
)

type authService struct {
	jS auth.JWTService
	pS auth.PasswordService
	uR repositories.UserRepository
	sS SessionService
}

type AuthService interface {
	Login(ctx context.Context, req *entities.UserLoginReq, userAgent, userIp string) (entities.UserAuthRes, *entities.Session, error)
	Register(ctx context.Context, req *entities.UserRegisterReq, userAgent, userIp string) (entities.UserAuthRes, *entities.Session, error)
	Logout(ctx context.Context, req *entities.UserLogoutReq) error
}

func NewAuthService(jS auth.JWTService, pS auth.PasswordService,
	uR repositories.UserRepository, rS SessionService) AuthService {
	return &authService{jS: jS, pS: pS, uR: uR, sS: rS}
}

func (s *authService) Login(ctx context.Context,
	req *entities.UserLoginReq, userAgent, userIp string) (entities.UserAuthRes, *entities.Session, error) {
	err := s.pS.ValidatePassword(req.Password)
	if err != nil {
		return entities.UserAuthRes{}, &entities.Session{}, err
	}

	user, err := s.uR.GetUserByEmail(req.Identifier)
	if err != nil {
		user, err = s.uR.GetUserByUsername(req.Identifier)
		if err != nil {
			// Обе попытки неудачны
			return entities.UserAuthRes{}, &entities.Session{}, err
		}
	}

	err = s.pS.ComparePasswords(user.Password, req.Password)
	if err != nil {
		return entities.UserAuthRes{}, &entities.Session{}, err
	}

	res := entities.UserAuthRes{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
		Role:     entities.RoleType(user.Role.String()),
	}

	sessionID := uuid.New().String()
	accToken, err := s.jS.GenerateToken(res.ID, sessionID, "access")
	if err != nil {
		return entities.UserAuthRes{}, &entities.Session{}, entities.ErrFailedToCreateAccessToken
	}

	refToken, err := s.jS.GenerateToken(res.ID, sessionID, "refresh")
	if err != nil {
		return entities.UserAuthRes{}, &entities.Session{}, entities.ErrFailedToCreateRefreshToken
	}

	session, err := s.sS.CreateSession(ctx, res.ID, sessionID, accToken, refToken, userAgent, userIp)
	if err != nil {
		return entities.UserAuthRes{}, &entities.Session{}, err
	}

	return res, session, nil
}

func (s *authService) Register(ctx context.Context,
	req *entities.UserRegisterReq, userAgent, userIp string) (entities.UserAuthRes, *entities.Session, error) {
	err := s.pS.ValidatePassword(req.Password)
	if err != nil {
		return entities.UserAuthRes{}, &entities.Session{}, err
	}

	password, err := s.pS.HashPassword(req.Password)
	if err != nil {
		return entities.UserAuthRes{}, &entities.Session{}, err
	}
	req.Password = password

	user, err := s.uR.Create(req)
	if err != nil {
		return entities.UserAuthRes{}, &entities.Session{}, err
	}

	res := entities.UserAuthRes{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
		Role:     entities.RoleType(user.Role.String()),
	}

	sessionID := uuid.New().String()
	accToken, err := s.jS.GenerateToken(res.ID, sessionID, "access")
	if err != nil {
		return entities.UserAuthRes{}, &entities.Session{}, entities.ErrFailedToCreateAccessToken
	}

	refToken, err := s.jS.GenerateToken(res.ID, sessionID, "refresh")
	if err != nil {
		return entities.UserAuthRes{}, &entities.Session{}, entities.ErrFailedToCreateRefreshToken
	}

	session, err := s.sS.CreateSession(ctx, res.ID, sessionID, accToken, refToken, userAgent, userIp)
	if err != nil {
		return entities.UserAuthRes{}, &entities.Session{}, err
	}

	return res, session, nil
}

func (s *authService) Logout(ctx context.Context, req *entities.UserLogoutReq) error {
	refreshClaim, err := s.jS.ValidateToken(req.RefreshToken, "refresh")
	if err != nil {
		return nil
	}
	log.Printf("refreshClaim - %+v", refreshClaim)

	if refreshClaim.SessionID != req.SessionID {
		return nil
	}

	isSessionValid, err := s.sS.IsSessionValid(ctx, refreshClaim.UserID, refreshClaim.SessionID, req.RefreshToken)
	if err != nil {
		return err
	}

	if isSessionValid {
		_ = s.sS.DeleteSession(ctx, refreshClaim.UserID, refreshClaim.SessionID)
	}

	return nil
}
