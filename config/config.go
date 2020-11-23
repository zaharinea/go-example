package config

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

// Config struct
type Config struct {
	AppVersion         string
	AppHost            string
	AppPort            string
	AppAddr            string
	MongoURI           string
	MongoDbName        string
	MongoMigrationsDir string
	PageSize           int64
	LogLevel           string
	LogFormat          string
	SentryDSN          string
}

// Simple helper function to read an environment or return error
func getRequiredEnv(key string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		logrus.Fatalf("Environment variable not set: %s", key)
	}
	return value
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
	err := godotenv.Load()
	if err != nil {
		logrus.Info("Error loading .env file")
	}

	appHost := getEnv("APP_HOST", "0.0.0.0")
	appPort := getEnv("APP_PORT", "8000")
	return &Config{
		AppVersion:         getEnv("APP_VERSION", "0.0.0"),
		AppHost:            appHost,
		AppPort:            appPort,
		AppAddr:            net.JoinHostPort(appHost, appPort),
		MongoURI:           getRequiredEnv("MONGODB_CONNECTION_STRING"),
		MongoDbName:        getRequiredEnv("MONGO_DBNAME"),
		MongoMigrationsDir: "file://migrations",
		PageSize:           25,
		LogLevel:           getEnv("LOGS_LEVEL", "INFO"),
		LogFormat:          getEnv("LOGS_FORMAT", "TEXT"),
		SentryDSN:          getEnv("SENTRY_DSN", ""),
	}
}

// NewTestingConfig returns a new Config struct for tests
func NewTestingConfig() *Config {
	config := NewConfig()
	config.MongoDbName = getEnv("MONGO_DBNAME_TEST", fmt.Sprintf("%s_test", config.MongoDbName))
	config.MongoMigrationsDir = "file://../../migrations"
	return config
}
