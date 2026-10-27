package repository

import (
	"todo-backend/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ChatRepository interface {
	Create(msg *models.ChatMessage) error
	FindByUserID(userID uuid.UUID, limit int) ([]models.ChatMessage, error)
	DeleteByUserID(userID uuid.UUID) error
}

type chatRepository struct {
	db *gorm.DB
}

func NewChatRepository(db *gorm.DB) ChatRepository {
	return &chatRepository{db: db}
}

func (r *chatRepository) Create(msg *models.ChatMessage) error {
	return r.db.Create(msg).Error
}

func (r *chatRepository) FindByUserID(userID uuid.UUID, limit int) ([]models.ChatMessage, error) {
	var messages []models.ChatMessage
	query := r.db.Where("user_id = ?", userID).Order("created_at ASC")
	if limit > 0 {
		// When limit is provided, fetch last N messages
		var count int64
		r.db.Model(&models.ChatMessage{}).Where("user_id = ?", userID).Count(&count)
		if count > int64(limit) {
			query = query.Offset(int(count) - limit)
		}
	}
	if err := query.Find(&messages).Error; err != nil {
		return nil, err
	}
	return messages, nil
}

func (r *chatRepository) DeleteByUserID(userID uuid.UUID) error {
	return r.db.Where("user_id = ?", userID).Delete(&models.ChatMessage{}).Error
}
