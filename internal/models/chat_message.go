package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ChatRole string

const (
	ChatRoleUser      ChatRole = "user"
	ChatRoleAssistant ChatRole = "assistant"
	ChatRoleSystem    ChatRole = "system"
)

type ChatMessage struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID    uuid.UUID `gorm:"type:uuid;index;not null" json:"user_id"`
	Role      ChatRole  `gorm:"type:varchar(50);not null" json:"role"`
	Content   string    `gorm:"type:text;not null" json:"content"`
	CreatedAt time.Time `gorm:"type:timestamptz;index;autoCreateTime" json:"created_at"`

	User *User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"user,omitempty"`
}

func (c *ChatMessage) BeforeCreate(tx *gorm.DB) (err error) {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return
}
