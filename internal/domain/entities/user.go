package entities

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrEmailAlreadyTaken    = errors.New("email already taken")
	ErrUsernameAlreadyTaken = errors.New("username already taken")
	ErrUserNotFound         = errors.New("user not found")
	ErrInvalidEmailFormat   = errors.New("invalid email format")
	ErrCantChangePhone2FA   = errors.New("you cant change phone number while 2FA active")
)

type User struct {
	ID               uuid.UUID `json:"id" gorm:"type:uuid;primaryKey"`
	Username         string    `json:"username" gorm:"unique;not null"`
	Email            string    `json:"email" gorm:"unique;not null"`
	PhoneNumber      string    `json:"phone_number" gorm:"unique"`
	TwoFactorEnabled bool      `json:"two_factor_enabled" gorm:"default:false"`
	Password         string    `json:"-" gorm:"not null"`
	Role             RoleType  `json:"role" gorm:"default:'user';not null"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
	Tasks            []Task    `json:"tasks" gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
}

func (User) TableName() string {
	return "users"
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	u.ID = uuid.New()
	u.CreatedAt = time.Now()
	u.UpdatedAt = time.Now()
	u.TwoFactorEnabled = false
	return nil
}

func (u *User) BeforeUpdate(tx *gorm.DB) error {
	u.UpdatedAt = time.Now()
	return nil
}

type UserUpdatePhoneReq struct {
	PhoneNumber string `json:"phone_number" validate:"required,phone"`
}

type Verify2FACodeReq struct {
	Code string `json:"code" validate:"required"`
}
