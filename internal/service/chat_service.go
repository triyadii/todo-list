package service

import (
	"errors"

	"todo-backend/internal/models"
	"todo-backend/internal/repository"

	"github.com/google/uuid"
)

type ChatService interface {
	SendMessage(userID uuid.UUID, req *SendMessageRequest) (*models.ChatMessage, error)
	GetHistory(userID uuid.UUID, limit int) ([]models.ChatMessage, error)
	ClearHistory(userID uuid.UUID) error
}

type chatService struct {
	chatRepo repository.ChatRepository
}

func NewChatService(chatRepo repository.ChatRepository) ChatService {
	return &chatService{chatRepo: chatRepo}
}

type SendMessageRequest struct {
	Role    models.ChatRole `json:"role"`
	Content string          `json:"content"`
}

func (s *chatService) SendMessage(userID uuid.UUID, req *SendMessageRequest) (*models.ChatMessage, error) {
	if req.Content == "" {
		return nil, errors.New("content is required")
	}

	role := req.Role
	if role == "" {
		role = models.ChatRoleUser
	}
	if role != models.ChatRoleUser && role != models.ChatRoleAssistant && role != models.ChatRoleSystem {
		return nil, errors.New("role must be 'user', 'assistant', or 'system'")
	}

	message := &models.ChatMessage{
		UserID:  userID,
		Role:    role,
		Content: req.Content,
	}

	if err := s.chatRepo.Create(message); err != nil {
		return nil, err
	}

	return message, nil
}

func (s *chatService) GetHistory(userID uuid.UUID, limit int) ([]models.ChatMessage, error) {
	if limit <= 0 {
		limit = 50 // default to last 50 messages
	}
	return s.chatRepo.FindByUserID(userID, limit)
}

func (s *chatService) ClearHistory(userID uuid.UUID) error {
	return s.chatRepo.DeleteByUserID(userID)
}
