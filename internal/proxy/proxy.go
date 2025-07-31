package proxy

import (
	"bytes"
	"errors"
	"github.com/hc/hc/internal/models"
	"io"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	httpClient *http.Client
}

func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *Client) ExecuteRequest(req *models.Request) (*models.Response, error) {
	httpReq, err := http.NewRequest(req.Method, req.URL, strings.NewReader(req.Body))
	if err != nil {
		return nil, err
	}
	for key, value := range req.Headers {
		httpReq.Header.Set(key, value)
	}
	if req.Body != "" && httpReq.Header.Get("Content-Type") == "" {
		httpReq.Header.Set("Content-Type", "application/json")
	}
	start := time.Now()
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	headers := make(map[string]string)
	for key, values := range resp.Header {
		headers[key] = strings.Join(values, ", ")
	}
	return &models.Response{
		StatusCode: resp.StatusCode,
		Headers:    headers,
		Body:       string(body),
		Duration:   time.Since(start).Milliseconds(),
	}, nil
}

type ProxyRequest struct {
	Method  string            `json:"method"`
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers"`
	Body    string            `json:"body"`
}

func (c *Client) ProxyRequest(proxyReq *ProxyRequest) (*models.Response, error) {
	return c.ExecuteRequest(&models.Request{
		Method:  proxyReq.Method,
		URL:     proxyReq.URL,
		Headers: proxyReq.Headers,
		Body:    proxyReq.Body,
	})
}

func ValidateURL(url string) error {
	if url == "" {
		return errors.New("URL is required")
	}
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		return errors.New("URL must start with http:// or https://")
	}
	return nil
}

func ValidateMethod(method string) error {
	method = strings.ToUpper(method)
	for _, valid := range []string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS"} {
		if method == valid {
			return nil
		}
	}
	return errors.New("invalid HTTP method: " + method)
}

func CopyHeaders(src http.Header) map[string]string {
	headers := make(map[string]string)
	for key, values := range src {
		headers[key] = strings.Join(values, ", ")
	}
	return headers
}

func SetHeaders(req *http.Request, headers map[string]string) {
	for key, value := range headers {
		req.Header.Set(key, value)
	}
}

func ReadBodyAsString(body io.ReadCloser) (string, error) {
	if body == nil {
		return "", nil
	}
	defer body.Close()
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, body); err != nil {
		return "", err
	}
	return buf.String(), nil
}
