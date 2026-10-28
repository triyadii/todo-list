package repository

import (
	"todo-backend/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserRepository interface {
	Create(user *models.User) error
	FindByEmail(email string) (*models.User, error)
	FindByID(id uuid.UUID) (*models.User, error)
	FindAll() ([]models.User, error)
	Update(user *models.User) error
	UpdateFCMToken(id uuid.UUID, fcmToken string) error
	UpdateTimezone(id uuid.UUID, timezone string) error
	FindAllWithFCMToken() ([]models.User, error)
	Delete(id uuid.UUID) error
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(user *models.User) error {
	return r.db.Create(user).Error
}

func (r *userRepository) FindByEmail(email string) (*models.User, error) {
	var user models.User
	if err := r.db.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) FindByID(id uuid.UUID) (*models.User, error) {
	var user models.User
	if err := r.db.First(&user, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) FindAll() ([]models.User, error) {
	var users []models.User
	if err := r.db.Order("created_at DESC").Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

func (r *userRepository) Update(user *models.User) error {
	return r.db.Save(user).Error
}

func (r *userRepository) UpdateFCMToken(id uuid.UUID, fcmToken string) error {
	return r.db.Model(&models.User{}).Where("id = ?", id).Update("fcm_token", fcmToken).Error
}

func (r *userRepository) UpdateTimezone(id uuid.UUID, timezone string) error {
	return r.db.Model(&models.User{}).Where("id = ?", id).Update("timezone", timezone).Error
}

func (r *userRepository) FindAllWithFCMToken() ([]models.User, error) {
	var users []models.User
	if err := r.db.Where("fcm_token IS NOT NULL AND fcm_token != ''").Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

func (r *userRepository) Delete(id uuid.UUID) error {
	return r.db.Where("id = ?", id).Delete(&models.User{}).Error
}
