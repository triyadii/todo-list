package routes

import (
	"todo-backend/internal/handler"
	"todo-backend/internal/middleware"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/swagger"
)

type RouterConfig struct {
	AuthHandler     *handler.AuthHandler
	UserHandler     *handler.UserHandler
	TaskHandler     *handler.TaskHandler
	ReminderHandler *handler.ReminderHandler
	ChatHandler     *handler.ChatHandler
}

func SetupRoutes(app *fiber.App, cfg *RouterConfig) {
	// Swagger Documentation UI
	app.Get("/swagger/*", swagger.HandlerDefault)

	// Root endpoint
	app.Get("/", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"status":  "ok",
			"message": "TodoList Backend AI Assistant API is running",
			"version": "1.0",
			"endpoints": fiber.Map{
				"health":  "health",
				"swagger": "swagger/index.html",
				"api":     "api/v1",
			},
		})
	})

	// Health check
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"status":  "ok",
			"message": "TodoList API is running smoothly",
		})
	})

	api := app.Group("/api/v1")

	// Authentication routes
	auth := api.Group("/auth")
	auth.Post("/register", cfg.AuthHandler.Register)
	auth.Post("/login", cfg.AuthHandler.Login)

	// Protected routes
	protected := api.Group("", middleware.Protected())

	// User Profile (Current logged in user)
	user := protected.Group("/user")
	user.Get("/profile", cfg.AuthHandler.GetProfile)
	user.Put("/profile", cfg.AuthHandler.UpdateProfile)

	// User Management (Add, List, Get, Delete)
	users := protected.Group("/users")
	users.Post("/", cfg.UserHandler.CreateUser)
	users.Get("/", cfg.UserHandler.GetAllUsers)
	users.Get("/:id", cfg.UserHandler.GetUserByID)
	users.Delete("/:id", cfg.UserHandler.DeleteUser)

	// Tasks
	tasks := protected.Group("/tasks")
	tasks.Post("/", cfg.TaskHandler.CreateTask)
	tasks.Get("/", cfg.TaskHandler.GetTasks)
	tasks.Get("/:id", cfg.TaskHandler.GetTaskByID)
	tasks.Put("/:id", cfg.TaskHandler.UpdateTask)
	tasks.Delete("/:id", cfg.TaskHandler.DeleteTask)
	tasks.Get("/:taskId/reminders", cfg.ReminderHandler.GetRemindersByTask)

	// Reminders
	reminders := protected.Group("/reminders")
	reminders.Post("/", cfg.ReminderHandler.CreateReminder)
	reminders.Delete("/:id", cfg.ReminderHandler.DeleteReminder)

	// Chat Messages (AI Context)
	chat := protected.Group("/chat")
	chat.Post("/messages", cfg.ChatHandler.SendMessage)
	chat.Get("/messages", cfg.ChatHandler.GetHistory)
	chat.Delete("/messages", cfg.ChatHandler.ClearHistory)
}
