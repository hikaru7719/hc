package proxy

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/hc/hc/internal/models"
)

func TestValidateURL(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "Valid HTTP URL",
			url:     "http://example.com",
			wantErr: false,
		},
		{
			name:    "Valid HTTPS URL",
			url:     "https://example.com",
			wantErr: false,
		},
		{
			name:    "Valid URL with path",
			url:     "https://example.com/api/v1/users",
			wantErr: false,
		},
		{
			name:    "Valid URL with query params",
			url:     "https://example.com/search?q=test&page=1",
			wantErr: false,
		},
		{
			name:    "Empty URL",
			url:     "",
			wantErr: true,
			errMsg:  "URL is required",
		},
		{
			name:    "Invalid URL format",
			url:     "not-a-url",
			wantErr: true,
			errMsg:  "URL must start with http:// or https://",
		},
		{
			name:    "URL without scheme",
			url:     "example.com",
			wantErr: true,
			errMsg:  "URL must start with http:// or https://",
		},
		{
			name:    "FTP URL",
			url:     "ftp://example.com",
			wantErr: true,
			errMsg:  "URL must start with http:// or https://",
		},
		{
			name:    "File URL",
			url:     "file:///etc/passwd",
			wantErr: true,
			errMsg:  "URL must start with http:// or https://",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateURL(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.errMsg != "" && err.Error() != tt.errMsg {
				t.Errorf("ValidateURL() error = %v, want %v", err.Error(), tt.errMsg)
			}
		})
	}
}

func TestValidateMethod(t *testing.T) {
	tests := []struct {
		name    string
		method  string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "GET method",
			method:  "GET",
			wantErr: false,
		},
		{
			name:    "POST method",
			method:  "POST",
			wantErr: false,
		},
		{
			name:    "PUT method",
			method:  "PUT",
			wantErr: false,
		},
		{
			name:    "DELETE method",
			method:  "DELETE",
			wantErr: false,
		},
		{
			name:    "PATCH method",
			method:  "PATCH",
			wantErr: false,
		},
		{
			name:    "HEAD method",
			method:  "HEAD",
			wantErr: false,
		},
		{
			name:    "OPTIONS method",
			method:  "OPTIONS",
			wantErr: false,
		},
		{
			name:    "Empty method",
			method:  "",
			wantErr: true,
			errMsg:  "invalid HTTP method: ",
		},
		{
			name:    "Invalid method",
			method:  "INVALID",
			wantErr: true,
			errMsg:  "invalid HTTP method: INVALID",
		},
		{
			name:    "Lowercase method (should be valid)",
			method:  "get",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateMethod(tt.method)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateMethod() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.errMsg != "" && err.Error() != tt.errMsg {
				t.Errorf("ValidateMethod() error = %v, want %v", err.Error(), tt.errMsg)
			}
		})
	}
}

func TestProxyRequest(t *testing.T) {
	// Create a test HTTP server that handles different request types
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Add a small delay to ensure duration is > 0
		time.Sleep(5 * time.Millisecond)

		// For POST requests, validate specific headers and body
		if r.Method == "POST" {
			// Check headers
			if r.Header.Get("X-Test-Header") != "test-value" {
				t.Errorf("Expected header X-Test-Header=test-value, got %s", r.Header.Get("X-Test-Header"))
			}

			// Check body
			body, _ := io.ReadAll(r.Body)
			if string(body) != `{"test":"data"}` {
				t.Errorf("Expected body {\"test\":\"data\"}, got %s", string(body))
			}
		}

		// Write response
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Response-Header", "response-value")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"result":"success"}`))
	}))
	defer testServer.Close()

	client := NewClient()

	tests := []struct {
		name    string
		request *ProxyRequest
		wantErr bool
		check   func(t *testing.T, resp *models.Response)
	}{
		{
			name: "Successful POST request",
			request: &ProxyRequest{
				Method: "POST",
				URL:    testServer.URL,
				Headers: map[string]string{
					"X-Test-Header": "test-value",
				},
				Body: `{"test":"data"}`,
			},
			wantErr: false,
			check: func(t *testing.T, resp *models.Response) {
				if resp.StatusCode != http.StatusOK {
					t.Errorf("Expected status 200, got %d", resp.StatusCode)
				}
				if resp.Headers["Content-Type"] != "application/json" {
					t.Errorf("Expected Content-Type header, got %v", resp.Headers)
				}
				if resp.Headers["X-Response-Header"] != "response-value" {
					t.Errorf("Expected X-Response-Header, got %v", resp.Headers)
				}
				if resp.Body != `{"result":"success"}` {
					t.Errorf("Expected body {\"result\":\"success\"}, got %s", resp.Body)
				}
				if resp.Duration <= 0 {
					t.Error("Expected positive duration")
				}
			},
		},
		{
			name: "GET request without body",
			request: &ProxyRequest{
				Method: "GET",
				URL:    testServer.URL,
			},
			wantErr: false,
			check: func(t *testing.T, resp *models.Response) {
				if resp.StatusCode != http.StatusOK {
					t.Errorf("Expected status 200, got %d", resp.StatusCode)
				}
			},
		},
		{
			name: "Request to invalid URL",
			request: &ProxyRequest{
				Method: "GET",
				URL:    "http://invalid-host-that-does-not-exist.test",
			},
			wantErr: true,
		},
		{
			name: "Request with timeout",
			request: &ProxyRequest{
				Method: "GET",
				URL:    testServer.URL + "/slow",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := client.ProxyRequest(tt.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("ProxyRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.check != nil {
				tt.check(t, resp)
			}
		})
	}
}

