package config

import (
	"fmt"
	"os"
	"strconv"
)

// DatabaseConfig holds the database connection settings
type DatabaseConfig struct {
	Driver   string
	Host     string
	Port     int
	Username string
	Password string
	DBName   string
}

// GetDatabaseConfig returns the database configuration
func GetDatabaseConfig() DatabaseConfig {
	// Read values from environment variables with defaults
	host := getEnv("DB_HOST", "localhost")
	port := getEnvAsInt("DB_PORT", 3306)
	username := getEnv("DB_USERNAME", "username")
	password := getEnv("DB_PASSWORD", "pass")
	dbName := getEnv("DB_NAME", "dbname")

	return DatabaseConfig{
		Driver:   "mysql",
		Host:     host,
		Port:     port,
		Username: username,
		Password: password,
		DBName:   dbName,
	}
}

// BuildConnectionString builds a connection string for the database
func (c *DatabaseConfig) BuildConnectionString() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true",
		c.Username, c.Password, c.Host, c.Port, c.DBName)
}

// getEnv reads an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// getEnvAsInt reads an environment variable as an integer or returns a default value
func getEnvAsInt(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
