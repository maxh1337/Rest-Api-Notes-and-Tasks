package routes

import (
	"rest-api-notes/internal/api/handlers"
	"rest-api-notes/internal/api/middleware"
	"rest-api-notes/internal/config"
	"rest-api-notes/internal/infrastructure/auth"

	"github.com/labstack/echo/v4"
)

func SetupRoutes(e *echo.Echo, cfg *config.Config, jwtService auth.JWTService,
	userHandler handlers.UserHandler, authHandler handlers.AuthHandler) {
	apiGroup := e.Group("/api/v1")

	mM := middleware.NewMiddlewareManager(cfg, jwtService)
	// // Global Middleware
	e.Use(mM.StrictCORS(), mM.RateLimit(6000))

	// User group
	userGroup := apiGroup.Group("/users")
	// Middleware for user group
	userGroup.Use(mM.RequireAuth())
	// User group routes
	RegisterUserRoutes(userGroup, userHandler, mM)

	// Authorization group
	authGroup := apiGroup.Group("/auth")
	// Middleware for auth group
	authGroup.Use(mM.RateLimit(10))
	// Auth group routes
	RegisterAuthRoutes(authGroup, authHandler, mM)
}
