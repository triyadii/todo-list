package handler

import (
	"todo-backend/internal/service"
	"todo-backend/pkg/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type NotificationHandler struct {
	notificationService service.NotificationService
}

func NewNotificationHandler(notificationService service.NotificationService) *NotificationHandler {
	return &NotificationHandler{
		notificationService: notificationService,
	}
}

// SendToToken godoc
// @Summary      Send push notification to specific FCM token
// @Description  Deliver push notification to a device token using Firebase Cloud Messaging (FCM HTTP v1)
// @Tags         Notifications & FCM
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body service.SendToTokenRequest true "Push Notification Payload"
// @Success      200  {object}  utils.APIResponse{data=fcm.SendResult} "Notification sent successfully"
// @Failure      400  {object}  utils.APIResponse "Bad Request"
// @Failure      401  {object}  utils.APIResponse "Unauthorized"
// @Failure      500  {object}  utils.APIResponse "Internal Server Error"
// @Router       /api/v1/notifications/send [post]
func (h *NotificationHandler) SendToToken(c *fiber.Ctx) error {
	var req service.SendToTokenRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid request body", err.Error())
	}

	result, err := h.notificationService.SendToToken(c.Context(), &req)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error(), result)
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "Push notification sent successfully", result)
}

// SendToUser godoc
// @Summary      Send push notification to user
// @Description  Send push notification to a user's registered FCM token. If user_id is omitted, notification goes to the current authenticated user.
// @Tags         Notifications & FCM
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body service.SendToUserRequest true "Send To User Payload"
// @Success      200  {object}  utils.APIResponse{data=fcm.SendResult} "Notification sent to user successfully"
// @Failure      400  {object}  utils.APIResponse "Bad Request"
// @Failure      401  {object}  utils.APIResponse "Unauthorized"
// @Failure      404  {object}  utils.APIResponse "User not found or no FCM token registered"
// @Router       /api/v1/notifications/send-to-user [post]
func (h *NotificationHandler) SendToUser(c *fiber.Ctx) error {
	authUserID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, "Unauthorized", nil)
	}

	var req service.SendToUserRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid request body", err.Error())
	}

	result, err := h.notificationService.SendToUser(c.Context(), authUserID, &req)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error(), result)
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "Push notification sent to user successfully", result)
}

// SendBroadcast godoc
// @Summary      Broadcast push notification to all users
// @Description  Send push notification to all users with active registered FCM tokens
// @Tags         Notifications & FCM
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body service.SendBroadcastRequest true "Broadcast Notification Payload"
// @Success      200  {object}  utils.APIResponse{data=[]fcm.SendResult} "Broadcast sent successfully"
// @Failure      400  {object}  utils.APIResponse "Bad Request"
// @Failure      401  {object}  utils.APIResponse "Unauthorized"
// @Failure      500  {object}  utils.APIResponse "Internal Server Error"
// @Router       /api/v1/notifications/broadcast [post]
func (h *NotificationHandler) SendBroadcast(c *fiber.Ctx) error {
	var req service.SendBroadcastRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid request body", err.Error())
	}

	results, err := h.notificationService.SendBroadcast(c.Context(), &req)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error(), nil)
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "Broadcast push notification processed successfully", results)
}

// UpdateFCMToken godoc
// @Summary      Update current user's FCM token
// @Description  Store or refresh the Firebase Cloud Messaging device token for push notifications
// @Tags         Notifications & FCM
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body service.UpdateFCMTokenRequest true "FCM Token Data"
// @Success      200  {object}  utils.APIResponse "FCM token updated successfully"
// @Failure      400  {object}  utils.APIResponse "Bad Request"
// @Failure      401  {object}  utils.APIResponse "Unauthorized"
// @Router       /api/v1/notifications/fcm-token [post]
func (h *NotificationHandler) UpdateFCMToken(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, "Unauthorized", nil)
	}

	var req service.UpdateFCMTokenRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid request body", err.Error())
	}

	if err := h.notificationService.UpdateUserFCMToken(userID, req.FCMToken); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error(), nil)
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "FCM token registered successfully", fiber.Map{
		"fcm_token": req.FCMToken,
	})
}
