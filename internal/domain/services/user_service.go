package services

import (
	"context"
	"log"
	"rest-api-notes/internal/domain/entities"
	"rest-api-notes/internal/domain/repositories"

	"github.com/google/uuid"
)

type UserService interface {
	GetUserProfile(userId uuid.UUID) (*entities.User, error)
	UpdateUserPhone(ctx context.Context, req *entities.UserUpdatePhoneReq, userID uuid.UUID) error
	TwoFactorEnableRequest(ctx context.Context, userID uuid.UUID) error
	VerifyTwoFactorEnableRequest(ctx context.Context, userID uuid.UUID, code string) error
}

type userService struct {
	userRepo         repositories.UserRepository
	sessionService   SessionService
	twoFactorService TwoFactorService
}

func NewUserService(userRepo repositories.UserRepository,
	sessionService SessionService, twoFactorService TwoFactorService) UserService {
	return &userService{
		userRepo:         userRepo,
		sessionService:   sessionService,
		twoFactorService: twoFactorService,
	}
}

func (s *userService) GetUserProfile(userId uuid.UUID) (*entities.User, error) {
	user, err := s.userRepo.GetUserById(userId)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *userService) UpdateUserPhone(ctx context.Context, req *entities.UserUpdatePhoneReq, userID uuid.UUID) error {
	user, err := s.userRepo.GetUserById(userID)
	if err != nil {
		return err
	}

	if user.TwoFactorEnabled {
		return entities.ErrCantChangePhone2FA
	}

	if err = s.userRepo.UpdatePhoneNumber(req.PhoneNumber, userID); err != nil {
		return err
	}

	return nil
}

func (s *userService) TwoFactorEnableRequest(ctx context.Context, userID uuid.UUID) error {
	user, err := s.userRepo.GetUserById(userID)
	if err != nil {
		return err
	}

	if user.TwoFactorEnabled {
		return entities.Err2FAAlreadyEnabled
	}

	code, err := s.twoFactorService.Generate2FACode()
	if err != nil {
		return err
	}

	log.Printf("UserID - %v. Code - %s", userID, code)

	if err := s.sessionService.Save2FACode(ctx, userID, code, "enable"); err != nil {
		return err
	}

	if err := s.twoFactorService.MakeRequestToGateway(ctx, user.PhoneNumber, code); err != nil {
		if err := s.sessionService.Delete2FACode(ctx, userID, "enable"); err != nil {
			return err
		}
		return err
	}

	return nil
}

func (s *userService) VerifyTwoFactorEnableRequest(ctx context.Context, userID uuid.UUID, code string) error {
	user, err := s.userRepo.GetUserById(userID)
	if err != nil {
		return err
	}

	if user.TwoFactorEnabled {
		return entities.Err2FAAlreadyEnabled
	}

	if _, err := s.sessionService.Verify2FACode(ctx, userID, code, "enable"); err != nil {
		return err
	}

	if err := s.userRepo.EnableUser2FA(userID); err != nil {
		return err
	}

	return nil
}
