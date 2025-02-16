package main

import (
	"context"
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/1way-market/v3/internal/config"
	"github.com/1way-market/v3/internal/database"
	"github.com/1way-market/v3/internal/delivery/http/router"
	"github.com/1way-market/v3/internal/repository"
	"github.com/1way-market/v3/internal/usecase"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	_ "github.com/lib/pq"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func initDatabase(cfg *config.Config) (*gorm.DB, error) {
	// First, try to connect to PostgreSQL server
	sqlDB, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("error connecting to PostgreSQL: %v", err)
	}

	// Try to connect to the specific database
	err = sqlDB.Ping()
	if err != nil {
		// If database doesn't exist, create it
		sqlDB.Close()

		// Connect to postgres database to create our database
		postgresDB, err := sql.Open("postgres", strings.Replace(cfg.DatabaseURL, cfg.DBName, "postgres", 1))
		if err != nil {
			return nil, fmt.Errorf("error connecting to postgres database: %v", err)
		}
		defer postgresDB.Close()

		_, err = postgresDB.Exec("CREATE DATABASE " + cfg.DBName)
		if err != nil && !strings.Contains(err.Error(), "already exists") {
			return nil, fmt.Errorf("error creating database: %v", err)
		}

		// Reconnect to the newly created database
		sqlDB, err = sql.Open("postgres", cfg.DatabaseURL)
		if err != nil {
			return nil, fmt.Errorf("error connecting to new database: %v", err)
		}
	}

	// Initialize GORM
	gormDB, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("error initializing GORM: %v", err)
	}

	// Validate schema
	if err := database.ValidateSchema(sqlDB); err != nil {
		// If tables don't exist, run migrations
		if strings.Contains(err.Error(), "does not exist") {
			log.Printf("Database schema not found, running migrations...")
			migrationSQL, err := ioutil.ReadFile(filepath.Join("migrations", "001_initial_schema.sql"))
			if err != nil {
				return nil, fmt.Errorf("error reading migration file: %v", err)
			}

			if err := gormDB.Exec(string(migrationSQL)).Error; err != nil {
				return nil, fmt.Errorf("error running migrations: %v", err)
			}

			// Validate schema again after migration
			if err := database.ValidateSchema(sqlDB); err != nil {
				return nil, fmt.Errorf("schema validation failed after migration: %v", err)
			}
		} else {
			// If schema validation failed for other reasons, return the error
			return nil, fmt.Errorf("schema validation failed: %v", err)
		}
	}

	return gormDB, nil
}

func initRedis(cfg *config.Config) (*redis.Client, error) {
	opt, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		return nil, fmt.Errorf("error parsing Redis URL: %v", err)
	}

	client := redis.NewClient(opt)
	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("error connecting to Redis: %v", err)
	}

	return client, nil
}

func main() {
	// Initialize configuration
	cfg := config.New()

	// Initialize database
	db, err := initDatabase(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Initialize Redis
	redisClient, err := initRedis(cfg)
	if err != nil {
		log.Printf("Warning: Failed to initialize Redis: %v", err)
	}

	// Initialize repositories
	repos := repository.NewRepositories(db)

	// Initialize use cases
	useCases := usecase.NewUseCases(repos, redisClient)

	// Initialize Gin router
	gin.SetMode(gin.ReleaseMode)
	r := router.Setup(useCases)

	// Create HTTP server
	srv := &http.Server{
		Addr:    cfg.ServerAddress,
		Handler: r,
	}

	// Graceful shutdown
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	log.Printf("Server is running on %s", cfg.ServerAddress)

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// Shutdown server
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}
}
