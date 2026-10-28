package repository

import (
	"todo-backend/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TaskRepository interface {
	Create(task *models.Task) error
	FindByID(id uuid.UUID, userID uuid.UUID) (*models.Task, error)
	FindByTitleKeyword(userID uuid.UUID, keyword string) (*models.Task, error)
	FindAllByUserID(userID uuid.UUID, status string, priority string, dueDate string) ([]models.Task, error)
	Update(task *models.Task) error
	Delete(id uuid.UUID, userID uuid.UUID) error
}

type taskRepository struct {
	db *gorm.DB
}

func NewTaskRepository(db *gorm.DB) TaskRepository {
	return &taskRepository{db: db}
}

func (r *taskRepository) Create(task *models.Task) error {
	return r.db.Create(task).Error
}

func (r *taskRepository) FindByID(id uuid.UUID, userID uuid.UUID) (*models.Task, error) {
	var task models.Task
	if err := r.db.Preload("Reminders").Where("id = ? AND user_id = ?", id, userID).First(&task).Error; err != nil {
		return nil, err
	}
	return &task, nil
}

func (r *taskRepository) FindByTitleKeyword(userID uuid.UUID, keyword string) (*models.Task, error) {
	var task models.Task
	pattern := "%" + keyword + "%"
	if err := r.db.Preload("Reminders").Where("user_id = ? AND title ILIKE ?", userID, pattern).Order("created_at DESC").First(&task).Error; err != nil {
		return nil, err
	}
	return &task, nil
}

func (r *taskRepository) FindAllByUserID(userID uuid.UUID, status string, priority string, dueDate string) ([]models.Task, error) {
	var tasks []models.Task
	query := r.db.Preload("Reminders").Where("user_id = ?", userID)

	if status != "" {
		query = query.Where("status = ?", status)
	}
	if priority != "" {
		query = query.Where("priority = ?", priority)
	}
	if dueDate != "" {
		query = query.Where("due_date = ?", dueDate)
	}

	if err := query.Order("due_date ASC, start_time ASC NULLS LAST, created_at DESC").Find(&tasks).Error; err != nil {
		return nil, err
	}
	return tasks, nil
}

func (r *taskRepository) Update(task *models.Task) error {
	return r.db.Save(task).Error
}

func (r *taskRepository) Delete(id uuid.UUID, userID uuid.UUID) error {
	return r.db.Where("id = ? AND user_id = ?", id, userID).Delete(&models.Task{}).Error
}
