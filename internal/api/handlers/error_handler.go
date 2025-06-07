package handlers

import (
	"net/http"

	"rest-api-notes/internal/domain/entities"

	"github.com/labstack/echo/v4"
)

func ErrorHandler(err error, c echo.Context) {
	convertedErr := entities.ConvertError(err)

	var (
		code     int
		response interface{}
	)

	switch e := convertedErr.(type) {
	case *entities.APIError:
		code = getHTTPStatusFromErrorCode(e.Code)
		response = e

	case *entities.ValidationError:
		code = http.StatusBadRequest
		response = e

	case *entities.TwoFactorRequiredError:
		code = http.StatusUnauthorized
		response = e

	case *echo.HTTPError:
		code = e.Code
		response = entities.NewAPIError(
			entities.ErrorCodeInternalError,
			e.Message.(string),
		)

	default:
		code = http.StatusInternalServerError
		response = entities.NewAPIError(
			entities.ErrorCodeInternalError,
			"Internal server error",
		)
	}

	if code >= 500 {
		c.Logger().Error(err)
	}

	if !c.Response().Committed {
		if c.Request().Method == echo.HEAD {
			c.NoContent(code)
		} else {
			c.JSON(code, response)
		}
	}
}

func getHTTPStatusFromErrorCode(errorCode string) int {
	switch errorCode {
	case entities.ErrorCodeUnauthorized,
		entities.ErrorCode2FARequired,
		entities.ErrorCodeRefreshTokenMissing,
		entities.ErrorCodeSessionIDMissing:
		return http.StatusUnauthorized

	case entities.ErrorCodeForbidden:
		return http.StatusForbidden

	case entities.ErrorCodeUserNotFound,
		entities.ErrorCodeSessionNotFound:
		return http.StatusNotFound

	case entities.ErrorCodeEmailTaken,
		entities.ErrorCodeUsernameTaken:
		return http.StatusConflict

	case entities.ErrorCodeInternalError:
		return http.StatusInternalServerError

	default:
		return http.StatusBadRequest
	}
}
