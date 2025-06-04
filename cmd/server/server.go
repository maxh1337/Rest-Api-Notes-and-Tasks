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
	"rest-api-notes/internal/domain/validator"
	"rest-api-notes/internal/infrastructure/auth"
	"rest-api-notes/internal/infrastructure/cache"
	"rest-api-notes/internal/infrastructure/database"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
)

func main() {
	e := echo.New()
	e.Validator = validator.NewValidator()

	if err := godotenv.Load(); err != nil {
		log.Fatalf("No .env file found: %v", err)
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
		log.Fatalf("Failed to connect to database: %v", err)
	}

	client, err := cache.NewRedisClient(&cfg.Redis)

	if err != nil {
		log.Fatal("Failed to connect to redis:", err)
	}

	// var origins = []string{
	// 	"*",
	// }

	// e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
	// 	AllowOrigins: origins,
	// }))

	// Non Entity Services if we can say like that
	jwtService := auth.NewJWTService(cfg.JWT)
	passwordService := auth.NewPasswordService()
	sessionService := services.NewSessionService(client, time.Duration(cfg.JWT.JWT_REFRESH_EXPIRATION)*time.Hour)

	// REPOS
	userRepository := repositories.NewUserRepository(db, passwordService)

	// SERVICES
	userService := services.NewUserService(userRepository)
	authService := services.NewAuthService(jwtService, passwordService, userRepository, sessionService)

	// HANDLERS
	userHandler := handlers.NewUserHandler(userService)
	authHandler := handlers.NewAuthHandler(authService, cfg)

	// ROUTES
	routes.NewRoutes(e, userHandler, authHandler)

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
