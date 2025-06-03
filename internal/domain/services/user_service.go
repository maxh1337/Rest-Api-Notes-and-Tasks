package services

import (
	"rest-api-notes/internal/domain/entities"
	"rest-api-notes/internal/domain/repositories"
)

type UserService interface {
	GetUserProfile(userId string) (*entities.User, error)
}

type userService struct {
	userRepo repositories.UserRepository
}

func NewUserService(userRepo repositories.UserRepository) UserService {
	return &userService{
		userRepo: userRepo,
	}
}

func (s *userService) GetUserProfile(userId string) (*entities.User, error) {
	user, err := s.userRepo.GetUserById(userId)
	if err != nil {
		return nil, err
	}

	return user, nil
}
