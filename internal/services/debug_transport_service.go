package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"neuroshell/internal/logger"
)

// DebugTransportService provides HTTP request/response capture for all LLM clients.
// This service is always enabled and captures network traffic for debugging purposes.
type DebugTransportService struct {
	capturedData string
	initialized  bool
	mutex        sync.RWMutex
}

// NewDebugTransportService creates a new DebugTransportService instance.
func NewDebugTransportService() *DebugTransportService {
	return &DebugTransportService{
		initialized: false,
	}
}

// Name returns the service name "debug-transport" for registration.
func (d *DebugTransportService) Name() string {
	return "debug-transport"
}

// Initialize sets up the DebugTransportService for operation.
func (d *DebugTransportService) Initialize() error {
	logger.ServiceOperation("debug-transport", "initialize", "starting")
	d.mutex.Lock()
	defer d.mutex.Unlock()

	d.initialized = true
	d.capturedData = ""

	logger.ServiceOperation("debug-transport", "initialize", "completed")
	return nil
}

// CreateTransport creates a new debug-enabled HTTP transport.
// This transport always captures HTTP request/response data.
func (d *DebugTransportService) CreateTransport() http.RoundTripper {
	if !d.initialized {
		logger.Error("Debug transport service not initialized")
		return http.DefaultTransport
	}

	return &debugTransport{
		base:    http.DefaultTransport,
		service: d,
	}
}

// GetCapturedData returns the captured HTTP debug data as JSON string.
func (d *DebugTransportService) GetCapturedData() string {
	d.mutex.RLock()
	defer d.mutex.RUnlock()
	return d.capturedData
}

// ClearCapturedData clears the captured debug data.
func (d *DebugTransportService) ClearCapturedData() {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	d.capturedData = ""
}

// setCapturedData sets the captured debug data (thread-safe).
func (d *DebugTransportService) setCapturedData(data string) {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	d.capturedData = data
}

// debugTransport implements http.RoundTripper with request/response capture.
type debugTransport struct {
	base    http.RoundTripper
	service *DebugTransportService
}

// RoundTrip implements http.RoundTripper interface with debug capture.
func (dt *debugTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	startTime := time.Now()

	// Capture request data
	requestData, err := dt.captureRequest(req)
	if err != nil {
		logger.Error("Failed to capture request", "error", err)
		// Continue with request even if capture fails
	}

	// Make the actual HTTP request
	resp, err := dt.base.RoundTrip(req)
	endTime := time.Now()

	if err != nil {
		// Capture error information
		dt.captureError(requestData, err, startTime, endTime)
		return resp, err
	}

	// Capture response data
	responseData, captureErr := dt.captureResponse(resp)
	if captureErr != nil {
		logger.Error("Failed to capture response", "error", captureErr)
		// Continue with response even if capture fails
		responseData = map[string]interface{}{
			"error": "failed to capture response data",
		}
	}

	// Combine request and response data
	dt.storeDebugData(requestData, responseData, startTime, endTime)

	return resp, err
}

// captureRequest captures HTTP request data.
func (dt *debugTransport) captureRequest(req *http.Request) (map[string]interface{}, error) {
	requestData := map[string]interface{}{
		"method":  req.Method,
		"url":     req.URL.String(),
		"headers": dt.sanitizeHeaders(req.Header),
	}

	// Capture request body if present
	if req.Body != nil {
		bodyBytes, err := io.ReadAll(req.Body)
		if err != nil {
			return requestData, fmt.Errorf("failed to read request body: %w", err)
		}

		// Restore the request body for actual transmission
		req.Body = io.NopCloser(bytes.NewReader(bodyBytes))

		// Parse JSON body if possible
		if len(bodyBytes) > 0 {
			var jsonBody interface{}
			if err := json.Unmarshal(bodyBytes, &jsonBody); err == nil {
				requestData["body"] = jsonBody
			} else {
				requestData["body"] = string(bodyBytes)
			}
		}
	}

	return requestData, nil
}

