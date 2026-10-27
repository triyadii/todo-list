package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Name         string    `gorm:"type:varchar(255);not null" json:"name"`
	Email        string    `gorm:"type:varchar(255);uniqueIndex;not null" json:"email"`
	PasswordHash string    `gorm:"type:varchar(255);not null" json:"-"`
	Timezone     string    `gorm:"type:varchar(100);default:'Asia/Jakarta';not null" json:"timezone"`
	FCMToken     *string   `gorm:"type:text" json:"fcm_token,omitempty"`
	CreatedAt    time.Time `gorm:"type:timestamptz;autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time `gorm:"type:timestamptz;autoUpdateTime" json:"updated_at"`

	Tasks        []Task        `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"tasks,omitempty"`
	ChatMessages []ChatMessage `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"chat_messages,omitempty"`
}

func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	if u.Timezone == "" {
		u.Timezone = "Asia/Jakarta"
	}
	return
}
