package repositories

import (
	"rest-api-notes/internal/domain/entities"

	"gorm.io/gorm"
)

type userRepository struct {
	db *gorm.DB
}

type UserRepository interface {
	Create(user entities.UserRegisterReq) (*entities.User, error)
	GetUserById(userId string) (*entities.User, error)
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(dto entities.UserRegisterReq) (*entities.User, error) {
	user := entities.User{
		Username: dto.Username,
		Email:    dto.Email,
		Password: dto.Password, //encription here argon2
	}

	if err := r.db.Create(&user).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *userRepository) GetUserById(userId string) (*entities.User, error) {
	var user entities.User
	if err := r.db.First(&user, "id = ?", userId).Error; err != nil {
		return nil, err
	}

	return &user, nil
}