// captureResponse captures HTTP response data.
func (dt *debugTransport) captureResponse(resp *http.Response) (map[string]interface{}, error) {
	responseData := map[string]interface{}{
		"status_code": resp.StatusCode,
		"status":      resp.Status,
		"headers":     dt.sanitizeHeaders(resp.Header),
	}

	// Capture response body if present
	if resp.Body != nil {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return responseData, fmt.Errorf("failed to read response body: %w", err)
		}

		// Restore the response body for client consumption
		resp.Body = io.NopCloser(bytes.NewReader(bodyBytes))

		// Parse JSON body if possible
		if len(bodyBytes) > 0 {
			var jsonBody interface{}
			if err := json.Unmarshal(bodyBytes, &jsonBody); err == nil {
				responseData["body"] = jsonBody
			} else {
				responseData["body"] = string(bodyBytes)
			}
		}
	}

	return responseData, nil
}

// captureError captures error information when request fails.
func (dt *debugTransport) captureError(requestData map[string]interface{}, err error, startTime, endTime time.Time) {
	debugData := map[string]interface{}{
		"http_request": requestData,
		"http_response": map[string]interface{}{
			"error": err.Error(),
		},
		"timing": map[string]interface{}{
			"request_time":  startTime.Format(time.RFC3339),
			"response_time": endTime.Format(time.RFC3339),
			"duration_ms":   endTime.Sub(startTime).Milliseconds(),
		},
	}

	jsonData, jsonErr := json.Marshal(debugData)
	if jsonErr != nil {
		logger.Error("Failed to marshal error debug data", "error", jsonErr)
		dt.service.setCapturedData(`{"error": "failed to marshal debug data"}`)
		return
	}

	dt.service.setCapturedData(string(jsonData))
}

// storeDebugData combines request/response data and stores it.
func (dt *debugTransport) storeDebugData(requestData, responseData map[string]interface{}, startTime, endTime time.Time) {
	debugData := map[string]interface{}{
		"http_request":  requestData,
		"http_response": responseData,
		"timing": map[string]interface{}{
			"request_time":  startTime.Format(time.RFC3339),
			"response_time": endTime.Format(time.RFC3339),
			"duration_ms":   endTime.Sub(startTime).Milliseconds(),
		},
	}

	jsonData, err := json.Marshal(debugData)
	if err != nil {
		logger.Error("Failed to marshal debug data", "error", err)
		dt.service.setCapturedData(`{"error": "failed to marshal debug data"}`)
		return
	}

	dt.service.setCapturedData(string(jsonData))
	logger.Debug("Debug data captured", "data_length", len(jsonData))
}

// sanitizeHeaders removes or masks sensitive headers.
func (dt *debugTransport) sanitizeHeaders(headers http.Header) map[string]interface{} {
	sanitized := make(map[string]interface{})

	for name, values := range headers {
		lowerName := strings.ToLower(name)

		// Mask sensitive headers
		if strings.Contains(lowerName, "authorization") ||
			strings.Contains(lowerName, "api-key") ||
			strings.Contains(lowerName, "token") {
			if len(values) > 0 && len(values[0]) > 10 {
				// Show first 10 characters and mask the rest
				masked := values[0][:10] + "***[MASKED]***"
				sanitized[name] = []string{masked}
			} else {
				sanitized[name] = []string{"***[MASKED]***"}
			}
		} else {
			// Copy non-sensitive headers as-is
			sanitized[name] = values
		}
	}

	return sanitized
}

// GetGlobalDebugTransportService returns the global debug transport service instance.
func GetGlobalDebugTransportService() (*DebugTransportService, error) {
	serviceInterface, err := GetGlobalRegistry().GetService("debug-transport")
	if err != nil {
		return nil, fmt.Errorf("debug transport service not registered: %w", err)
	}

	debugService, ok := serviceInterface.(*DebugTransportService)
	if !ok {
		return nil, fmt.Errorf("service is not a DebugTransportService")
	}

	return debugService, nil
}

func init() {
	// Register the debug transport service globally
	if err := GetGlobalRegistry().RegisterService(NewDebugTransportService()); err != nil {
		panic(fmt.Sprintf("failed to register debug transport service: %v", err))
	}
}
