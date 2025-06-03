package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"rest-api-notes/internal/api/handlers"
	"rest-api-notes/internal/api/routes"
	"rest-api-notes/internal/config"
	"rest-api-notes/internal/domain/repositories"
	"rest-api-notes/internal/domain/services"
	"rest-api-notes/internal/infrastructure/database"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
)

func main() {
	e := echo.New()

	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// DB CONNECT
	dbConfig := config.DatabaseConfig{
		Host:     cfg.Database.Host,
		Port:     cfg.Database.Port,
		User:     cfg.Database.User,
		Password: cfg.Database.Password,
		DBName:   cfg.Database.DBName,
		SSLMode:  cfg.Database.SSLMode,
	}

	db, err := database.NewPostgresDB(dbConfig)

	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// REPOS
	userRepository := repositories.NewUserRepository(db)

	// SERVICES
	userService := services.NewUserService(userRepository)

	// HANDLERS
	userHandler := handlers.NewUserHandler(userService)

	// ROUTES
	routes.NewRoutes(e, userHandler)

	// START
	go func() {
		if err := e.Start(":" + cfg.Port); err != nil {
			e.Logger.Infof("Shutting down the server: %v", err)
		}
	}()

	// GRACEFUL SHUTDOWN
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
	}
}
