package routes

import (
	"rest-api-notes/internal/api/handlers"
	"rest-api-notes/internal/api/middleware"

	"github.com/labstack/echo/v4"
)

func RegisterAuthRoutes(g *echo.Group, authHandler handlers.AuthHandler, m *middleware.MiddlewareManager) {
	g.POST("/login", authHandler.Login)
	g.POST("/register", authHandler.Register)
	g.POST("/logout", authHandler.Logout)
	g.POST("/token/refresh", authHandler.GetNewTokens, m.RateLimit(5))
	g.POST("/2fa/verify/:code", authHandler.Verify2FA, m.TwoFactorTokenCheck(), m.RateLimit(5))
	g.POST("/2fa/resend", authHandler.Resend2FA, m.TwoFactorTokenCheck(), m.RateLimit(5))
}
