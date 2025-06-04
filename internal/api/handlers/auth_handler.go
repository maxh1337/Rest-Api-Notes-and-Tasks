package handlers

import (
	"net/http"
	"rest-api-notes/internal/config"
	"rest-api-notes/internal/domain/entities"
	"rest-api-notes/internal/domain/services"
	"rest-api-notes/internal/infrastructure/auth"
	"time"

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
}

func NewAuthHandler(authService services.AuthService, config *config.Config) AuthHandler {
	return &authHandler{authService: authService, cfg: config}
}

func (h *authHandler) Login(c echo.Context) error {
	ctx := c.Request().Context()
	req := new(entities.UserLoginReq)
	err := c.Bind(req)
	if err != nil {
		return err
	}
	userAgent := c.Request().UserAgent()
	userIp := c.RealIP()

	err = c.Validate(req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	res, session, err := h.authService.Login(ctx, req, userAgent, userIp)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	setCookiesToResponse(c, session, h.cfg)

	return c.JSON(http.StatusOK, res)
}

func (h *authHandler) Register(c echo.Context) error {
	ctx := c.Request().Context()
	req := new(entities.UserRegisterReq)
	err := c.Bind(req)
	if err != nil {
		return err
	}
	userAgent := c.Request().UserAgent()
	userIp := c.RealIP()

	err = c.Validate(req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	res, session, err := h.authService.Register(ctx, req, userAgent, userIp)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	setCookiesToResponse(c, session, h.cfg)

	return c.JSON(http.StatusCreated, res)
}

func (h *authHandler) Logout(c echo.Context) error {
	ctx := c.Request().Context()
	refreshToken, err := c.Request().Cookie("refreshToken")
	if err != nil {
		removeCookiesFromResponse(c, h.cfg)
		return entities.ErrFailedToLoginRefreshTokenNotProvide
	}

	sessionId, err := c.Request().Cookie("session_id")
	if err != nil {
		removeCookiesFromResponse(c, h.cfg)
		return entities.ErrFailedToLoginRefreshTokenNotProvide
	}

	req := entities.UserLogoutReq{
		RefreshToken: refreshToken.Value,
		SessionID:    sessionId.Value,
	}

	err = h.authService.Logout(ctx, &req)
	removeCookiesFromResponse(c, h.cfg)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, "Success")
}

func setCookiesToResponse(c echo.Context, session *entities.Session, cfg *config.Config) error {
	isProduction := cfg.NODE_ENV == "production"
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
