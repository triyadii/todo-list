package service

import (
	"errors"

	"todo-backend/internal/models"
	"todo-backend/internal/repository"
	"todo-backend/pkg/utils"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AuthService interface {
	Register(req *RegisterRequest) (*AuthResponse, error)
	Login(req *LoginRequest) (*AuthResponse, error)
	GetProfile(userID uuid.UUID) (*models.User, error)
	UpdateProfile(userID uuid.UUID, req *UpdateProfileRequest) (*models.User, error)
}

type authService struct {
	userRepo repository.UserRepository
}

func NewAuthService(userRepo repository.UserRepository) AuthService {
	return &authService{userRepo: userRepo}
}

type RegisterRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Timezone string `json:"timezone"`
	FCMToken string `json:"fcm_token,omitempty"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UpdateProfileRequest struct {
	Name     *string `json:"name,omitempty"`
	Timezone *string `json:"timezone,omitempty"`
	FCMToken *string `json:"fcm_token,omitempty"`
}

type AuthResponse struct {
	Token string       `json:"token"`
	User  *models.User `json:"user"`
}

func (s *authService) Register(req *RegisterRequest) (*AuthResponse, error) {
	if req.Email == "" || req.Password == "" || req.Name == "" {
		return nil, errors.New("name, email and password are required")
	}

	existingUser, err := s.userRepo.FindByEmail(req.Email)
	if err == nil && existingUser != nil {
		return nil, errors.New("email already registered")
	} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, errors.New("failed to hash password")
	}

	tz := req.Timezone
	if tz == "" {
		tz = "Asia/Jakarta"
	}

	var fcm *string
	if req.FCMToken != "" {
		fcm = &req.FCMToken
	}

	user := &models.User{
		Name:         req.Name,
		Email:        req.Email,
		PasswordHash: hashedPassword,
		Timezone:     tz,
		FCMToken:     fcm,
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}

	token, err := utils.GenerateToken(user.ID, user.Email, user.Timezone)
	if err != nil {
		return nil, errors.New("failed to generate token")
	}

	return &AuthResponse{
		Token: token,
		User:  user,
	}, nil
}

func (s *authService) Login(req *LoginRequest) (*AuthResponse, error) {
	if req.Email == "" || req.Password == "" {
		return nil, errors.New("email and password are required")
	}

	user, err := s.userRepo.FindByEmail(req.Email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("invalid email or password")
		}
		return nil, err
	}

	if !utils.CheckPassword(req.Password, user.PasswordHash) {
		return nil, errors.New("invalid email or password")
	}

	token, err := utils.GenerateToken(user.ID, user.Email, user.Timezone)
	if err != nil {
		return nil, errors.New("failed to generate token")
	}

	return &AuthResponse{
		Token: token,
		User:  user,
	}, nil
}

func (s *authService) GetProfile(userID uuid.UUID) (*models.User, error) {
	return s.userRepo.FindByID(userID)
}

func (s *authService) UpdateProfile(userID uuid.UUID, req *UpdateProfileRequest) (*models.User, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, err
	}

	if req.Name != nil && *req.Name != "" {
		user.Name = *req.Name
	}
	if req.Timezone != nil && *req.Timezone != "" {
		user.Timezone = *req.Timezone
	}
	if req.FCMToken != nil {
		user.FCMToken = req.FCMToken
	}

	if err := s.userRepo.Update(user); err != nil {
		return nil, err
	}

	return user, nil
}
