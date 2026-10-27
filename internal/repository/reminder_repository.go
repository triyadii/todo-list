package repository

import (
	"time"

	"todo-backend/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ReminderRepository interface {
	Create(reminder *models.Reminder) error
	FindByID(id uuid.UUID) (*models.Reminder, error)
	FindByTaskID(taskID uuid.UUID) ([]models.Reminder, error)
	FindPendingReminders(until time.Time) ([]models.Reminder, error)
	MarkAsSent(id uuid.UUID, sentAt time.Time) error
	Update(reminder *models.Reminder) error
	Delete(id uuid.UUID) error
}

type reminderRepository struct {
	db *gorm.DB
}

func NewReminderRepository(db *gorm.DB) ReminderRepository {
	return &reminderRepository{db: db}
}

func (r *reminderRepository) Create(reminder *models.Reminder) error {
	return r.db.Create(reminder).Error
}

func (r *reminderRepository) FindByID(id uuid.UUID) (*models.Reminder, error) {
	var reminder models.Reminder
	if err := r.db.Preload("Task").First(&reminder, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &reminder, nil
}

func (r *reminderRepository) FindByTaskID(taskID uuid.UUID) ([]models.Reminder, error) {
	var reminders []models.Reminder
	if err := r.db.Where("task_id = ?", taskID).Order("remind_at ASC").Find(&reminders).Error; err != nil {
		return nil, err
	}
	return reminders, nil
}

func (r *reminderRepository) FindPendingReminders(until time.Time) ([]models.Reminder, error) {
	var reminders []models.Reminder
	if err := r.db.Preload("Task.User").
		Where("is_sent = ? AND remind_at <= ?", false, until).
		Order("remind_at ASC").
		Find(&reminders).Error; err != nil {
		return nil, err
	}
	return reminders, nil
}

func (r *reminderRepository) MarkAsSent(id uuid.UUID, sentAt time.Time) error {
	return r.db.Model(&models.Reminder{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"is_sent": true,
			"sent_at": sentAt,
		}).Error
}

func (r *reminderRepository) Update(reminder *models.Reminder) error {
	return r.db.Save(reminder).Error
}

func (r *reminderRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.Reminder{}, "id = ?", id).Error
}
