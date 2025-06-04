package routes

import (
	"rest-api-notes/internal/api/handlers"

	"github.com/labstack/echo/v4"
)

func RegisterAuthRoutes(g *echo.Group, authHandler handlers.AuthHandler) {
	authGroup := g.Group("/auth")

	// Apply auth_middleware for entire group

	authGroup.POST("/login", authHandler.Login)
	authGroup.POST("/register", authHandler.Register)
	authGroup.POST("/logout", authHandler.Logout)
}
