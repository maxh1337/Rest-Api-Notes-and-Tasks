package services

import (
	"context"
	"rest-api-notes/internal/domain/entities"
	"rest-api-notes/internal/domain/repositories"
	"slices"

	"rest-api-notes/internal/infrastructure/auth"

	"github.com/google/uuid"
)

type authService struct {
	jS               auth.JWTService
	pS               auth.PasswordService
	uR               repositories.UserRepository
	sS               SessionService
	twoFactorService TwoFactorService
}

type AuthService interface {
	Login(ctx context.Context, req *entities.UserLoginReq, userAgent, userIp string) (*entities.UserAuthRes, *entities.Session, error)
	Register(ctx context.Context, req *entities.UserRegisterReq, userAgent, userIp string) (*entities.UserAuthRes, *entities.Session, error)
	Logout(ctx context.Context, req *entities.UserLogoutReq) error
	GetNewTokens(ctx context.Context, req *entities.UserGetNewTokensReq, userAgent, userIp string) (*entities.Session, error)
	CreateNewSessionAndTokens(ctx context.Context, userId uuid.UUID, userAgent, userIp string) (*entities.Session, error)
	Verify2FACode(ctx context.Context, code, token, userAgent, userIP string, userID uuid.UUID) (*entities.Session, error)
	Resend2FACode(ctx context.Context, userID uuid.UUID, userAgent, userIP string) (*string, error)
}

func NewAuthService(jS auth.JWTService, pS auth.PasswordService,
	uR repositories.UserRepository, rS SessionService, twoFactorService TwoFactorService) AuthService {
	return &authService{jS: jS, pS: pS, uR: uR, sS: rS, twoFactorService: twoFactorService}
}

func (s *authService) Login(ctx context.Context,
	req *entities.UserLoginReq, userAgent, userIp string) (*entities.UserAuthRes, *entities.Session, error) {
	//
	user, err := s.uR.FindUserByEmailOrUsername(req.Identifier)
	if err != nil {
		return nil, nil, err
	}

	err = s.pS.ValidatePassword(req.Password)
	if err != nil {
		return nil, nil, err
	}

	err = s.pS.ComparePasswords(user.Password, req.Password)
	if err != nil {
		return nil, nil, err
	}

	userSessions, err := s.sS.GetAllUserSessions(ctx, user.ID)
	if err != nil {
		return nil, nil, err
	}

	if userSessions != nil && len(*userSessions) > 0 {
		index := slices.IndexFunc(*userSessions, func(session entities.Session) bool {
			return session.UserAgent == userAgent && session.IP == userIp
		})
		// Можно добавить логику с ограничением общего количества сессий, а так же удалением самой старой

		if index != -1 {
			s.sS.DeleteSession(ctx, (*userSessions)[index].UserID, (*userSessions)[index].SessionID)
		}
	}

	res := entities.UserAuthRes{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
		Role:     entities.RoleType(user.Role.String()),
	}

	// Проверка на то, включена ли 2FA, если да -> что-то делаем и не отдаем сессию.
	if user.TwoFactorEnabled {
		code, err := s.twoFactorService.Generate2FACode()
		if err != nil {
			return nil, nil, err
		}
		token, err := s.jS.Generate2FAToken(user.ID, userAgent, userIp)
		if err != nil {
			return nil, nil, err
		}

		if err := s.sS.Save2FACode(ctx, user.ID, code, "login", token); err != nil {
			return nil, nil, err
		}
		if err = s.twoFactorService.MakeRequestToGateway(ctx, user.PhoneNumber, code); err != nil {
			return nil, nil, err
		}

		// Создание cookie только для 2FA с userID для последующей идентификации пользователя

		return &entities.UserAuthRes{TwoFactorToken: token}, nil, entities.Err2FARequired
	}

	session, err := s.CreateNewSessionAndTokens(ctx, res.ID, userAgent, userIp)
	if err != nil {
		return nil, nil, err
	}

	return &res, session, nil

}

func (s *authService) Register(ctx context.Context,
	req *entities.UserRegisterReq, userAgent, userIp string) (*entities.UserAuthRes, *entities.Session, error) {
	//
	if err := s.uR.CheckUserUniqueness(req.Email, req.Username); err != nil {
		return nil, nil, err
	}

	err := s.pS.ValidatePassword(req.Password)
	if err != nil {
		return nil, nil, err
	}

	password, err := s.pS.HashPassword(req.Password)
	if err != nil {
		return nil, nil, err
	}
	req.Password = password

	user, err := s.uR.Create(req)
	if err != nil {
		return nil, nil, err
	}

	res := entities.UserAuthRes{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
		Role:     entities.RoleType(user.Role.String()),
	}

	session, err := s.CreateNewSessionAndTokens(ctx, res.ID, userAgent, userIp)
	if err != nil {
		return nil, nil, err
	}

	return &res, session, nil
}

