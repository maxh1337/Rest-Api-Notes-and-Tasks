package routes

import (
	"rest-api-notes/internal/api/handlers"

	"github.com/labstack/echo/v4"
)

func RegisterUserRoutes(g *echo.Group, handlers handlers.UserHandler) {
	userGroup := g.Group("/users")

	// Apply auth_middleware for entire group

	userGroup.GET("/get-profile", handlers.GetProfile)
}
