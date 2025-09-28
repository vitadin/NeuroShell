package services

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewHTTPRequestService(t *testing.T) {
	service := NewHTTPRequestService()
	assert.NotNil(t, service)
	assert.Equal(t, "http_request", service.Name())
	assert.False(t, service.initialized)
	assert.Equal(t, 30*time.Second, service.timeout)
	assert.Nil(t, service.client)
}

func TestHTTPRequestService_Name(t *testing.T) {
	service := NewHTTPRequestService()
	assert.Equal(t, "http_request", service.Name())
}

func TestHTTPRequestService_Initialize(t *testing.T) {
	service := NewHTTPRequestService()

	err := service.Initialize()
	assert.NoError(t, err)
	assert.True(t, service.initialized)
	assert.NotNil(t, service.client)
	assert.Equal(t, 30*time.Second, service.client.Timeout)
}

func TestHTTPRequestService_SetTimeout(t *testing.T) {
	service := NewHTTPRequestService()

	// Test setting timeout before initialization
	newTimeout := 60 * time.Second
	service.SetTimeout(newTimeout)
	assert.Equal(t, newTimeout, service.timeout)

	// Test setting timeout after initialization
	err := service.Initialize()
	assert.NoError(t, err)

	anotherTimeout := 45 * time.Second
	service.SetTimeout(anotherTimeout)
	assert.Equal(t, anotherTimeout, service.timeout)
	assert.Equal(t, anotherTimeout, service.client.Timeout)
}

func TestHTTPRequestService_SendRequest_NotInitialized(t *testing.T) {
	service := NewHTTPRequestService()

	request := HTTPRequest{
		Method: "GET",
		URL:    "http://example.com",
	}

	response, err := service.SendRequest(request)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "http request service not initialized")
	assert.Nil(t, response)
}

func TestHTTPRequestService_SendRequest_EmptyURL(t *testing.T) {
	service := NewHTTPRequestService()
	err := service.Initialize()
	require.NoError(t, err)

	request := HTTPRequest{
		Method: "GET",
		URL:    "",
	}

	response, err := service.SendRequest(request)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "URL is required")
	assert.Nil(t, response)
}

func TestHTTPRequestService_SendRequest_GET(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/test", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, "test-value", r.Header.Get("X-Test-Header"))

		w.Header().Set("X-Response-Header", "response-value")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"message": "success"}`))
	}))
	defer server.Close()

	service := NewHTTPRequestService()
	err := service.Initialize()
	require.NoError(t, err)

	request := HTTPRequest{
		Method: "GET",
		URL:    server.URL + "/test",
		Headers: map[string]string{
			"Content-Type":  "application/json",
			"X-Test-Header": "test-value",
		},
	}

	response, err := service.SendRequest(request)
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, http.StatusOK, response.StatusCode)
	assert.Equal(t, "200 OK", response.Status)
	assert.Equal(t, `{"message": "success"}`, response.Body)
	assert.Equal(t, "response-value", response.Headers["X-Response-Header"])
}

func TestHTTPRequestService_SendRequest_POST(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/api/data", r.URL.Path)

		body, err := io.ReadAll(r.Body)
		assert.NoError(t, err)
		assert.Equal(t, `{"key": "value"}`, string(body))

		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id": "123"}`))
	}))
	defer server.Close()

	service := NewHTTPRequestService()
	err := service.Initialize()
	require.NoError(t, err)

	request := HTTPRequest{
		Method: "POST",
		URL:    server.URL + "/api/data",
		Body:   `{"key": "value"}`,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
	}

	response, err := service.SendRequest(request)
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, http.StatusCreated, response.StatusCode)
	assert.Equal(t, `{"id": "123"}`, response.Body)
}

func TestHTTPRequestService_SendRequest_DefaultMethod(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	service := NewHTTPRequestService()
	err := service.Initialize()
	require.NoError(t, err)

	// Test with empty method - should default to GET
	request := HTTPRequest{
		Method: "",
		URL:    server.URL,
	}

	response, err := service.SendRequest(request)
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, http.StatusOK, response.StatusCode)
}

