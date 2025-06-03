package entities

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SubTask struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;primaryKey"`
	Title       string    `json:"title" gorm:"not null"`
	Description string    `json:"description" gorm:"not null"`
	TaskID      uuid.UUID `json:"task_id" gorm:"type:uuid"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (SubTask) TableName() string {
	return "sub_tasks"
}

func (u *SubTask) BeforeCreate(tx *gorm.DB) error {
	u.ID = uuid.New()
	u.CreatedAt = time.Now()
	u.UpdatedAt = time.Now()
	return nil
}

func (u *SubTask) BeforeUpdate(tx *gorm.DB) error {
	u.UpdatedAt = time.Now()
	return nil
}