func TestProxyRequestWithLargeResponse(t *testing.T) {
	// Create a test server that returns a large response
	largeBody := strings.Repeat("x", 1024*1024) // 1MB
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(largeBody))
	}))
	defer testServer.Close()

	client := NewClient()
	resp, err := client.ProxyRequest(&ProxyRequest{
		Method: "GET",
		URL:    testServer.URL,
	})

	if err != nil {
		t.Fatalf("ProxyRequest() error = %v", err)
	}

	if len(resp.Body) != len(largeBody) {
		t.Errorf("Expected response body length %d, got %d", len(largeBody), len(resp.Body))
	}
}

func TestProxyRequestWithHeaders(t *testing.T) {
	// Test that headers are properly forwarded
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Echo back the headers as JSON
		headers := make(map[string]string)
		for key, values := range r.Header {
			if strings.HasPrefix(key, "X-") {
				headers[key] = values[0]
			}
		}
		json.NewEncoder(w).Encode(headers)
	}))
	defer testServer.Close()

	client := NewClient()
	resp, err := client.ProxyRequest(&ProxyRequest{
		Method: "GET",
		URL:    testServer.URL,
		Headers: map[string]string{
			"X-Custom-1": "value1",
			"X-Custom-2": "value2",
		},
	})

	if err != nil {
		t.Fatalf("ProxyRequest() error = %v", err)
	}

	var receivedHeaders map[string]string
	if err := json.Unmarshal([]byte(resp.Body), &receivedHeaders); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if receivedHeaders["X-Custom-1"] != "value1" {
		t.Errorf("Expected X-Custom-1=value1, got %s", receivedHeaders["X-Custom-1"])
	}
	if receivedHeaders["X-Custom-2"] != "value2" {
		t.Errorf("Expected X-Custom-2=value2, got %s", receivedHeaders["X-Custom-2"])
	}
}

func TestProxyRequestWithRedirect(t *testing.T) {
	// Test that redirects are followed
	redirectCount := 0
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			redirectCount++
			http.Redirect(w, r, "/redirected", http.StatusFound)
			return
		}
		if r.URL.Path == "/redirected" {
			w.Write([]byte("redirected successfully"))
			return
		}
	}))
	defer testServer.Close()

	client := NewClient()
	resp, err := client.ProxyRequest(&ProxyRequest{
		Method: "GET",
		URL:    testServer.URL,
	})

	if err != nil {
		t.Fatalf("ProxyRequest() error = %v", err)
	}

	if resp.Body != "redirected successfully" {
		t.Errorf("Expected 'redirected successfully', got %s", resp.Body)
	}

	if redirectCount != 1 {
		t.Errorf("Expected 1 redirect, got %d", redirectCount)
	}
}

func TestNewClient(t *testing.T) {
	client := NewClient()

	if client == nil {
		t.Fatal("NewClient() returned nil")
	}

	if client.httpClient == nil {
		t.Fatal("NewClient() httpClient is nil")
	}

	// Check timeout is set
	if client.httpClient.Timeout != 30*time.Second {
		t.Errorf("Expected timeout 30s, got %v", client.httpClient.Timeout)
	}
}

func TestProxyRequestDuration(t *testing.T) {
	// Test that duration is measured correctly
	delay := 100 * time.Millisecond
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(delay)
		w.Write([]byte("delayed response"))
	}))
	defer testServer.Close()

	client := NewClient()
	resp, err := client.ProxyRequest(&ProxyRequest{
		Method: "GET",
		URL:    testServer.URL,
	})

	if err != nil {
		t.Fatalf("ProxyRequest() error = %v", err)
	}

	// Duration should be at least the delay
	if resp.Duration < int64(delay.Milliseconds()) {
		t.Errorf("Expected duration >= %dms, got %dms", delay.Milliseconds(), resp.Duration)
	}

	// But not too much more (allowing for some overhead)
	maxDuration := int64(delay.Milliseconds() + 100)
	if resp.Duration > maxDuration {
		t.Errorf("Expected duration <= %dms, got %dms", maxDuration, resp.Duration)
	}
}
