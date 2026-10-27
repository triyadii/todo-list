package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TaskStatus string

const (
	TaskStatusPending    TaskStatus = "pending"
	TaskStatusInProgress TaskStatus = "in_progress"
	TaskStatusCompleted  TaskStatus = "completed"
	TaskStatusCancelled  TaskStatus = "cancelled"
)

type TaskPriority string

const (
	TaskPriorityLow    TaskPriority = "low"
	TaskPriorityMedium TaskPriority = "medium"
	TaskPriorityHigh   TaskPriority = "high"
)

type Task struct {
	ID          uuid.UUID    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID      uuid.UUID    `gorm:"type:uuid;index;not null" json:"user_id"`
	Title       string       `gorm:"type:varchar(255);not null" json:"title"`
	Description *string      `gorm:"type:text" json:"description,omitempty"`
	DueDate     string       `gorm:"type:date;not null" json:"due_date"`    // Format: YYYY-MM-DD
	StartTime   *string      `gorm:"type:time" json:"start_time,omitempty"` // Format: HH:MM:SS
	EndTime     *string      `gorm:"type:time" json:"end_time,omitempty"`   // Format: HH:MM:SS
	Status      TaskStatus   `gorm:"type:varchar(50);default:'pending';not null" json:"status"`
	Priority    TaskPriority `gorm:"type:varchar(50);default:'medium';not null" json:"priority"`
	CreatedAt   time.Time    `gorm:"type:timestamptz;autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time    `gorm:"type:timestamptz;autoUpdateTime" json:"updated_at"`

	User      *User      `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"user,omitempty"`
	Reminders []Reminder `gorm:"foreignKey:TaskID;constraint:OnDelete:CASCADE" json:"reminders,omitempty"`
}

func (t *Task) BeforeCreate(tx *gorm.DB) (err error) {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	if t.Status == "" {
		t.Status = TaskStatusPending
	}
	if t.Priority == "" {
		t.Priority = TaskPriorityMedium
	}
	return
}
