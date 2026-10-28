package handler

import (
	"todo-backend/internal/service"
	"todo-backend/pkg/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type AuthHandler struct {
	authService service.AuthService
}

func NewAuthHandler(authService service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// Register godoc
// @Summary      Register a new user
// @Description  Register new user with name, email, password, timezone, and optional FCM token
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request body service.RegisterRequest true "User Registration Data"
// @Success      201  {object}  utils.APIResponse{data=service.AuthResponse} "User registered successfully"
// @Failure      400  {object}  utils.APIResponse "Bad Request"
// @Router       /api/v1/auth/register [post]
func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var req service.RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid request body", err.Error())
	}

	res, err := h.authService.Register(&req)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error(), nil)
	}

	return utils.SuccessResponse(c, fiber.StatusCreated, "User registered successfully", res)
}

// Login godoc
// @Summary      Login user
// @Description  Authenticate user with email, password, and FCM device token. Verifies device token to enforce single-device authorization.
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request body service.LoginRequest true "Login Credentials with FCM Token"
// @Success      200  {object}  utils.APIResponse{data=service.AuthResponse} "Login successful"
// @Failure      400  {object}  utils.APIResponse "Invalid request"
// @Failure      401  {object}  utils.APIResponse "Invalid email, password, or device mismatch"
// @Router       /api/v1/auth/login [post]
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req service.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid request body", err.Error())
	}

	res, err := h.authService.Login(&req)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, err.Error(), nil)
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "Login successful", res)
}

// GetProfile godoc
// @Summary      Get current user profile
// @Description  Retrieve logged-in user profile data
// @Tags         User
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  utils.APIResponse{data=models.User} "User profile retrieved successfully"
// @Failure      401  {object}  utils.APIResponse "Unauthorized"
// @Failure      404  {object}  utils.APIResponse "User not found"
// @Router       /api/v1/user/profile [get]
func (h *AuthHandler) GetProfile(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, "Unauthorized", nil)
	}

	user, err := h.authService.GetProfile(userID)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "User not found", err.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "User profile retrieved successfully", user)
}

// UpdateProfile godoc
// @Summary      Update user profile
// @Description  Update user profile details such as name, timezone, or FCM token
// @Tags         User
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body service.UpdateProfileRequest true "Update Profile Request"
// @Success      200  {object}  utils.APIResponse{data=models.User} "Profile updated successfully"
// @Failure      400  {object}  utils.APIResponse "Bad Request"
// @Failure      401  {object}  utils.APIResponse "Unauthorized"
// @Failure      500  {object}  utils.APIResponse "Internal Server Error"
// @Router       /api/v1/user/profile [put]
func (h *AuthHandler) UpdateProfile(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, "Unauthorized", nil)
	}

	var req service.UpdateProfileRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid request body", err.Error())
	}

	user, err := h.authService.UpdateProfile(userID, &req)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to update profile", err.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "Profile updated successfully", user)
}
