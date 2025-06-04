package repositories

import (
	"errors"
	"regexp"
	"rest-api-notes/internal/domain/entities"
	"rest-api-notes/internal/infrastructure/auth"
	"strings"

	"gorm.io/gorm"
)

type userRepository struct {
	db *gorm.DB
	ps auth.PasswordService
}

type UserRepository interface {
	Create(dto *entities.UserRegisterReq) (*entities.User, error)
	GetUserById(userId string) (*entities.User, error)
	GetUserByUsername(username string) (*entities.User, error)
	GetUserByEmail(email string) (*entities.User, error)
}

func NewUserRepository(db *gorm.DB, ps auth.PasswordService) UserRepository {
	return &userRepository{db: db, ps: ps}
}

func (r *userRepository) Create(dto *entities.UserRegisterReq) (*entities.User, error) {
	user := entities.User{
		Username: dto.Username,
		Email:    dto.Email,
		Password: dto.Password,
	}

	user.Email = strings.ToLower(strings.TrimSpace(user.Email))
	user.Username = strings.TrimSpace(user.Username)

	emailRegex := regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,}$`)
	if !emailRegex.MatchString(user.Email) {
		return nil, entities.ErrInvalidEmailFormat
	}

	if _, err := r.GetUserByEmail(user.Email); err == nil {
		return nil, entities.ErrEmailAlreadyTaken
	} else if !errors.Is(err, entities.ErrUserNotFound) {
		return nil, err
	}

	if _, err := r.GetUserByUsername(user.Username); err == nil {
		return nil, entities.ErrUsernameAlreadyTaken
	} else if !errors.Is(err, entities.ErrUserNotFound) {
		return nil, err
	}

	if err := r.db.Create(&user).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *userRepository) GetUserById(userId string) (*entities.User, error) {
	var user entities.User
	if err := r.db.First(&user, "id = ?", userId).Error; err != nil {
		return nil, entities.ErrUserNotFound
	}

	return &user, nil
}

func (r *userRepository) GetUserByUsername(username string) (*entities.User, error) {
	var user entities.User
	if err := r.db.Where("username = ?", username).First(&user).Error; err != nil {
		return nil, entities.ErrUserNotFound
	}

	return &user, nil
}

func (r *userRepository) GetUserByEmail(email string) (*entities.User, error) {
	var user entities.User
	if err := r.db.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, entities.ErrUserNotFound
	}

	return &user, nil
}
