package entities

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Task struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;primaryKey"`
	User        User      `json:"user" gorm:"foreignKey:UserID"`
	UserID      uuid.UUID `json:"task_id" gorm:"type:uuid;not null"`
	Title       string    `json:"title" gorm:"not null"`
	Description string    `json:"description" gorm:"not null"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`
	SubTasks    []SubTask `json:"sub_tasks" gorm:"foreignKey:TaskID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
}

func (Task) TableName() string {
	return "tasks"
}

func (u *Task) BeforeCreate(tx *gorm.DB) error {
	u.ID = uuid.New()
	u.CreatedAt = time.Now()
	u.UpdatedAt = time.Now()
	return nil
}

func (u *Task) BeforeUpdate(tx *gorm.DB) error {
	u.UpdatedAt = time.Now()
	return nil
}
