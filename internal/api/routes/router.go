package routes

import (
	"rest-api-notes/internal/api/handlers"

	"github.com/labstack/echo/v4"
)

func NewRoutes(e *echo.Echo, handlers handlers.UserHandler) {
	apiGroup := e.Group("/api/v1")
	RegisterUserRoutes(apiGroup, handlers)
}
