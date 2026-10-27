package service

import (
	"errors"
	"time"

	"todo-backend/internal/models"
	"todo-backend/internal/repository"

	"github.com/google/uuid"
)

type ReminderService interface {
	CreateReminder(userID uuid.UUID, req *CreateReminderRequest) (*models.Reminder, error)
	GetRemindersByTask(userID, taskID uuid.UUID) ([]models.Reminder, error)
	DeleteReminder(userID, reminderID uuid.UUID) error
	GetPendingReminders(until time.Time) ([]models.Reminder, error)
	MarkReminderSent(reminderID uuid.UUID) error
}

type reminderService struct {
	reminderRepo repository.ReminderRepository
	taskRepo     repository.TaskRepository
}

func NewReminderService(reminderRepo repository.ReminderRepository, taskRepo repository.TaskRepository) ReminderService {
	return &reminderService{
		reminderRepo: reminderRepo,
		taskRepo:     taskRepo,
	}
}

type CreateReminderRequest struct {
	TaskID    uuid.UUID `json:"task_id"`
	RemindAt  time.Time `json:"remind_at"`
	AIMessage *string   `json:"ai_message,omitempty"`
}

func (s *reminderService) CreateReminder(userID uuid.UUID, req *CreateReminderRequest) (*models.Reminder, error) {
	if req.TaskID == uuid.Nil {
		return nil, errors.New("task_id is required")
	}
	if req.RemindAt.IsZero() {
		return nil, errors.New("remind_at is required")
	}

	// Validate task exists and belongs to the user
	task, err := s.taskRepo.FindByID(req.TaskID, userID)
	if err != nil || task == nil {
		return nil, errors.New("task not found or does not belong to user")
	}

	reminder := &models.Reminder{
		TaskID:    req.TaskID,
		RemindAt:  req.RemindAt,
		IsSent:    false,
		AIMessage: req.AIMessage,
	}

	if err := s.reminderRepo.Create(reminder); err != nil {
		return nil, err
	}

	return reminder, nil
}

func (s *reminderService) GetRemindersByTask(userID, taskID uuid.UUID) ([]models.Reminder, error) {
	// Verify task ownership
	task, err := s.taskRepo.FindByID(taskID, userID)
	if err != nil || task == nil {
		return nil, errors.New("task not found or does not belong to user")
	}

	return s.reminderRepo.FindByTaskID(taskID)
}

func (s *reminderService) DeleteReminder(userID, reminderID uuid.UUID) error {
	reminder, err := s.reminderRepo.FindByID(reminderID)
	if err != nil {
		return errors.New("reminder not found")
	}

	// Verify task ownership
	task, err := s.taskRepo.FindByID(reminder.TaskID, userID)
	if err != nil || task == nil {
		return errors.New("unauthorized to delete this reminder")
	}

	return s.reminderRepo.Delete(reminderID)
}

func (s *reminderService) GetPendingReminders(until time.Time) ([]models.Reminder, error) {
	return s.reminderRepo.FindPendingReminders(until)
}

func (s *reminderService) MarkReminderSent(reminderID uuid.UUID) error {
	return s.reminderRepo.MarkAsSent(reminderID, time.Now())
}
