package config

import (
	"os"
	"strings"
)

type EnvConfig struct {
	DBHost              string
	DBPort              string
	DBUsername          string
	DBPassword          string
	DBName              string
	AWSRegion           string
	S3BucketName        string
	AllowedPhoneNumbers map[string]bool
}

// Global variable to hold the loaded environment configuration
var Env EnvConfig

// LoadEnv loads and validates all required environment variables
func LoadEnv() {
	Env = EnvConfig{
		DBHost:              getEnv("DB_HOST", "localhost"),
		DBPort:              getEnv("DB_PORT", "3306"),
		DBUsername:          getEnv("DB_USERNAME", "username"),
		DBPassword:          getEnv("DB_PASSWORD", "password"),
		DBName:              getEnv("DB_NAME", "wa_serv_db"),
		AWSRegion:           getEnv("AWS_REGION", ""),
		S3BucketName:        getEnv("S3_BUCKET_NAME", ""),
		AllowedPhoneNumbers: parseAllowedPhoneNumbers(getEnv("ALLOWED_PHONE_NUMBERS", "")),
	}

	// Validate required environment variables
	if Env.AWSRegion == "" {
		panic("AWS_REGION environment variable is required but not set")
	}
	if Env.S3BucketName == "" {
		panic("S3_BUCKET_NAME environment variable is required but not set")
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
