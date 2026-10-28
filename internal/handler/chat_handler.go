package handler

import (
	"strconv"

	"todo-backend/internal/service"
	"todo-backend/pkg/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type ChatHandler struct {
	chatService service.ChatService
}

func NewChatHandler(chatService service.ChatService) *ChatHandler {
	return &ChatHandler{chatService: chatService}
}

// SendMessage godoc
// @Summary      Send message to AI assistant
// @Description  Send a chat message to Groq AI Assistant. The AI automatically processes task or reminder actions and replies back.
// @Tags         Chat & AI
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body service.SendMessageRequest true "Chat Message Data"
// @Success      201  {object}  utils.APIResponse{data=service.ChatResponse} "Message processed with AI successfully"
// @Failure      400  {object}  utils.APIResponse "Bad Request"
// @Failure      401  {object}  utils.APIResponse "Unauthorized"
// @Router       /api/v1/chat/messages [post]
func (h *ChatHandler) SendMessage(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, "Unauthorized", nil)
	}

	var req service.SendMessageRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid request body", err.Error())
	}

	resp, err := h.chatService.SendMessage(userID, &req)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error(), nil)
	}

	return utils.SuccessResponse(c, fiber.StatusCreated, "Message processed successfully", resp)
}

// GetHistory godoc
// @Summary      Get chat history
// @Description  Retrieve chat history conversation for current user (used to feed AI memory context)
// @Tags         Chat & AI
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        limit  query    int  false  "Number of recent messages to return (default: 50)"
// @Success      200    {object} utils.APIResponse{data=[]models.ChatMessage} "Chat history retrieved successfully"
// @Failure      401    {object} utils.APIResponse "Unauthorized"
// @Failure      500    {object} utils.APIResponse "Internal Server Error"
// @Router       /api/v1/chat/messages [get]
func (h *ChatHandler) GetHistory(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, "Unauthorized", nil)
	}

	limitStr := c.Query("limit", "50")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 50
	}

	messages, err := h.chatService.GetHistory(userID, limit)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to fetch chat history", err.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "Chat history retrieved successfully", messages)
}

// ClearHistory godoc
// @Summary      Clear chat history
// @Description  Delete all chat history messages for current user
// @Tags         Chat & AI
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object} utils.APIResponse "Chat history cleared successfully"
// @Failure      401  {object} utils.APIResponse "Unauthorized"
// @Failure      500  {object} utils.APIResponse "Internal Server Error"
// @Router       /api/v1/chat/messages [delete]
func (h *ChatHandler) ClearHistory(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, "Unauthorized", nil)
	}

	if err := h.chatService.ClearHistory(userID); err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to clear chat history", err.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "Chat history cleared successfully", nil)
}
