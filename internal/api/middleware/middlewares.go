package middleware

import (
	"log"
	"net/http"
	"rest-api-notes/internal/config"
	"rest-api-notes/internal/infrastructure/auth"
	"time"

	"github.com/labstack/echo/v4"
	"golang.org/x/time/rate"
)

type MiddlewareManager struct {
	cfg         *config.Config
	jwtService  auth.JWTService
	rateLimiter *rate.Limiter
}

func NewMiddlewareManager(cfg *config.Config, jwtService auth.JWTService) *MiddlewareManager {
	return &MiddlewareManager{
		cfg:         cfg,
		jwtService:  jwtService,
		rateLimiter: rate.NewLimiter(rate.Every(time.Minute), 100),
	}
}

func (m *MiddlewareManager) StrictCORS() echo.MiddlewareFunc {
	if m.cfg.NODE_ENV == "development" {
		return func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(c echo.Context) error {
				return next(c)
			}
		}
	}
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()
			res := c.Response()

			origin := req.Header.Get("Origin")

			if origin == "" {
				return echo.NewHTTPError(http.StatusForbidden, "Origin header required")
			}

			log.Printf("Origin from insomnia - %v", origin)

			allowedOrigins := []string{
				m.cfg.CLIENT_URL,
			}

			isAllowed := false

			for _, allowedOrigin := range allowedOrigins {
				if origin == allowedOrigin {
					isAllowed = true
					break
				}
			}

			if !isAllowed {
				return echo.NewHTTPError(http.StatusForbidden, "Origin not allowed")
			}

			// Устанавливаем CORS заголовки только для разрешенных Origin
			res.Header().Set("Access-Control-Allow-Origin", origin)
			res.Header().Set("Access-Control-Allow-Credentials", "true")

			// Обрабатываем preflight запросы
			if req.Method == http.MethodOptions {
				res.Header().Set("Access-Control-Allow-Methods",
					"GET, POST, PUT, PATCH, DELETE, OPTIONS")
				res.Header().Set("Access-Control-Allow-Headers",
					"Origin, Content-Type, Accept, Authorization, X-Requested-With")
				res.Header().Set("Access-Control-Max-Age", "86400")
				return c.NoContent(http.StatusNoContent)
			}

			return next(c)
		}
	}
}

func (m *MiddlewareManager) RateLimit(requestsPerHour int) echo.MiddlewareFunc {
	limiter := rate.NewLimiter(rate.Every(time.Hour/time.Duration(requestsPerHour)), requestsPerHour)

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if !limiter.Allow() {
				return c.JSON(http.StatusTooManyRequests, map[string]string{
					"error":   "RATE_LIMIT_EXCEEDED",
					"message": "Too many requests. Please try again later.",
				})
			}
			return next(c)
		}
	}
}

func (m *MiddlewareManager) RequireAuth() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			accessToken, err := c.Cookie("accessToken")
			log.Print(accessToken)
			if err != nil {
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"error":   "MISSING_ACCESS_TOKEN",
					"message": "Access token is required",
				})
			}

			claims, err := m.jwtService.ValidateToken(accessToken.Value, "access")
			if err != nil {
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"error":   "INVALID_ACCESS_TOKEN",
					"message": "Invalid or expired access token",
				})
			}

			c.Set("user_id", claims.UserID)
			c.Set("session_id", claims.SessionID)

			return next(c)
		}
	}
}

func (m *MiddlewareManager) TwoFactorTokenCheck() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			token, err := c.Cookie("2fa_token")
			userAgent := c.Request().UserAgent()
			userIP := c.RealIP()
			if err != nil {
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"error":   "MISSING_2FA_TOKEN",
					"message": "2FA token is required",
				})
			}

			claims, err := m.jwtService.Validate2FAToken(token.Value)
			if err != nil {
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"error":   "INVALID_2FA_TOKEN",
					"message": "Invalid or expired 2fa token",
				})
			}

			if claims.UserAgent != userAgent || claims.UserIP != userIP {
				return c.JSON(http.StatusForbidden, map[string]string{
					"error":   "DEVICE_MISMATCH",
					"message": "2FA code must be entered from the same device where it was requested",
				})
			}

			c.Set("user_id", claims.UserID)

			return next(c)
		}
	}
}
