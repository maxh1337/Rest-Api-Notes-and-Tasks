package routes

import (
	"rest-api-notes/internal/api/handlers"
	"rest-api-notes/internal/api/middleware"

	"github.com/labstack/echo/v4"
)

func RegisterUserRoutes(g *echo.Group, handlers handlers.UserHandler, m *middleware.MiddlewareManager) {
	g.GET("/get-profile", handlers.GetProfile)
	g.POST("/update-phone", handlers.UpdateTelephoneNumber)

	//2FA
	g.POST("/2fa/request-toggle", handlers.TwoFactorToggleRequest, m.RateLimit(1))
	g.POST("/2fa/verify-code/:code", handlers.VerifyTwoFactorToggleRequest, m.RateLimit(3))
	g.POST("/2fa/resend-code", handlers.ResendTwoFactorCode, m.RateLimit(3))
}
