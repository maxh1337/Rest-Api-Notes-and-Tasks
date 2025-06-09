package services

import (
	"context"
	"rest-api-notes/internal/domain/entities"
	"rest-api-notes/internal/domain/repositories"

	"github.com/google/uuid"
)

type UserService interface {
	GetUserProfile(userId uuid.UUID) (*entities.User, error)
	UpdateUserPhone(ctx context.Context, req *entities.UserUpdatePhoneReq, userID uuid.UUID) error
	TwoFactorToggleRequest(ctx context.Context, userID uuid.UUID) error
	VerifyTwoFactorToggleRequest(ctx context.Context, userID uuid.UUID, code string) error
	ResendTwoFactorCode(ctx context.Context, userID uuid.UUID) error
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

func (s *userService) TwoFactorToggleRequest(ctx context.Context, userID uuid.UUID) error {
	user, err := s.userRepo.GetUserById(userID)
	if err != nil {
		return err
	}

	if user.PhoneNumber == "" {
		return entities.ErrNoPhoneNumberToEnable2FA
	}

	code, err := s.twoFactorService.Generate2FACode()
	if err != nil {
		return err
	}

	if err := s.sessionService.Save2FACode(ctx, userID, code, "toggle"); err != nil {
		return err
	}

	if err := s.twoFactorService.MakeRequestToGateway(ctx, user.PhoneNumber, code); err != nil {
		if err := s.sessionService.Delete2FACode(ctx, userID, "toggle"); err != nil {
			return err
		}
		return err
	}

	return nil
}

func (s *userService) VerifyTwoFactorToggleRequest(ctx context.Context, userID uuid.UUID, code string) error {
	_, err := s.userRepo.GetUserById(userID)
	if err != nil {
		return err
	}

	if _, err := s.sessionService.Verify2FACode(ctx, userID, code, "toggle"); err != nil {
		return err
	}

	if err := s.userRepo.ToggleUser2FA(userID); err != nil {
		return err
	}

	return nil
}

func (s *userService) ResendTwoFactorCode(ctx context.Context, userID uuid.UUID) error {
	user, err := s.userRepo.GetUserById(userID)
	if err != nil {
		return err
	}

	if user.PhoneNumber == "" {
		return entities.ErrNoPhoneNumberToEnable2FA
	}

	code, err := s.twoFactorService.Generate2FACode()
	if err != nil {
		return err
	}

	if err := s.sessionService.Save2FACode(ctx, userID, code, "toggle"); err != nil {
		return err
	}

	if err := s.twoFactorService.MakeRequestToGateway(ctx, user.PhoneNumber, code); err != nil {
		if err := s.sessionService.Delete2FACode(ctx, userID, "toggle"); err != nil {
			return err
		}
		return err
	}

	return nil
}
