package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
)

func TestNew(t *testing.T) {
	logger := New()
	if logger == nil {
		t.Fatal("New() returned nil")
	}

	// Test that it's a valid slog.Logger
	if _, ok := logger.Handler().(*slog.JSONHandler); !ok {
		t.Error("Logger should use JSON handler")
	}
}

func TestGet(t *testing.T) {
	// Test that Get returns the same logger instance
	logger1 := Get()
	logger2 := Get()

	if logger1 != logger2 {
		t.Error("Get() should return the same logger instance")
	}

	if logger1 == nil {
		t.Fatal("Get() returned nil")
	}
}

func TestLoggerOutput(t *testing.T) {
	// Capture log output
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelInfo,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				if t, ok := a.Value.Any().(time.Time); ok {
					a.Value = slog.StringValue(t.Format(time.RFC3339))
				}
			}
			return a
		},
	}))

	// Log a test message
	logger.Info("test message", slog.String("key", "value"))

	// Parse the output
	var logEntry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
		t.Fatalf("Failed to parse log output: %v", err)
	}

	// Check fields
	if logEntry["level"] != "INFO" {
		t.Errorf("Expected level INFO, got %v", logEntry["level"])
	}

	if logEntry["msg"] != "test message" {
		t.Errorf("Expected msg 'test message', got %v", logEntry["msg"])
	}

	if logEntry["key"] != "value" {
		t.Errorf("Expected key 'value', got %v", logEntry["key"])
	}

	// Check time format
	if timeStr, ok := logEntry["time"].(string); ok {
		if _, err := time.Parse(time.RFC3339, timeStr); err != nil {
			t.Errorf("Time not in RFC3339 format: %v", timeStr)
		}
	} else {
		t.Error("Time field missing or not a string")
	}
}

func TestWithContext(t *testing.T) {
	ctx := context.Background()
	ctxLogger := WithContext(ctx)

	if ctxLogger == nil {
		t.Fatal("WithContext returned nil")
	}

	// Should return a logger (for now it returns the global logger)
	if ctxLogger != Get() {
		t.Error("WithContext should return the global logger")
	}
}

func TestEchoMiddleware(t *testing.T) {
	// Create Echo instance
	e := echo.New()

	// Add our middleware
	middleware := EchoMiddleware()
	e.Use(middleware)

	// Create a test handler
	e.GET("/test", func(c echo.Context) error {
		return c.String(http.StatusOK, "test response")
	})

	// Create test request
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	// Capture logs
	var buf bytes.Buffer
	oldLogger := defaultLogger
	defaultLogger = slog.New(slog.NewJSONHandler(&buf, nil))
	defer func() {
		defaultLogger = oldLogger
	}()

	// Execute request
	e.ServeHTTP(rec, req)

	// Check response
	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}

	// Check that logs were written
	logs := buf.String()
	if !strings.Contains(logs, "request_started") {
		t.Error("Expected 'request_started' log entry")
	}
	if !strings.Contains(logs, "request_completed") {
		t.Error("Expected 'request_completed' log entry")
	}

	// Parse log entries
	for _, line := range strings.Split(strings.TrimSpace(logs), "\n") {
		if line == "" {
			continue
		}

		var logEntry map[string]interface{}
		if err := json.Unmarshal([]byte(line), &logEntry); err != nil {
			t.Errorf("Failed to parse log line: %v", err)
			continue
		}

		// Check for expected fields in request completed log
		if logEntry["msg"] == "request_completed" {
			if logEntry["method"] != "GET" {
				t.Errorf("Expected method GET, got %v", logEntry["method"])
			}
			if logEntry["path"] != "/test" {
				t.Errorf("Expected path /test, got %v", logEntry["path"])
			}
			if logEntry["status"] != float64(200) {
				t.Errorf("Expected status 200, got %v", logEntry["status"])
			}
			if _, ok := logEntry["duration"]; !ok {
				t.Error("Expected duration field")
			}
		}
	}
}

func TestEchoMiddlewareError(t *testing.T) {
	// Create Echo instance
	e := echo.New()

	// Add our middleware
	e.Use(EchoMiddleware())

	// Create a handler that returns an error
	e.GET("/error", func(c echo.Context) error {
		return echo.NewHTTPError(http.StatusInternalServerError, "test error")
	})

	// Create test request
	req := httptest.NewRequest(http.MethodGet, "/error", nil)
	rec := httptest.NewRecorder()

	// Capture logs
	var buf bytes.Buffer
	oldLogger := defaultLogger
	defaultLogger = slog.New(slog.NewJSONHandler(&buf, nil))
	defer func() {
		defaultLogger = oldLogger
	}()

	// Execute request
	e.ServeHTTP(rec, req)

	// Check response
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", rec.Code)
	}

	// Check logs
	logs := buf.String()
	if !strings.Contains(logs, "request_completed") {
		t.Error("Expected 'request_completed' log entry")
	}
	if !strings.Contains(logs, "request_error") {
		t.Error("Expected 'request_error' log entry")
	}

	// Parse and check error was logged
	foundError := false
	for _, line := range strings.Split(strings.TrimSpace(logs), "\n") {
		if line == "" {
			continue
		}

		var logEntry map[string]interface{}
		if err := json.Unmarshal([]byte(line), &logEntry); err != nil {
			continue
		}

		if logEntry["msg"] == "request_error" {
			if _, ok := logEntry["error"]; ok {
				foundError = true
			}
		}
	}

	if !foundError {
		t.Error("Did not find error log entry")
	}
}
