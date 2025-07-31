package middleware

import (
	"fmt"
	"log/slog"
	"net/http"
	"net/url"

	"github.com/hc/hc/internal/logger"
	"github.com/hc/hc/internal/models"
	"github.com/labstack/echo/v4"
)

func OriginValidator(allowedPort int) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if !isAPIRoute(c.Request().URL.Path) {
				return next(c)
			}
			origin := c.Request().Header.Get("Origin")
			if origin == "" {
				referer := c.Request().Header.Get("Referer")
				if referer != "" {
					refURL, err := url.Parse(referer)
					if err == nil {
						origin = fmt.Sprintf("%s://%s", refURL.Scheme, refURL.Host)
					}
				}
			}
			if origin == "" {
				logger.Get().Warn("No Origin or Referer header found",
					slog.String("path", c.Request().URL.Path),
					slog.String("method", c.Request().Method))
				return next(c)
			}
			if !isAllowedOrigin(origin, allowedPort) {
				logger.Get().Error("Origin validation failed",
					slog.String("origin", origin),
					slog.String("path", c.Request().URL.Path),
					slog.String("method", c.Request().Method))
				return c.JSON(http.StatusForbidden, models.NewErrorResponse("Forbidden: Invalid origin"))
			}
			return next(c)
		}
	}
}

func isAPIRoute(path string) bool {
	return path == "/api" || (len(path) > 4 && path[:5] == "/api/")
}

func isAllowedOrigin(origin string, allowedPort int) bool {
	for _, allowed := range []string{
		fmt.Sprintf("http://localhost:%d", allowedPort),
		fmt.Sprintf("http://127.0.0.1:%d", allowedPort),
		fmt.Sprintf("http://[::1]:%d", allowedPort),
	} {
		if origin == allowed {
			return true
		}
	}
	return false
}
