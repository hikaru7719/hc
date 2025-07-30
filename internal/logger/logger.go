package logger

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/labstack/echo/v4"
)

var defaultLogger *slog.Logger

func init() {
	defaultLogger = New()
}

// New creates a new JSON logger with custom options
func New() *slog.Logger {
	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			// Customize timestamp format
			if a.Key == slog.TimeKey {
				if t, ok := a.Value.Any().(time.Time); ok {
					a.Value = slog.StringValue(t.Format(time.RFC3339))
				}
			}
			return a
		},
	}

	handler := slog.NewJSONHandler(os.Stdout, opts)
	return slog.New(handler)
}

// Get returns the default logger instance
func Get() *slog.Logger {
	return defaultLogger
}

// WithContext returns a logger with context values
func WithContext(ctx context.Context) *slog.Logger {
	return defaultLogger
}

// EchoMiddleware creates an Echo middleware for request logging
func EchoMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()
			
			// Log request
			defaultLogger.Info("request_started",
				slog.String("method", c.Request().Method),
				slog.String("path", c.Path()),
				slog.String("remote_addr", c.RealIP()),
				slog.String("user_agent", c.Request().UserAgent()),
			)
			
			// Process request
			err := next(c)
			
			// Calculate duration
			duration := time.Since(start)
			
			// Get status code
			status := c.Response().Status
			if err != nil {
				if he, ok := err.(*echo.HTTPError); ok {
					status = he.Code
				} else {
					status = 500
				}
			}
			
			// Log response
			logLevel := slog.LevelInfo
			if status >= 400 {
				logLevel = slog.LevelError
			}
			
			defaultLogger.Log(c.Request().Context(), logLevel, "request_completed",
				slog.String("method", c.Request().Method),
				slog.String("path", c.Path()),
				slog.Int("status", status),
				slog.Duration("duration", duration),
				slog.Int64("bytes_out", c.Response().Size),
			)
			
			if err != nil {
				defaultLogger.Error("request_error",
					slog.String("error", err.Error()),
				)
			}
			
			return err
		}
	}
}

