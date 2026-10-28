package service

import (
	"context"
	"errors"
	"fmt"
	"log"

	"todo-backend/internal/repository"
	"todo-backend/pkg/fcm"

	"github.com/google/uuid"
)

type SendToTokenRequest struct {
	Token string            `json:"token"`
	Title string            `json:"title"`
	Body  string            `json:"body"`
	Image string            `json:"image,omitempty"`
	Data  map[string]string `json:"data,omitempty"`
}

type SendToUserRequest struct {
	UserID *uuid.UUID        `json:"user_id,omitempty"` // If nil, will use authenticated userID
	Title  string            `json:"title"`
	Body   string            `json:"body"`
	Image  string            `json:"image,omitempty"`
	Data   map[string]string `json:"data,omitempty"`
}

type SendBroadcastRequest struct {
	Title string            `json:"title"`
	Body  string            `json:"body"`
	Image string            `json:"image,omitempty"`
	Data  map[string]string `json:"data,omitempty"`
}

type UpdateFCMTokenRequest struct {
	FCMToken string `json:"fcm_token"`
}

type NotificationService interface {
	SendToToken(ctx context.Context, req *SendToTokenRequest) (*fcm.SendResult, error)
	SendToUser(ctx context.Context, userID uuid.UUID, req *SendToUserRequest) (*fcm.SendResult, error)
	SendBroadcast(ctx context.Context, req *SendBroadcastRequest) ([]fcm.SendResult, error)
	UpdateUserFCMToken(userID uuid.UUID, token string) error
	GetFCMClient() *fcm.Client
}

type notificationService struct {
	userRepo  repository.UserRepository
	fcmClient *fcm.Client
}

func NewNotificationService(userRepo repository.UserRepository, fcmClient *fcm.Client) NotificationService {
	return &notificationService{
		userRepo:  userRepo,
		fcmClient: fcmClient,
	}
}

func (s *notificationService) GetFCMClient() *fcm.Client {
	return s.fcmClient
}

func (s *notificationService) SendToToken(ctx context.Context, req *SendToTokenRequest) (*fcm.SendResult, error) {
	if s.fcmClient == nil {
		return nil, errors.New("fcm client is not configured")
	}
	if req.Token == "" {
		return nil, errors.New("fcm token is required")
	}
	if req.Title == "" {
		return nil, errors.New("notification title is required")
	}
	if req.Body == "" {
		return nil, errors.New("notification body is required")
	}

	payload := &fcm.NotificationPayload{
		Title: req.Title,
		Body:  req.Body,
		Image: req.Image,
		Data:  req.Data,
	}

	return s.fcmClient.Send(ctx, req.Token, payload)
}

func (s *notificationService) SendToUser(ctx context.Context, authUserID uuid.UUID, req *SendToUserRequest) (*fcm.SendResult, error) {
	if s.fcmClient == nil {
		return nil, errors.New("fcm client is not configured")
	}

	targetUserID := authUserID
	if req.UserID != nil && *req.UserID != uuid.Nil {
		targetUserID = *req.UserID
	}

	user, err := s.userRepo.FindByID(targetUserID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	if user.FCMToken == nil || *user.FCMToken == "" {
		return nil, fmt.Errorf("user '%s' (%s) does not have a registered FCM token", user.Name, user.Email)
	}

	if req.Title == "" {
		return nil, errors.New("notification title is required")
	}
	if req.Body == "" {
		return nil, errors.New("notification body is required")
	}

	payload := &fcm.NotificationPayload{
		Title: req.Title,
		Body:  req.Body,
		Image: req.Image,
		Data:  req.Data,
	}

	return s.fcmClient.Send(ctx, *user.FCMToken, payload)
}

func (s *notificationService) SendBroadcast(ctx context.Context, req *SendBroadcastRequest) ([]fcm.SendResult, error) {
	if s.fcmClient == nil {
		return nil, errors.New("fcm client is not configured")
	}
	if req.Title == "" {
		return nil, errors.New("notification title is required")
	}
	if req.Body == "" {
		return nil, errors.New("notification body is required")
	}

	users, err := s.userRepo.FindAllWithFCMToken()
	if err != nil {
		return nil, fmt.Errorf("failed to query users with FCM tokens: %w", err)
	}

	if len(users) == 0 {
		return []fcm.SendResult{}, nil
	}

	tokens := make([]string, 0, len(users))
	for _, u := range users {
		if u.FCMToken != nil && *u.FCMToken != "" {
			tokens = append(tokens, *u.FCMToken)
		}
	}

	payload := &fcm.NotificationPayload{
		Title: req.Title,
		Body:  req.Body,
		Image: req.Image,
		Data:  req.Data,
	}

	log.Printf("[INFO] [FCM] Broadcasting push notification to %d registered devices...", len(tokens))
	results := s.fcmClient.SendMulticast(ctx, tokens, payload)
	return results, nil
}

func (s *notificationService) UpdateUserFCMToken(userID uuid.UUID, token string) error {
	if token == "" {
		return errors.New("fcm_token is required")
	}
	return s.userRepo.UpdateFCMToken(userID, token)
}
