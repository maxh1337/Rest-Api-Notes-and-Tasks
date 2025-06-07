package handlers

import (
	"log"
	"net/http"
	"rest-api-notes/internal/config"
	"rest-api-notes/internal/domain/entities"
	"rest-api-notes/internal/domain/services"
	"rest-api-notes/internal/infrastructure/auth"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type authHandler struct {
	authService services.AuthService
	cfg         *config.Config
}

type AuthHandler interface {
	Login(c echo.Context) error
	Register(c echo.Context) error
	Logout(c echo.Context) error
	GetNewTokens(c echo.Context) error
	Verify2FA(c echo.Context) error
	Resend2FA(c echo.Context) error
}

func NewAuthHandler(authService services.AuthService, config *config.Config) AuthHandler {
	return &authHandler{authService: authService, cfg: config}
}

func (h *authHandler) Login(c echo.Context) error {
	ctx := c.Request().Context()
	req := new(entities.UserLoginReq)

	if err := c.Bind(req); err != nil {
		return entities.NewAPIError(entities.ErrorCodeInvalidInput, "Invalid request format")
	}

	if err := c.Validate(req); err != nil {
		return err
	}

	userAgent := c.Request().UserAgent()
	userIP := c.RealIP()

	res, session, err := h.authService.Login(ctx, req, userAgent, userIP)
	if err != nil {
		if err == entities.Err2FARequired {
			if res != nil && res.TwoFactorToken != "" {
				if cookieErr := setTwoFactorCookieToResponse(c, h.cfg, res.TwoFactorToken); cookieErr != nil {
					return entities.ConvertError(cookieErr)
				}
			}
			return entities.NewTwoFactorRequiredError()
		}
		return entities.ConvertError(err)
	}

	setCookiesToResponse(c, session, h.cfg)
	return c.JSON(http.StatusOK, res)
}

func (h *authHandler) Register(c echo.Context) error {
	ctx := c.Request().Context()
	req := new(entities.UserRegisterReq)

	if err := c.Bind(req); err != nil {
		return entities.NewAPIError(entities.ErrorCodeInvalidInput, "Invalid request format")
	}

	if err := c.Validate(req); err != nil {
		return err
	}

	userAgent := c.Request().UserAgent()
	userIP := c.RealIP()

	res, session, err := h.authService.Register(ctx, req, userAgent, userIP)
	if err != nil {
		return entities.ConvertError(err)
	}

	setCookiesToResponse(c, session, h.cfg)
	return c.JSON(http.StatusCreated, res)
}

func (h *authHandler) Logout(c echo.Context) error {
	ctx := c.Request().Context()

	refreshToken, err := c.Request().Cookie("refreshToken")
	if err != nil {
		removeCookiesFromResponse(c, h.cfg)
		return entities.ConvertError(entities.ErrRefreshTokenNotProvided)
	}

	sessionID, err := c.Request().Cookie("session_id")
	if err != nil {
		removeCookiesFromResponse(c, h.cfg)
		return entities.ConvertError(entities.ErrSessionIDTokenNotProvided)
	}

	req := entities.UserLogoutReq{
		RefreshToken: refreshToken.Value,
		SessionID:    sessionID.Value,
	}

	if err := c.Validate(&req); err != nil {
		removeCookiesFromResponse(c, h.cfg)
		return err
	}

	if err := h.authService.Logout(ctx, &req); err != nil {
		removeCookiesFromResponse(c, h.cfg)
		return entities.ConvertError(err)
	}

	removeCookiesFromResponse(c, h.cfg)
	return c.JSON(http.StatusOK, map[string]string{
		"message": "Successfully logged out",
	})
}

func (h *authHandler) GetNewTokens(c echo.Context) error {
	ctx := c.Request().Context()
	refreshToken, err := c.Request().Cookie("refreshToken")
	if err != nil {
		return c.JSON(http.StatusUnauthorized, entities.ErrRefreshTokenNotProvided.Error())
	}

	sessionId, err := c.Request().Cookie("session_id")
	if err != nil {
		return c.JSON(http.StatusUnauthorized, entities.ErrRefreshTokenNotProvided.Error())
	}

	userAgent := c.Request().UserAgent()
	userIp := c.RealIP()

	req := &entities.UserGetNewTokensReq{
		RefreshToken: refreshToken.Value,
		SessionID:    sessionId.Value,
	}

	session, err := h.authService.GetNewTokens(ctx, req, userAgent, userIp)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	setCookiesToResponse(c, session, h.cfg)

	return c.JSON(http.StatusOK, map[string]string{"session_id": session.SessionID})
}

func (h *authHandler) Verify2FA(c echo.Context) error {
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
		req := new(entities.Verify2FACodeReq)
		if err := c.Bind(req); err != nil {
			return entities.NewAPIError(entities.ErrorCodeInvalidInput, "Invalid request format")
		}
		if err := c.Validate(req); err != nil {
			return err
		}
		code = req.Code
	}

	if code == "" {
		return entities.NewAPIError(entities.ErrorCode2FACodeInvalid, "2FA code is required")
	}

	twoFactorCookie, err := c.Cookie("2fa_token")
	if err != nil {
		return entities.NewAPIError(entities.ErrorCodeRefreshTokenMissing, "2FA token is required")
	}

	userAgent := c.Request().UserAgent()
	userIP := c.RealIP()

	session, err := h.authService.Verify2FACode(ctx, code, twoFactorCookie.Value, userAgent, userIP, userID)
	if err != nil {
		return entities.ConvertError(err)
	}

	setCookiesToResponse(c, session, h.cfg)
	removeTwoFactorCookieFromResponse(c, h.cfg)

	return c.JSON(http.StatusOK, map[string]string{
		"message": "2FA verification successful",
	})
}

