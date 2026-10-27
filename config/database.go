package config

import (
	"fmt"
	"log"
	"time"

	"todo-backend/internal/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func ConnectDB(cfg *Config) *gorm.DB {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=%s",
		cfg.DBHost,
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBName,
		cfg.DBPort,
		cfg.DBSSLMode,
		cfg.DBTimezone,
	)

	gormLogLevel := logger.Info
	if cfg.AppEnv == "production" {
		gormLogLevel = logger.Error
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(gormLogLevel),
	})
	if err != nil {
		log.Fatalf("[FATAL] Failed to connect to database: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("[FATAL] Failed to get database instance: %v", err)
	}

	// Connection pool settings
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	log.Println("[INFO] Successfully connected to PostgreSQL")

	// Run AutoMigrate
	log.Println("[INFO] Running database auto-migration...")
	if err := db.AutoMigrate(
		&models.User{},
		&models.Task{},
		&models.Reminder{},
		&models.ChatMessage{},
	); err != nil {
		log.Fatalf("[FATAL] AutoMigration failed: %v", err)
	}
	log.Println("[INFO] Database auto-migration completed successfully")

	DB = db
	return DB
}
