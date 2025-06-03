package handlers

import (
	"net/http"
	"rest-api-notes/internal/domain/services"

	"github.com/labstack/echo/v4"
)

type userHandler struct {
	userService services.UserService
}

type UserHandler interface {
	GetProfile(c echo.Context) error
}

func NewUserHandler(userService services.UserService) UserHandler {
	return &userHandler{
		userService: userService,
	}
}

func (h userHandler) GetProfile(c echo.Context) error {
	user, err := h.userService.GetUserProfile("1")
	if err != nil {
		return c.String(http.StatusNotFound, "User you are looking for not found")
	}

	return c.JSON(http.StatusOK, user)
}
