package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestOriginValidator(t *testing.T) {
	e := echo.New()
	handler := func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	}

	tests := []struct {
		name           string
		path           string
		origin         string
		referer        string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "Valid origin for API route",
			path:           "/api/requests",
			origin:         "http://localhost:8080",
			expectedStatus: http.StatusOK,
			expectedBody:   "OK",
		},
		{
			name:           "Valid origin with 127.0.0.1",
			path:           "/api/requests",
			origin:         "http://127.0.0.1:8080",
			expectedStatus: http.StatusOK,
			expectedBody:   "OK",
		},
		{
			name:           "Valid origin with IPv6",
			path:           "/api/requests",
			origin:         "http://[::1]:8080",
			expectedStatus: http.StatusOK,
			expectedBody:   "OK",
		},
		{
			name:           "Invalid origin",
			path:           "/api/requests",
			origin:         "http://evil.com",
			expectedStatus: http.StatusForbidden,
			expectedBody:   `{"messages":["Forbidden: Invalid origin"]}`,
		},
		{
			name:           "No origin but valid referer",
			path:           "/api/requests",
			referer:        "http://localhost:8080/some-page",
			expectedStatus: http.StatusOK,
			expectedBody:   "OK",
		},
		{
			name:           "No origin or referer (allowed for now)",
			path:           "/api/requests",
			expectedStatus: http.StatusOK,
			expectedBody:   "OK",
		},
		{
			name:           "Static file route (no check)",
			path:           "/index.html",
			origin:         "http://evil.com",
			expectedStatus: http.StatusOK,
			expectedBody:   "OK",
		},
		{
			name:           "Root route (no check)",
			path:           "/",
			origin:         "http://evil.com",
			expectedStatus: http.StatusOK,
			expectedBody:   "OK",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			if tt.origin != "" {
				req.Header.Set("Origin", tt.origin)
			}
			if tt.referer != "" {
				req.Header.Set("Referer", tt.referer)
			}

			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			middleware := OriginValidator(8080)
			h := middleware(handler)
			h(c)

			assert.Equal(t, tt.expectedStatus, rec.Code)
			if tt.expectedBody != "" {
				assert.Contains(t, rec.Body.String(), tt.expectedBody)
			}
		})
	}
}

func TestIsAPIRoute(t *testing.T) {
	tests := []struct {
		path     string
		expected bool
	}{
		{"/api/requests", true},
		{"/api/folders", true},
		{"/api", true},
		{"/apifoo", false},
		{"/", false},
		{"/index.html", false},
		{"/static/css/main.css", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := isAPIRoute(tt.path)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsAllowedOrigin(t *testing.T) {
	port := 8080

	tests := []struct {
		origin   string
		expected bool
	}{
		{"http://localhost:8080", true},
		{"http://127.0.0.1:8080", true},
		{"http://[::1]:8080", true},
		{"https://localhost:8080", false},
		{"http://localhost:3000", false},
		{"http://evil.com", false},
		{"http://localhost", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.origin, func(t *testing.T) {
			result := isAllowedOrigin(tt.origin, port)
			assert.Equal(t, tt.expected, result)
		})
	}
}