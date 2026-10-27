package handler

import (
	"todo-backend/internal/service"
	"todo-backend/pkg/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type UserHandler struct {
	userService service.UserService
}

func NewUserHandler(userService service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

// CreateUser godoc
// @Summary      Add / Create a new user
// @Description  Create a new user with name, email, password, timezone, and optional FCM token
// @Tags         User Management
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body service.CreateUserRequest true "New User Data"
// @Success      201  {object}  utils.APIResponse{data=models.User} "User created successfully"
// @Failure      400  {object}  utils.APIResponse "Bad Request"
// @Failure      401  {object}  utils.APIResponse "Unauthorized"
// @Router       /api/v1/users [post]
func (h *UserHandler) CreateUser(c *fiber.Ctx) error {
	var req service.CreateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid request body", err.Error())
	}

	user, err := h.userService.CreateUser(&req)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error(), nil)
	}

	return utils.SuccessResponse(c, fiber.StatusCreated, "User created successfully", user)
}

// GetAllUsers godoc
// @Summary      Get all users
// @Description  Retrieve list of all registered users
// @Tags         User Management
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  utils.APIResponse{data=[]models.User} "Users retrieved successfully"
// @Failure      401  {object}  utils.APIResponse "Unauthorized"
// @Failure      500  {object}  utils.APIResponse "Internal Server Error"
// @Router       /api/v1/users [get]
func (h *UserHandler) GetAllUsers(c *fiber.Ctx) error {
	users, err := h.userService.GetAllUsers()
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to fetch users", err.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "Users retrieved successfully", users)
}

// GetUserByID godoc
// @Summary      Get user by ID
// @Description  Retrieve single user details by UUID
// @Tags         User Management
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path     string  true  "User UUID"
// @Success      200  {object}  utils.APIResponse{data=models.User} "User retrieved successfully"
// @Failure      400  {object}  utils.APIResponse "Invalid user ID"
// @Failure      401  {object}  utils.APIResponse "Unauthorized"
// @Failure      404  {object}  utils.APIResponse "User not found"
// @Router       /api/v1/users/{id} [get]
func (h *UserHandler) GetUserByID(c *fiber.Ctx) error {
	idParam := c.Params("id")
	userID, err := uuid.Parse(idParam)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid user ID format", nil)
	}

	user, err := h.userService.GetUserByID(userID)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "User not found", err.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "User retrieved successfully", user)
}

// DeleteUser godoc
// @Summary      Delete a user
// @Description  Remove user by UUID (associated tasks, reminders, and chat messages will be cascade deleted)
// @Tags         User Management
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path     string  true  "User UUID"
// @Success      200  {object}  utils.APIResponse "User deleted successfully"
// @Failure      400  {object}  utils.APIResponse "Invalid user ID"
// @Failure      401  {object}  utils.APIResponse "Unauthorized"
// @Failure      500  {object}  utils.APIResponse "Internal Server Error"
// @Router       /api/v1/users/{id} [delete]
func (h *UserHandler) DeleteUser(c *fiber.Ctx) error {
	idParam := c.Params("id")
	userID, err := uuid.Parse(idParam)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid user ID format", nil)
	}

	if err := h.userService.DeleteUser(userID); err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to delete user", err.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "User deleted successfully", nil)
}