func (s *authService) Verify2FACode(ctx context.Context, code, token, userAgent, userIP string, userID uuid.UUID) (*entities.Session, error) {
	data, err := s.sS.Verify2FACode(ctx, userID, code, "login")
	if err != nil {
		return nil, err
	}

	if data.Token != token {
		return nil, entities.Err2FASessionAndTokenMismatch
	}

	session, err := s.CreateNewSessionAndTokens(ctx, userID, userAgent, userIP)
	if err != nil {
		return nil, err
	}

	return session, nil
}

func (s *authService) Resend2FACode(ctx context.Context, userID uuid.UUID, userAgent, userIP string) (*string, error) {
	user, err := s.uR.GetUserById(userID)
	if err != nil {
		return nil, err
	}

	if user.TwoFactorEnabled {
		code, err := s.twoFactorService.Generate2FACode()
		if err != nil {
			return nil, err
		}
		token, err := s.jS.Generate2FAToken(user.ID, userAgent, userIP)
		if err != nil {
			return nil, err
		}

		if err := s.sS.Save2FACode(ctx, user.ID, code, "login", token); err != nil {
			return nil, err
		}
		if err = s.twoFactorService.MakeRequestToGateway(ctx, user.PhoneNumber, code); err != nil {
			return nil, err
		}

		return &token, nil
	}
	return nil, entities.Err2FADisabled
}

func (s *authService) Logout(ctx context.Context, req *entities.UserLogoutReq) error {
	refreshClaim, err := s.jS.ValidateToken(req.RefreshToken, "refresh")
	if err != nil {
		return err
	}

	if refreshClaim.SessionID != req.SessionID {
		return nil
	}

	isSessionValid, err := s.sS.IsSessionValid(ctx, refreshClaim.UserID, refreshClaim.SessionID, req.RefreshToken)
	if err != nil {
		return err
	}

	if isSessionValid {
		err = s.sS.DeleteSession(ctx, refreshClaim.UserID, refreshClaim.SessionID)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *authService) GetNewTokens(ctx context.Context,
	req *entities.UserGetNewTokensReq, userAgent, userIp string) (*entities.Session, error) {
	// Получаем refresh токен -> Проверяем его валидность -> Проверяем сессию ->
	// -> Сверяем refresh токен с сессией и UserAgent && IP -> Перевыпускаем новые токены -> Обновляем сессию
	refreshClaim, err := s.jS.ValidateToken(req.RefreshToken, "refresh")
	if err != nil {
		return nil, err
	}

	if refreshClaim.SessionID != req.SessionID {
		return nil, auth.ErrInvalidRefreshToken
	}

	isValid, err := s.sS.IsSessionValid(ctx, refreshClaim.UserID, refreshClaim.SessionID, req.RefreshToken)
	if err != nil {
		return nil, err
	}

	if !isValid {
		return nil, entities.ErrSessionExpired
	}

	prevSession, err := s.sS.GetSession(ctx, refreshClaim.UserID, refreshClaim.SessionID)
	if err != nil {
		return nil, err
	}

	if prevSession.RefreshToken != req.RefreshToken {
		return nil, entities.ErrTokensMismatch
	}

	if prevSession.UserAgent != userAgent || prevSession.IP != userIp {
		return nil, entities.ErrSessionBelongsToAnotherDevice
	}

	if err = s.sS.DeleteSession(ctx, refreshClaim.UserID, refreshClaim.SessionID); err != nil {
		return nil, err
	}

	newSession, err := s.CreateNewSessionAndTokens(ctx, refreshClaim.UserID, userAgent, userIp)
	if err != nil {
		return nil, err
	}

	return newSession, nil
}

func (s *authService) CreateNewSessionAndTokens(ctx context.Context, userId uuid.UUID, userAgent, userIP string) (*entities.Session, error) {
	sessionID := uuid.New().String()

	accToken, err := s.jS.GenerateToken(userId, sessionID, "access")
	if err != nil {
		return nil, entities.ErrFailedToCreateAccessToken
	}

	refToken, err := s.jS.GenerateToken(userId, sessionID, "refresh")
	if err != nil {
		return nil, entities.ErrFailedToCreateRefreshToken
	}

	session, err := s.sS.CreateSession(ctx, userId, sessionID, accToken, refToken, userAgent, userIP)
	if err != nil {
		return nil, err
	}

	return session, nil
}
