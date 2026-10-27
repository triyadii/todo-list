package handler

import (
	"todo-backend/internal/service"
	"todo-backend/pkg/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type ReminderHandler struct {
	reminderService service.ReminderService
}

func NewReminderHandler(reminderService service.ReminderService) *ReminderHandler {
	return &ReminderHandler{reminderService: reminderService}
}

// CreateReminder godoc
// @Summary      Create a reminder for a task
// @Description  Schedule a reminder execution time with optional AI personalized message
// @Tags         Reminders
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body service.CreateReminderRequest true "Reminder Data"
// @Success      201  {object}  utils.APIResponse{data=models.Reminder} "Reminder scheduled successfully"
// @Failure      400  {object}  utils.APIResponse "Bad Request"
// @Failure      401  {object}  utils.APIResponse "Unauthorized"
// @Router       /api/v1/reminders [post]
func (h *ReminderHandler) CreateReminder(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, "Unauthorized", nil)
	}

	var req service.CreateReminderRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid request body", err.Error())
	}

	reminder, err := h.reminderService.CreateReminder(userID, &req)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error(), nil)
	}

	return utils.SuccessResponse(c, fiber.StatusCreated, "Reminder scheduled successfully", reminder)
}

// GetRemindersByTask godoc
// @Summary      Get reminders of a task
// @Description  Retrieve all reminders created for a specific task
// @Tags         Reminders
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        taskId   path     string  true  "Task UUID"
// @Success      200      {object} utils.APIResponse{data=[]models.Reminder} "Reminders retrieved successfully"
// @Failure      400      {object} utils.APIResponse "Invalid task ID"
// @Failure      401      {object} utils.APIResponse "Unauthorized"
// @Router       /api/v1/tasks/{taskId}/reminders [get]
func (h *ReminderHandler) GetRemindersByTask(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, "Unauthorized", nil)
	}

	taskIDParam := c.Params("taskId")
	taskID, err := uuid.Parse(taskIDParam)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid task ID format", nil)
	}

	reminders, err := h.reminderService.GetRemindersByTask(userID, taskID)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error(), nil)
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "Reminders retrieved successfully", reminders)
}

// DeleteReminder godoc
// @Summary      Delete a reminder
// @Description  Remove a reminder by its UUID
// @Tags         Reminders
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path     string  true  "Reminder UUID"
// @Success      200  {object} utils.APIResponse "Reminder deleted successfully"
// @Failure      400  {object} utils.APIResponse "Invalid reminder ID"
// @Failure      401  {object} utils.APIResponse "Unauthorized"
// @Router       /api/v1/reminders/{id} [delete]
func (h *ReminderHandler) DeleteReminder(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, "Unauthorized", nil)
	}

	reminderIDParam := c.Params("id")
	reminderID, err := uuid.Parse(reminderIDParam)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid reminder ID format", nil)
	}

	if err := h.reminderService.DeleteReminder(userID, reminderID); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error(), nil)
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "Reminder deleted successfully", nil)
}
