package config

import (
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	AppPort         string
	AppEnv          string
	DBHost          string
	DBPort          string
	DBUser          string
	DBPassword      string
	DBName          string
	DBSSLMode       string
	DBTimezone      string
	JWTSecret       string
	JWTExpireHr     string
	SwaggerHost     string
	SwaggerBasePath string
	SwaggerSchemes  []string
	GroqAPIKey      string
	GroqModel       string
}

var AppConfig *Config

func LoadConfig() *Config {
	envFiles := []string{".env", "../.env", "../../.env"}
	loaded := false
	for _, f := range envFiles {
		if err := godotenv.Load(f); err == nil {
			log.Printf("[INFO] Loaded environment from %s\n", f)
			loaded = true
			break
		}
	}
	if !loaded {
		log.Println("[INFO] No .env file found, using system environment variables")
	}

	schemesRaw := getEnv("SWAGGER_SCHEMES", "https,http")
	var schemes []string
	if schemesRaw != "" {
		for _, s := range strings.Split(schemesRaw, ",") {
			trimmed := strings.TrimSpace(s)
			if trimmed != "" {
				schemes = append(schemes, trimmed)
			}
		}
	}

	AppConfig = &Config{
		AppPort:         getEnv("APP_PORT", "8080"),
		AppEnv:          getEnv("APP_ENV", "development"),
		DBHost:          getEnv("DB_HOST", "localhost"),
		DBPort:          getEnv("DB_PORT", "5432"),
		DBUser:          getEnv("DB_USER", "postgres"),
		DBPassword:      getEnv("DB_PASSWORD", "postgres"),
		DBName:          getEnv("DB_NAME", "todolist_db"),
		DBSSLMode:       getEnv("DB_SSLMODE", "disable"),
		DBTimezone:      getEnv("DB_TIMEZONE", "Asia/Jakarta"),
		JWTSecret:       getEnv("JWT_SECRET", "super-secret-todolist-jwt-key"),
		JWTExpireHr:     getEnv("JWT_EXPIRE_HOURS", "72"),
		SwaggerHost:     getEnv("SWAGGER_HOST", ""),
		SwaggerBasePath: getEnv("SWAGGER_BASE_PATH", "/todo-list"),
		SwaggerSchemes:  schemes,
		GroqAPIKey:      getEnv("GROQ_API_KEY", ""),
		GroqModel:       getEnv("GROQ_MODEL", "openai/gpt-oss-120b"),
	}

	return AppConfig
}

func getEnv(key, fallback string) string {
	if val, exists := os.LookupEnv(key); exists && val != "" {
		return val
	}
	return fallback
}
