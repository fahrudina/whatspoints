package config

import (
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type EnvConfig struct {
	DBHost              string
	DBPort              string
	DBUsername          string
	DBPassword          string
	DBName              string
	DBSSLMode           string // Add SSL mode for Supabase
	AWSRegion           string
	S3BucketName        string
	AllowedPhoneNumbers map[string]bool
}

// Global variable to hold the loaded environment configuration
var Env EnvConfig

// LoadEnv loads and validates all required environment variables
func LoadEnv() {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		// Don't panic if .env file doesn't exist, just log a warning
		log.Printf("Warning: .env file not found, using system environment variables: %v", err)
	}

	Env = EnvConfig{
		DBHost:              getEnv("SUPABASE_HOST", ""),
		DBPort:              getEnv("SUPABASE_PORT", "5432"),
		DBUsername:          getEnv("SUPABASE_USER", ""),
		DBPassword:          getEnv("SUPABASE_PASSWORD", ""),
		DBName:              getEnv("SUPABASE_DB", ""),
		DBSSLMode:           getEnv("SUPABASE_SSLMODE", "require"),
		AWSRegion:           getEnv("AWS_REGION", ""),
		S3BucketName:        getEnv("S3_BUCKET_NAME", ""),
		AllowedPhoneNumbers: parseAllowedPhoneNumbers(getEnv("ALLOWED_PHONE_NUMBERS", "")),
	}

	// Only validate AWS variables if they are actually needed (when S3 functionality is used)
	// For now, we'll make them optional to allow the app to start without AWS configuration
	if Env.AWSRegion != "" && Env.S3BucketName == "" {
		log.Printf("Warning: AWS_REGION is set but S3_BUCKET_NAME is missing. S3 functionality may not work properly.")
	}
	if Env.S3BucketName != "" && Env.AWSRegion == "" {
		log.Printf("Warning: S3_BUCKET_NAME is set but AWS_REGION is missing. S3 functionality may not work properly.")
	}
}

// getEnv retrieves the value of the environment variable or returns a default value if not set
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// parseAllowedPhoneNumbers parses a comma-separated string into a map
func parseAllowedPhoneNumbers(csv string) map[string]bool {
	phoneNumbers := strings.Split(csv, ",")
	allowed := make(map[string]bool)
	for _, phone := range phoneNumbers {
		trimmed := strings.TrimSpace(phone)
		if trimmed != "" {
			allowed[trimmed] = true
		}
	}
	return allowed
}
