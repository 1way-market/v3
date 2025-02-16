package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	ServerAddress string
	DatabaseURL   string
	RedisURL      string
	Environment   string
	DBName        string
}

func New() *Config {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		fmt.Printf("Warning: .env file not found: %v\n", err)
	}

	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "postgres")
	dbPass := getEnv("DB_PASSWORD", "postgres")
	dbName := getEnv("DB_NAME", "market")
	dbSSLMode := getEnv("DB_SSLMODE", "disable")

	dbURL := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		dbHost, dbPort, dbUser, dbPass, dbName, dbSSLMode)

	redisHost := getEnv("REDIS_HOST", "localhost")
	redisPort := getEnv("REDIS_PORT", "6379")
	redisDB := getEnv("REDIS_DB", "0")
	redisPass := getEnv("REDIS_PASSWORD", "")

	redisURL := fmt.Sprintf("redis://%s:%s@%s:%s/%s",
		redisPass, redisPass, redisHost, redisPort, redisDB)

	return &Config{
		ServerAddress: getEnv("SERVER_ADDRESS", ":8080"),
		DatabaseURL:   dbURL,
		RedisURL:      redisURL,
		Environment:   getEnv("ENVIRONMENT", "development"),
		DBName:        dbName,
	}
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
