// Package services provides HTTP request operations for NeuroShell.
package services

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"neuroshell/internal/logger"
)

// HTTPRequestService provides HTTP/HTTPS request operations.
// This service is stateless and focuses on simple request/response operations.
type HTTPRequestService struct {
	initialized bool
	timeout     time.Duration
	client      *http.Client
}

// HTTPRequest represents an HTTP request configuration.
type HTTPRequest struct {
	Method  string            // HTTP method (GET, POST, PUT, DELETE, etc.)
	URL     string            // Request URL
	Headers map[string]string // HTTP headers
	Body    string            // Request body (for POST, PUT, etc.)
}

// HTTPResponse represents an HTTP response.
type HTTPResponse struct {
	StatusCode int               // HTTP status code
	Status     string            // HTTP status message
	Headers    map[string]string // Response headers
	Body       string            // Response body
}

// NewHTTPRequestService creates a new HTTPRequestService instance with default timeout of 30 seconds.
func NewHTTPRequestService() *HTTPRequestService {
	return &HTTPRequestService{
		initialized: false,
		timeout:     30 * time.Second,
	}
}

// Name returns the service name "http_request" for registration.
func (h *HTTPRequestService) Name() string {
	return "http_request"
}

// Initialize sets up the HTTPRequestService for operation.
func (h *HTTPRequestService) Initialize() error {
	h.client = &http.Client{
		Timeout: h.timeout,
	}
	h.initialized = true
	logger.Debug("HTTPRequestService initialized", "timeout", h.timeout.String())
	return nil
}

// SetTimeout configures the request timeout.
func (h *HTTPRequestService) SetTimeout(timeout time.Duration) {
	oldTimeout := h.timeout
	h.timeout = timeout
	if h.client != nil {
		h.client.Timeout = timeout
	}
	logger.Debug("HTTP request timeout updated", "old_timeout", oldTimeout.String(), "new_timeout", timeout.String())
}

// SendRequest sends an HTTP request and returns the response.
func (h *HTTPRequestService) SendRequest(request HTTPRequest) (*HTTPResponse, error) {
	if !h.initialized {
		logger.Error("HTTP request attempted on uninitialized service")
		return nil, fmt.Errorf("http request service not initialized")
	}

	// Validate required fields
	if request.URL == "" {
		logger.Error("HTTP request attempted with empty URL")
		return nil, fmt.Errorf("URL is required")
	}

	method := request.Method
	if method == "" {
		method = "GET"
	}
	method = strings.ToUpper(method)

	logger.Debug("Starting HTTP request",
		"method", method,
		"url", request.URL,
		"timeout", h.timeout.String(),
		"headers_count", len(request.Headers),
		"has_body", request.Body != "")

	// Create HTTP request
	var bodyReader io.Reader
	if request.Body != "" {
		bodyReader = strings.NewReader(request.Body)
	}

	ctx, cancel := context.WithTimeout(context.Background(), h.timeout)
	defer cancel()

	httpReq, err := http.NewRequestWithContext(ctx, method, request.URL, bodyReader)
	if err != nil {
		logger.Error("Failed to create HTTP request", "error", err, "method", method, "url", request.URL)
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set headers
	for key, value := range request.Headers {
		httpReq.Header.Set(key, value)
	}

	// Execute request
	resp, err := h.client.Do(httpReq)
	if err != nil {
		logger.Error("Failed to execute HTTP request", "error", err, "method", method, "url", request.URL)
		return nil, fmt.Errorf("failed to execute HTTP request: %w", err)
	}
	defer func() {
		_ = resp.Body.Close() // Ignore error on close
	}()

	logger.Debug("HTTP request completed",
		"method", method,
		"url", request.URL,
		"status_code", resp.StatusCode,
		"status", resp.Status)

	// Read response body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error("Failed to read response body", "error", err, "method", method, "url", request.URL, "status_code", resp.StatusCode)
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Convert response headers to map
	responseHeaders := make(map[string]string)
	for key, values := range resp.Header {
		if len(values) > 0 {
			responseHeaders[key] = values[0] // Take first value if multiple
		}
	}

	logger.Debug("HTTP response processed successfully",
		"method", method,
		"url", request.URL,
		"status_code", resp.StatusCode,
		"body_length", len(bodyBytes),
		"response_headers_count", len(responseHeaders))

	return &HTTPResponse{
		StatusCode: resp.StatusCode,
		Status:     resp.Status,
		Headers:    responseHeaders,
		Body:       string(bodyBytes),
	}, nil
}

// Get performs a simple GET request.
func (h *HTTPRequestService) Get(url string, headers map[string]string) (*HTTPResponse, error) {
	return h.SendRequest(HTTPRequest{
		Method:  "GET",
		URL:     url,
		Headers: headers,
	})
}

// Post performs a POST request with the given body.
func (h *HTTPRequestService) Post(url string, body string, headers map[string]string) (*HTTPResponse, error) {
	return h.SendRequest(HTTPRequest{
		Method:  "POST",
		URL:     url,
		Headers: headers,
		Body:    body,
	})
}

// Put performs a PUT request with the given body.
func (h *HTTPRequestService) Put(url string, body string, headers map[string]string) (*HTTPResponse, error) {
	return h.SendRequest(HTTPRequest{
		Method:  "PUT",
		URL:     url,
		Headers: headers,
		Body:    body,
	})
}

// Delete performs a DELETE request.
func (h *HTTPRequestService) Delete(url string, headers map[string]string) (*HTTPResponse, error) {
	return h.SendRequest(HTTPRequest{
		Method:  "DELETE",
		URL:     url,
		Headers: headers,
	})
}

// GetServiceInfo returns information about the HTTP request service.
func (h *HTTPRequestService) GetServiceInfo() map[string]interface{} {
	return map[string]interface{}{
		"name":        h.Name(),
		"initialized": h.initialized,
		"timeout":     h.timeout.String(),
	}
}
