package handlers

import (
	"net/http"
	"rest-api-notes/internal/domain/entities"
	"rest-api-notes/internal/domain/services"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type userHandler struct {
	userService      services.UserService
	twoFactorService services.TwoFactorService
}

type UserHandler interface {
	GetProfile(c echo.Context) error
	UpdateTelephoneNumber(c echo.Context) error
	TwoFactorEnableRequest(c echo.Context) error
	VerifyTwoFactorEnableRequest(c echo.Context) error
}

func NewUserHandler(userService services.UserService, twoFactorService services.TwoFactorService) UserHandler {
	return &userHandler{
		userService:      userService,
		twoFactorService: twoFactorService,
	}
}

func (h *userHandler) GetProfile(c echo.Context) error {
	userID := c.Get("user_id").(uuid.UUID)
	if userID == uuid.Nil {
		return c.String(http.StatusBadRequest, "Invalid user ID")
	}

	user, err := h.userService.GetUserProfile(userID)
	if err != nil {
		return c.String(http.StatusNotFound, "User you are looking for not found")
	}

	return c.JSON(http.StatusOK, user)
}

func (h *userHandler) UpdateTelephoneNumber(c echo.Context) error {
	ctx := c.Request().Context()
	userID := c.Get("user_id").(uuid.UUID)
	if userID == uuid.Nil {
		return c.String(http.StatusBadRequest, "Invalid user ID")
	}
	req := &entities.UserUpdatePhoneReq{}

	if err := c.Bind(req); err != nil {
		return c.String(http.StatusBadRequest, "Please provide telephone number")
	}
	if err := c.Validate(req); err != nil {
		return c.String(http.StatusBadRequest, "Invalid telephone number")
	}

	if err := h.userService.UpdateUserPhone(ctx, req, userID); err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	return c.JSON(http.StatusOK, "Successfully updated")
}

// Переделать в Toggle !prev
func (h *userHandler) TwoFactorEnableRequest(c echo.Context) error {
	ctx := c.Request().Context()
	userID := c.Get("user_id").(uuid.UUID)
	if userID == uuid.Nil {
		return c.String(http.StatusBadRequest, "Invalid user ID")
	}

	if err := h.userService.TwoFactorEnableRequest(ctx, userID); err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	return c.JSON(http.StatusOK, "Code was sent")
}

// Переделать в Toggle !prev
func (h *userHandler) VerifyTwoFactorEnableRequest(c echo.Context) error {
	ctx := c.Request().Context()
	userID := c.Get("user_id").(uuid.UUID)
	if userID == uuid.Nil {
		return c.JSON(http.StatusBadRequest, "Invalid user ID")
	}

	code := c.Param("code")
	if code == "" {
		return c.JSON(http.StatusBadRequest, "Please provide verification code")
	}

	if err := h.userService.VerifyTwoFactorEnableRequest(ctx, userID, code); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	return c.JSON(http.StatusOK, "2FA was enabled successfully ")
}
