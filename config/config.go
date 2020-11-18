package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Config struct
type Config struct {
	AppHost  string
	AppPort  string
	AppAddr  string
	MongoURI string
	MongoDB  string
	PageSize int64
}

// Simple helper function to read an environment or return a default value
func getEnv(key string, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}

	return defaultVal
}

// Simple helper function to read an environment variable into integer or return a default value
func getEnvAsInt(name string, defaultVal int) int {
	valueStr := getEnv(name, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}

	return defaultVal
}

// Helper to read an environment variable into a bool or return default value
func getEnvAsBool(name string, defaultVal bool) bool {
	valStr := getEnv(name, "")
	if val, err := strconv.ParseBool(valStr); err == nil {
		return val
	}

	return defaultVal
}

// Helper to read an environment variable into a string slice or return default value
func getEnvAsSlice(name string, defaultVal []string, sep string) []string {
	valStr := getEnv(name, "")

	if valStr == "" {
		return defaultVal
	}

	val := strings.Split(valStr, sep)

	return val
}

// NewConfig returns a new Config struct
func NewConfig() *Config {
	appHost := getEnv("APP_HOST", "0.0.0.0")
	appPort := getEnv("APP_PORT", "8000")
	return &Config{
		AppHost:  appHost,
		AppPort:  appPort,
		AppAddr:  fmt.Sprintf("%s:%s", appHost, appPort),
		MongoURI: os.Getenv("MONGODB_CONNECTION_STRING"),
		MongoDB:  os.Getenv("MONGO_DBNAME"),
		PageSize: 25,
	}
}
