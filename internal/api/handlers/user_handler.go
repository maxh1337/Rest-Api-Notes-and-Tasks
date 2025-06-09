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
	TwoFactorToggleRequest(c echo.Context) error
	VerifyTwoFactorToggleRequest(c echo.Context) error
	ResendTwoFactorCode(c echo.Context) error
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

func (h *userHandler) TwoFactorToggleRequest(c echo.Context) error {
	ctx := c.Request().Context()
	userIDInterface := c.Get("user_id")
	if userIDInterface == nil {
		return entities.NewAPIError(entities.ErrorCodeUnauthorized, "User not authenticated")
	}

	userID, ok := userIDInterface.(uuid.UUID)
	if !ok {
		return entities.NewAPIError(entities.ErrorCodeUnauthorized, "Invalid user ID")
	}

	if err := h.userService.TwoFactorToggleRequest(ctx, userID); err != nil {
		return entities.ConvertError(err)
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "2FA toggle code sent successfully",
	})
}

func (h *userHandler) VerifyTwoFactorToggleRequest(c echo.Context) error {
	ctx := c.Request().Context()
	userIDInterface := c.Get("user_id")
	if userIDInterface == nil {
		return entities.NewAPIError(entities.ErrorCodeUnauthorized, "User not authenticated")
	}

	userID, ok := userIDInterface.(uuid.UUID)
	if !ok {
		return entities.NewAPIError(entities.ErrorCodeUnauthorized, "Invalid user ID")
	}

	code := c.Param("code")
	if code == "" {
		return entities.NewAPIError(entities.ErrorCode2FACodeInvalid, "2FA code is required")
	}

	if err := h.userService.VerifyTwoFactorToggleRequest(ctx, userID, code); err != nil {
		return entities.ConvertError(err)
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "2FA toggle code was toggled successfully",
	})
}

func (h *userHandler) ResendTwoFactorCode(c echo.Context) error {
	ctx := c.Request().Context()

	userIDInterface := c.Get("user_id")
	if userIDInterface == nil {
		return entities.NewAPIError(entities.ErrorCodeUnauthorized, "User not authenticated")
	}

	userID, ok := userIDInterface.(uuid.UUID)
	if !ok {
		return entities.NewAPIError(entities.ErrorCodeUnauthorized, "Invalid user ID")
	}

	if err := h.userService.ResendTwoFactorCode(ctx, userID); err != nil {
		return entities.ConvertError(err)
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "2FA toggle code resent successfully",
	})
}
