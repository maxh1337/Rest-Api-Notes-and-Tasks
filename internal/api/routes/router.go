package routes

import (
	"rest-api-notes/internal/api/handlers"

	"github.com/labstack/echo/v4"
)

func NewRoutes(e *echo.Echo, userHandler handlers.UserHandler, authHandler handlers.AuthHandler) {
	apiGroup := e.Group("/api/v1")
	RegisterUserRoutes(apiGroup, userHandler)
	RegisterAuthRoutes(apiGroup, authHandler)
}