func (h *authHandler) Resend2FA(c echo.Context) error {
	ctx := c.Request().Context()
	userAgent := c.Request().UserAgent()
	userIP := c.RealIP()

	userIDInterface := c.Get("user_id")
	if userIDInterface == nil {
		return entities.NewAPIError(entities.ErrorCodeUnauthorized, "User not authenticated")
	}

	userID, ok := userIDInterface.(uuid.UUID)
	if !ok {
		return entities.NewAPIError(entities.ErrorCodeUnauthorized, "Invalid user ID")
	}

	newToken, err := h.authService.Resend2FACode(ctx, userID, userAgent, userIP)
	if err != nil {
		return entities.ConvertError(err)
	}

	if err := setTwoFactorCookieToResponse(c, h.cfg, *newToken); err != nil {
		return entities.ConvertError(err)
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "2FA code resent successfully",
	})
}

func setCookiesToResponse(c echo.Context, session *entities.Session, cfg *config.Config) error {
	isProduction := cfg.NODE_ENV == "production"
	log.Printf("Access Exp: %+v", cfg)
	c.SetCookie(&http.Cookie{
		Name:     auth.CookieTokenAccess,
		Value:    session.AccessToken,
		Expires:  time.Now().Add(time.Duration(cfg.JWT.JWT_ACCESS_EXPIRATION) * time.Hour),
		HttpOnly: true,
		Secure:   isProduction,
		SameSite: getSameSiteMode(isProduction),
		Path:     cfg.JWT.JWT_PATH,
		Domain:   cfg.JWT.JWT_DOMAIN,
	})

	c.SetCookie(&http.Cookie{
		Name:     auth.CookieTokenRefresh,
		Value:    session.RefreshToken,
		Expires:  time.Unix(session.ExpiresAt, 0),
		HttpOnly: true,
		Secure:   isProduction,
		SameSite: getSameSiteMode(isProduction),
		Path:     cfg.JWT.JWT_PATH,
		Domain:   cfg.JWT.JWT_DOMAIN,
	})

	c.SetCookie(&http.Cookie{
		Name:     auth.CookieTokenSessionID,
		Value:    session.SessionID,
		Expires:  time.Unix(session.ExpiresAt, 0),
		HttpOnly: true,
		Secure:   isProduction,
		SameSite: getSameSiteMode(isProduction),
		Path:     cfg.JWT.JWT_PATH,
		Domain:   cfg.JWT.JWT_DOMAIN,
	})

	return nil
}

func removeCookiesFromResponse(c echo.Context, cfg *config.Config) error {
	isProduction := cfg.NODE_ENV == "production"
	c.SetCookie(&http.Cookie{
		Name:     auth.CookieTokenAccess,
		Value:    "",
		Expires:  time.Now().Add(-1 * time.Hour),
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   isProduction,
		SameSite: getSameSiteMode(isProduction),
		Path:     cfg.JWT.JWT_PATH,
		Domain:   cfg.JWT.JWT_DOMAIN,
	})

	c.SetCookie(&http.Cookie{
		Name:     auth.CookieTokenRefresh,
		Value:    "",
		Expires:  time.Now().Add(-1 * time.Hour),
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   isProduction,
		SameSite: getSameSiteMode(isProduction),
		Path:     cfg.JWT.JWT_PATH,
		Domain:   cfg.JWT.JWT_DOMAIN,
	})

	c.SetCookie(&http.Cookie{
		Name:     auth.CookieTokenSessionID,
		Value:    "",
		Expires:  time.Now().Add(-1 * time.Hour),
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   isProduction,
		SameSite: getSameSiteMode(isProduction),
		Path:     cfg.JWT.JWT_PATH,
		Domain:   cfg.JWT.JWT_DOMAIN,
	})

	return nil
}

func setTwoFactorCookieToResponse(c echo.Context, cfg *config.Config, token string) error {
	isProduction := cfg.NODE_ENV == "production"
	c.SetCookie(&http.Cookie{
		Name:     auth.CookieToken2Fa,
		Value:    token,
		Expires:  time.Now().Add(time.Duration(cfg.JWT.JWT_2FA_EXPIRATION) * time.Minute),
		HttpOnly: true,
		Secure:   isProduction,
		SameSite: getSameSiteMode(isProduction),
		Path:     cfg.JWT.JWT_PATH,
		Domain:   cfg.JWT.JWT_DOMAIN,
	})
	return nil
}

func removeTwoFactorCookieFromResponse(c echo.Context, cfg *config.Config) error {
	isProduction := cfg.NODE_ENV == "production"
	c.SetCookie(&http.Cookie{
		Name:     auth.CookieToken2Fa,
		Value:    "",
		Expires:  time.Now().Add(-1 * time.Hour),
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   isProduction,
		SameSite: getSameSiteMode(isProduction),
		Path:     cfg.JWT.JWT_PATH,
		Domain:   cfg.JWT.JWT_DOMAIN,
	})

	return nil
}

func getSameSiteMode(isProduction bool) http.SameSite {
	if isProduction {
		return http.SameSiteLaxMode
	}
	return http.SameSiteNoneMode
}
