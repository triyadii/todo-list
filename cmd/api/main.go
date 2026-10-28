package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"todo-backend/config"
	"todo-backend/docs"
	"todo-backend/internal/handler"
	"todo-backend/internal/repository"
	"todo-backend/internal/routes"
	"todo-backend/internal/service"
	"todo-backend/internal/worker"
	"todo-backend/pkg/fcm"
	"todo-backend/pkg/groq"

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

// @host      mdt-solution.site
// @BasePath  /todo-list
// @schemes   https http

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Format: "Bearer {token}". Masukkan token JWT yang didapat dari endpoint login.

func main() {
	// 1. Load Configurations
	cfg := config.LoadConfig()

	// Configure Swagger metadata dynamically from config/.env
	docs.SwaggerInfo.Host = cfg.SwaggerHost
	docs.SwaggerInfo.BasePath = cfg.SwaggerBasePath
	if len(cfg.SwaggerSchemes) > 0 {
		docs.SwaggerInfo.Schemes = cfg.SwaggerSchemes
	}

	// 2. Initialize Database and Auto-Migrate Models
	db := config.ConnectDB(cfg)

	// 3. Initialize Repositories
	userRepo := repository.NewUserRepository(db)
	taskRepo := repository.NewTaskRepository(db)
	reminderRepo := repository.NewReminderRepository(db)
	chatRepo := repository.NewChatRepository(db)

	// Initialize FCM Push Notification Client
	var fcmClient *fcm.Client
	if cfg.FirebaseCredPath != "" {
		client, err := fcm.NewClient(cfg.FirebaseCredPath)
		if err != nil {
			log.Printf("[WARN] Failed to initialize Firebase Cloud Messaging (FCM): %v", err)
		} else {
			fcmClient = client
			log.Printf("[INFO] Firebase Cloud Messaging (FCM) initialized successfully for project: '%s'", fcmClient.GetProjectID())
		}
	}

	// 4. Initialize Services
	authSvc := service.NewAuthService(userRepo)
	userSvc := service.NewUserService(userRepo)
	taskSvc := service.NewTaskService(taskRepo)
	reminderSvc := service.NewReminderService(reminderRepo, taskRepo)
	notificationSvc := service.NewNotificationService(userRepo, fcmClient)
	groqClient := groq.NewClient(cfg.GroqAPIKey, cfg.GroqModel)
	chatSvc := service.NewChatService(chatRepo, taskSvc, reminderSvc, groqClient, cfg.DBTimezone)

	// 5. Initialize Handlers
	authHandler := handler.NewAuthHandler(authSvc)
	userHandler := handler.NewUserHandler(userSvc)
	taskHandler := handler.NewTaskHandler(taskSvc)
	reminderHandler := handler.NewReminderHandler(reminderSvc)
	notificationHandler := handler.NewNotificationHandler(notificationSvc)
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
		AuthHandler:         authHandler,
		UserHandler:         userHandler,
		TaskHandler:         taskHandler,
		ReminderHandler:     reminderHandler,
		ChatHandler:         chatHandler,
		NotificationHandler: notificationHandler,
	})

	// 8. Start Background Reminder Worker Goroutine
	workerCtx, cancelWorker := context.WithCancel(context.Background())
	reminderWorker := worker.NewReminderWorker(reminderSvc, fcmClient, 15*time.Second)
	go reminderWorker.Start(workerCtx)

	// 9. Graceful Shutdown listener
	shutdownChan := make(chan os.Signal, 1)
	signal.Notify(shutdownChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-shutdownChan
		log.Println("[INFO] Shutting down server gracefully...")
		cancelWorker()
		_ = app.ShutdownWithTimeout(5 * time.Second)
	}()

	// 10. Start Server
	addr := fmt.Sprintf(":%s", cfg.AppPort)
	log.Printf("[INFO] Server starting on port %s (env: %s)...\n", cfg.AppPort, cfg.AppEnv)
	log.Printf("[INFO] Swagger documentation available at http://localhost:%s/swagger/\n", cfg.AppPort)
	if err := app.Listen(addr); err != nil {
		log.Fatalf("[FATAL] Server failed to start: %v\n", err)
	}
}
