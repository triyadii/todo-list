package service

import (
	"errors"

	"todo-backend/internal/models"
	"todo-backend/internal/repository"

	"github.com/google/uuid"
)

type TaskService interface {
	CreateTask(userID uuid.UUID, req *CreateTaskRequest) (*models.Task, error)
	GetTasks(userID uuid.UUID, status, priority, dueDate string) ([]models.Task, error)
	GetTaskByID(userID, taskID uuid.UUID) (*models.Task, error)
	FindTaskByKeywordOrID(userID uuid.UUID, idStr, keyword string) (*models.Task, error)
	UpdateTask(userID, taskID uuid.UUID, req *UpdateTaskRequest) (*models.Task, error)
	DeleteTask(userID, taskID uuid.UUID) error
}

type taskService struct {
	taskRepo repository.TaskRepository
}

func NewTaskService(taskRepo repository.TaskRepository) TaskService {
	return &taskService{taskRepo: taskRepo}
}

type CreateTaskRequest struct {
	Title       string              `json:"title"`
	Description *string             `json:"description,omitempty"`
	DueDate     string              `json:"due_date"`    // Format: YYYY-MM-DD
	StartTime   *string             `json:"start_time,omitempty"` // Format: HH:MM:SS
	EndTime     *string             `json:"end_time,omitempty"`   // Format: HH:MM:SS
	Status      models.TaskStatus   `json:"status,omitempty"`
	Priority    models.TaskPriority `json:"priority,omitempty"`
}

type UpdateTaskRequest struct {
	Title       *string              `json:"title,omitempty"`
	Description *string              `json:"description,omitempty"`
	DueDate     *string              `json:"due_date,omitempty"`
	StartTime   *string              `json:"start_time,omitempty"`
	EndTime     *string              `json:"end_time,omitempty"`
	Status      *models.TaskStatus   `json:"status,omitempty"`
	Priority    *models.TaskPriority `json:"priority,omitempty"`
}

func (s *taskService) CreateTask(userID uuid.UUID, req *CreateTaskRequest) (*models.Task, error) {
	if req.Title == "" {
		return nil, errors.New("title is required")
	}
	if req.DueDate == "" {
		return nil, errors.New("due_date is required (format: YYYY-MM-DD)")
	}

	task := &models.Task{
		UserID:      userID,
		Title:       req.Title,
		Description: req.Description,
		DueDate:     req.DueDate,
		StartTime:   req.StartTime,
		EndTime:     req.EndTime,
		Status:      req.Status,
		Priority:    req.Priority,
	}

	if err := s.taskRepo.Create(task); err != nil {
		return nil, err
	}

	return task, nil
}

func (s *taskService) GetTasks(userID uuid.UUID, status, priority, dueDate string) ([]models.Task, error) {
	return s.taskRepo.FindAllByUserID(userID, status, priority, dueDate)
}

func (s *taskService) GetTaskByID(userID, taskID uuid.UUID) (*models.Task, error) {
	return s.taskRepo.FindByID(taskID, userID)
}

func (s *taskService) FindTaskByKeywordOrID(userID uuid.UUID, idStr, keyword string) (*models.Task, error) {
	if idStr != "" {
		if id, err := uuid.Parse(idStr); err == nil {
			return s.taskRepo.FindByID(id, userID)
		}
	}
	if keyword != "" {
		return s.taskRepo.FindByTitleKeyword(userID, keyword)
	}
	return nil, errors.New("neither id nor keyword provided")
}

func (s *taskService) UpdateTask(userID, taskID uuid.UUID, req *UpdateTaskRequest) (*models.Task, error) {
	task, err := s.taskRepo.FindByID(taskID, userID)
	if err != nil {
		return nil, errors.New("task not found")
	}

	if req.Title != nil {
		task.Title = *req.Title
	}
	if req.Description != nil {
		task.Description = req.Description
	}
	if req.DueDate != nil {
		task.DueDate = *req.DueDate
	}
	if req.StartTime != nil {
		task.StartTime = req.StartTime
	}
	if req.EndTime != nil {
		task.EndTime = req.EndTime
	}
	if req.Status != nil {
		task.Status = *req.Status
	}
	if req.Priority != nil {
		task.Priority = *req.Priority
	}

	if err := s.taskRepo.Update(task); err != nil {
		return nil, err
	}

	return task, nil
}

func (s *taskService) DeleteTask(userID, taskID uuid.UUID) error {
	return s.taskRepo.Delete(taskID, userID)
}
