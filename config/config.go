package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DBHost     string
	DBPort     string
	DBUser     string
	DBPass     string
	DBName     string
	DBSSLMode  string
	JWTSecret  string
	ServerPort string
	SMTPHost   string
	SMTPPort   string
	SMTPUser   string
	SMTPPass   string
	SMTPSender string
}

var AppConfig *Config

func LoadConfig() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: .env file not found, using system environment variables")
	}

	AppConfig = &Config{
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPass:     getEnv("DB_PASSWORD", ""),
		DBName:     getEnv("DB_NAME", "internship_hub"),
		DBSSLMode:  getEnv("DB_SSLMODE", "disable"),
		JWTSecret:  getEnv("JWT_SECRET", "secret"),
		ServerPort: getEnv("PORT", "8080"),
		SMTPHost:   getEnv("SMTP_HOST", "localhost"),
		SMTPPort:   getEnv("SMTP_PORT", "1025"),
		SMTPUser:   getEnv("SMTP_USER", ""),
		SMTPPass:   getEnv("SMTP_PASSWORD", ""),
		SMTPSender: getEnv("SMTP_SENDER", "no-reply@internshiphub.com"),
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
