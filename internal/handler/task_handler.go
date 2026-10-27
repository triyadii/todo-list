package handler

import (
	"todo-backend/internal/service"
	"todo-backend/pkg/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type TaskHandler struct {
	taskService service.TaskService
}

func NewTaskHandler(taskService service.TaskService) *TaskHandler {
	return &TaskHandler{taskService: taskService}
}

// CreateTask godoc
// @Summary      Create a new task
// @Description  Create a new todo task/activity with due date, time, status, and priority
// @Tags         Tasks
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body service.CreateTaskRequest true "Task Creation Data"
// @Success      201  {object}  utils.APIResponse{data=models.Task} "Task created successfully"
// @Failure      400  {object}  utils.APIResponse "Bad Request"
// @Failure      401  {object}  utils.APIResponse "Unauthorized"
// @Router       /api/v1/tasks [post]
func (h *TaskHandler) CreateTask(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, "Unauthorized", nil)
	}

	var req service.CreateTaskRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid request body", err.Error())
	}

	task, err := h.taskService.CreateTask(userID, &req)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error(), nil)
	}

	return utils.SuccessResponse(c, fiber.StatusCreated, "Task created successfully", task)
}

// GetTasks godoc
// @Summary      Get all tasks
// @Description  Get tasks of logged-in user with optional filters by status, priority, or due_date
// @Tags         Tasks
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        status    query    string  false  "Filter by status (pending, in_progress, completed, cancelled)"
// @Param        priority  query    string  false  "Filter by priority (low, medium, high)"
// @Param        due_date  query    string  false  "Filter by due date (YYYY-MM-DD)"
// @Success      200       {object} utils.APIResponse{data=[]models.Task} "Tasks retrieved successfully"
// @Failure      401       {object} utils.APIResponse "Unauthorized"
// @Failure      500       {object} utils.APIResponse "Internal Server Error"
// @Router       /api/v1/tasks [get]
func (h *TaskHandler) GetTasks(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, "Unauthorized", nil)
	}

	status := c.Query("status")
	priority := c.Query("priority")
	dueDate := c.Query("due_date")

	tasks, err := h.taskService.GetTasks(userID, status, priority, dueDate)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to fetch tasks", err.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "Tasks retrieved successfully", tasks)
}

// GetTaskByID godoc
// @Summary      Get task by ID
// @Description  Retrieve details of a specific task including its reminders
// @Tags         Tasks
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path     string  true  "Task UUID"
// @Success      200  {object} utils.APIResponse{data=models.Task} "Task retrieved successfully"
// @Failure      400  {object} utils.APIResponse "Invalid task ID"
// @Failure      401  {object} utils.APIResponse "Unauthorized"
// @Failure      404  {object} utils.APIResponse "Task not found"
// @Router       /api/v1/tasks/{id} [get]
func (h *TaskHandler) GetTaskByID(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, "Unauthorized", nil)
	}

	idParam := c.Params("id")
	taskID, err := uuid.Parse(idParam)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid task ID format", nil)
	}

	task, err := h.taskService.GetTaskByID(userID, taskID)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "Task not found", err.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "Task retrieved successfully", task)
}

// UpdateTask godoc
// @Summary      Update a task
// @Description  Update fields of an existing task
// @Tags         Tasks
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id       path     string                   true  "Task UUID"
// @Param        request  body     service.UpdateTaskRequest true  "Task Update Data"
// @Success      200      {object} utils.APIResponse{data=models.Task} "Task updated successfully"
// @Failure      400      {object} utils.APIResponse "Invalid request"
// @Failure      401      {object} utils.APIResponse "Unauthorized"
// @Router       /api/v1/tasks/{id} [put]
func (h *TaskHandler) UpdateTask(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, "Unauthorized", nil)
	}

	idParam := c.Params("id")
	taskID, err := uuid.Parse(idParam)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid task ID format", nil)
	}

	var req service.UpdateTaskRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid request body", err.Error())
	}

	task, err := h.taskService.UpdateTask(userID, taskID, &req)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error(), nil)
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "Task updated successfully", task)
}

// DeleteTask godoc
// @Summary      Delete a task
// @Description  Delete a task by ID (reminders associated will also be deleted via cascade)
// @Tags         Tasks
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path     string  true  "Task UUID"
// @Success      200  {object} utils.APIResponse "Task deleted successfully"
// @Failure      400  {object} utils.APIResponse "Invalid task ID"
// @Failure      401  {object} utils.APIResponse "Unauthorized"
// @Failure      500  {object} utils.APIResponse "Internal Server Error"
// @Router       /api/v1/tasks/{id} [delete]
func (h *TaskHandler) DeleteTask(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, "Unauthorized", nil)
	}

	idParam := c.Params("id")
	taskID, err := uuid.Parse(idParam)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid task ID format", nil)
	}

	if err := h.taskService.DeleteTask(userID, taskID); err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to delete task", err.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "Task deleted successfully", nil)
}