func TestHTTPRequestService_Get(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "test-value", r.Header.Get("X-Test"))
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("GET response"))
	}))
	defer server.Close()

	service := NewHTTPRequestService()
	err := service.Initialize()
	require.NoError(t, err)

	headers := map[string]string{"X-Test": "test-value"}
	response, err := service.Get(server.URL, headers)

	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, http.StatusOK, response.StatusCode)
	assert.Equal(t, "GET response", response.Body)
}

func TestHTTPRequestService_Post(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)

		body, err := io.ReadAll(r.Body)
		assert.NoError(t, err)
		assert.Equal(t, "test body", string(body))

		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte("POST response"))
	}))
	defer server.Close()

	service := NewHTTPRequestService()
	err := service.Initialize()
	require.NoError(t, err)

	headers := map[string]string{"Content-Type": "text/plain"}
	response, err := service.Post(server.URL, "test body", headers)

	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, http.StatusCreated, response.StatusCode)
	assert.Equal(t, "POST response", response.Body)
}

func TestHTTPRequestService_Put(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)

		body, err := io.ReadAll(r.Body)
		assert.NoError(t, err)
		assert.Equal(t, "updated data", string(body))

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("PUT response"))
	}))
	defer server.Close()

	service := NewHTTPRequestService()
	err := service.Initialize()
	require.NoError(t, err)

	response, err := service.Put(server.URL, "updated data", nil)

	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, http.StatusOK, response.StatusCode)
	assert.Equal(t, "PUT response", response.Body)
}

func TestHTTPRequestService_Delete(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "DELETE", r.Method)
		assert.Equal(t, "Bearer token123", r.Header.Get("Authorization"))

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	service := NewHTTPRequestService()
	err := service.Initialize()
	require.NoError(t, err)

	headers := map[string]string{"Authorization": "Bearer token123"}
	response, err := service.Delete(server.URL, headers)

	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, http.StatusNoContent, response.StatusCode)
}

func TestHTTPRequestService_GetServiceInfo(t *testing.T) {
	service := NewHTTPRequestService()

	// Test before initialization
	info := service.GetServiceInfo()
	assert.Equal(t, "http_request", info["name"])
	assert.Equal(t, false, info["initialized"])
	assert.Equal(t, "30s", info["timeout"])

	// Test after initialization
	err := service.Initialize()
	assert.NoError(t, err)

	info = service.GetServiceInfo()
	assert.Equal(t, "http_request", info["name"])
	assert.Equal(t, true, info["initialized"])
	assert.Equal(t, "30s", info["timeout"])
}

func TestHTTPRequestService_HTTPMethods(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		expectedMethod string
	}{
		{"GET method", "get", "GET"},
		{"POST method", "post", "POST"},
		{"PUT method", "put", "PUT"},
		{"DELETE method", "delete", "DELETE"},
		{"PATCH method", "patch", "PATCH"},
		{"HEAD method", "head", "HEAD"},
		{"OPTIONS method", "options", "OPTIONS"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, tt.expectedMethod, r.Method)
				w.WriteHeader(http.StatusOK)
			}))
			defer server.Close()

			service := NewHTTPRequestService()
			err := service.Initialize()
			require.NoError(t, err)

			request := HTTPRequest{
				Method: tt.method,
				URL:    server.URL,
			}

			response, err := service.SendRequest(request)
			assert.NoError(t, err)
			assert.NotNil(t, response)
			assert.Equal(t, http.StatusOK, response.StatusCode)
		})
	}
}

func TestHTTPRequestService_ErrorHandling(t *testing.T) {
	service := NewHTTPRequestService()
	err := service.Initialize()
	require.NoError(t, err)

	// Test invalid URL
	request := HTTPRequest{
		Method: "GET",
		URL:    "not-a-valid-url",
	}

	response, err := service.SendRequest(request)
	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "failed to execute HTTP request")
}
