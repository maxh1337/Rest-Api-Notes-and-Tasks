package routes

import (
	"rest-api-notes/internal/api/handlers"

	"github.com/labstack/echo/v4"
)

func RegisterUserRoutes(g *echo.Group, handlers handlers.UserHandler) {
	g.GET("/get-profile", handlers.GetProfile)
	g.POST("/update-phone", handlers.UpdateTelephoneNumber)
	g.POST("/request-enable-2fa", handlers.TwoFactorEnableRequest)
	g.POST("/verify-enable-2fa/:code", handlers.VerifyTwoFactorEnableRequest)
}
