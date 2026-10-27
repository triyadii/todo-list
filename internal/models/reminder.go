package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Reminder struct {
	ID        uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TaskID    uuid.UUID  `gorm:"type:uuid;index;not null" json:"task_id"`
	RemindAt  time.Time  `gorm:"type:timestamptz;index;not null" json:"remind_at"`
	IsSent    bool       `gorm:"type:boolean;default:false;index;not null" json:"is_sent"`
	AIMessage *string    `gorm:"type:text" json:"ai_message,omitempty"`
	SentAt    *time.Time `gorm:"type:timestamptz" json:"sent_at,omitempty"`
	CreatedAt time.Time  `gorm:"type:timestamptz;autoCreateTime" json:"created_at"`
	UpdatedAt time.Time  `gorm:"type:timestamptz;autoUpdateTime" json:"updated_at"`

	Task *Task `gorm:"foreignKey:TaskID;constraint:OnDelete:CASCADE" json:"task,omitempty"`
}

func (r *Reminder) BeforeCreate(tx *gorm.DB) (err error) {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	return
}
