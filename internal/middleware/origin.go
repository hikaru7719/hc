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

// OriginValidator creates a middleware that validates the Origin header
// to ensure requests are coming from the same origin
func OriginValidator(allowedPort int) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			log := logger.Get()
			
			// Skip origin check for non-API routes (static files)
			if !isAPIRoute(c.Request().URL.Path) {
				return next(c)
			}

			origin := c.Request().Header.Get("Origin")
			
			// If no Origin header, check Referer as fallback
			if origin == "" {
				referer := c.Request().Header.Get("Referer")
				if referer != "" {
					refURL, err := url.Parse(referer)
					if err == nil {
						origin = fmt.Sprintf("%s://%s", refURL.Scheme, refURL.Host)
					}
				}
			}

			// For local development, if no Origin or Referer, allow the request
			// This handles cases like direct API testing
			if origin == "" {
				log.Warn("No Origin or Referer header found", 
					slog.String("path", c.Request().URL.Path),
					slog.String("method", c.Request().Method))
				// In production, you might want to reject these requests
				// return c.JSON(http.StatusForbidden, models.NewErrorResponse("Missing Origin header"))
				return next(c)
			}

			// Validate the origin
			if !isAllowedOrigin(origin, allowedPort) {
				log.Error("Origin validation failed",
					slog.String("origin", origin),
					slog.String("path", c.Request().URL.Path),
					slog.String("method", c.Request().Method))
				return c.JSON(http.StatusForbidden, models.NewErrorResponse("Forbidden: Invalid origin"))
			}

			return next(c)
		}
	}
}

// isAPIRoute checks if the given path is an API route
func isAPIRoute(path string) bool {
	// Check if path starts with "/api/" or is exactly "/api"
	return path == "/api" || (len(path) > 4 && path[:5] == "/api/")
}

// isAllowedOrigin checks if the origin is allowed
func isAllowedOrigin(origin string, allowedPort int) bool {
	allowedOrigins := []string{
		fmt.Sprintf("http://localhost:%d", allowedPort),
		fmt.Sprintf("http://127.0.0.1:%d", allowedPort),
		fmt.Sprintf("http://[::1]:%d", allowedPort), // IPv6 localhost
	}

	for _, allowed := range allowedOrigins {
		if origin == allowed {
			return true
		}
	}

	return false
}