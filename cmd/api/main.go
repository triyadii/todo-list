package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"todo-backend/config"
	_ "todo-backend/docs"
	"todo-backend/internal/handler"
	"todo-backend/internal/repository"
	"todo-backend/internal/routes"
	"todo-backend/internal/service"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

// @title           TodoList Backend AI Assistant API
// @version         1.0
// @description     Backend RESTful API untuk aplikasi Todo List cerdas berbasis Go Fiber, GORM, dan PostgreSQL.
// @termsOfService  http://swagger.io/terms/

// @contact.name    Dii Support
// @contact.email   support@todolist.local

// @license.name    Apache 2.0
// @license.url     http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8080
// @BasePath  /

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Format: "Bearer {token}". Masukkan token JWT yang didapat dari endpoint login.

func main() {
	// 1. Load Configurations
	cfg := config.LoadConfig()

	// 2. Initialize Database and Auto-Migrate Models
	db := config.ConnectDB(cfg)

	// 3. Initialize Repositories
	userRepo := repository.NewUserRepository(db)
	taskRepo := repository.NewTaskRepository(db)
	reminderRepo := repository.NewReminderRepository(db)
	chatRepo := repository.NewChatRepository(db)

	// 4. Initialize Services
	authSvc := service.NewAuthService(userRepo)
	userSvc := service.NewUserService(userRepo)
	taskSvc := service.NewTaskService(taskRepo)
	reminderSvc := service.NewReminderService(reminderRepo, taskRepo)
	chatSvc := service.NewChatService(chatRepo)

	// 5. Initialize Handlers
	authHandler := handler.NewAuthHandler(authSvc)
	userHandler := handler.NewUserHandler(userSvc)
	taskHandler := handler.NewTaskHandler(taskSvc)
	reminderHandler := handler.NewReminderHandler(reminderSvc)
	chatHandler := handler.NewChatHandler(chatSvc)

	// 6. Initialize Fiber App
	app := fiber.New(fiber.Config{
		AppName:      "TodoList Backend AI Assistant API v1.0",
		ServerHeader: "Fiber",
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).JSON(fiber.Map{
				"success": false,
				"message": err.Error(),
			})
		},
	})

	// Middlewares
	app.Use(recover.New())
	app.Use(logger.New(logger.Config{
		Format:     "[${time}] ${status} - ${latency} | ${method} ${path}\n",
		TimeFormat: "2006-01-02 15:04:05",
		TimeZone:   cfg.DBTimezone,
	}))
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
		AllowMethods: "GET, POST, PUT, DELETE, OPTIONS",
	}))

	// 7. Register Routes
	routes.SetupRoutes(app, &routes.RouterConfig{
		AuthHandler:     authHandler,
		UserHandler:     userHandler,
		TaskHandler:     taskHandler,
		ReminderHandler: reminderHandler,
		ChatHandler:     chatHandler,
	})

	// 8. Graceful Shutdown listener
	shutdownChan := make(chan os.Signal, 1)
	signal.Notify(shutdownChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-shutdownChan
		log.Println("[INFO] Shutting down server gracefully...")
		_ = app.ShutdownWithTimeout(5 * time.Second)
	}()

	// 9. Start Server
	addr := fmt.Sprintf(":%s", cfg.AppPort)
	log.Printf("[INFO] Server starting on port %s (env: %s)...\n", cfg.AppPort, cfg.AppEnv)
	log.Printf("[INFO] Swagger documentation available at http://localhost:%s/swagger/\n", cfg.AppPort)
	if err := app.Listen(addr); err != nil {
		log.Printf("[INFO] Server stopped: %v\n", err)
	}
}
