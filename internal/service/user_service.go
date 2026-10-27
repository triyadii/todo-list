package service

import (
	"errors"

	"todo-backend/internal/models"
	"todo-backend/internal/repository"
	"todo-backend/pkg/utils"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CreateUserRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Timezone string `json:"timezone,omitempty"`
	FCMToken string `json:"fcm_token,omitempty"`
}

type UserService interface {
	CreateUser(req *CreateUserRequest) (*models.User, error)
	GetAllUsers() ([]models.User, error)
	GetUserByID(id uuid.UUID) (*models.User, error)
	DeleteUser(id uuid.UUID) error
}

type userService struct {
	userRepo repository.UserRepository
}

func NewUserService(userRepo repository.UserRepository) UserService {
	return &userService{userRepo: userRepo}
}

func (s *userService) CreateUser(req *CreateUserRequest) (*models.User, error) {
	if req.Name == "" || req.Email == "" || req.Password == "" {
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

	return user, nil
}

func (s *userService) GetAllUsers() ([]models.User, error) {
	return s.userRepo.FindAll()
}

func (s *userService) GetUserByID(id uuid.UUID) (*models.User, error) {
	return s.userRepo.FindByID(id)
}

func (s *userService) DeleteUser(id uuid.UUID) error {
	return s.userRepo.Delete(id)
}
