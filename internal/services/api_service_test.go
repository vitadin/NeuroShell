package services

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	neuroshellcontext "neuroshell/internal/context"
	"neuroshell/internal/testutils"
)

// TestAPIService_Name tests the service name.
func TestAPIService_Name(t *testing.T) {
	service := NewAPIService()
	assert.Equal(t, "api", service.Name())
}

// TestAPIService_Initialize tests service initialization.
func TestAPIService_Initialize(t *testing.T) {
	service := NewAPIService()
	ctx := testutils.NewMockContext()

	// Test successful initialization
	err := service.Initialize(ctx)
	assert.NoError(t, err)
	assert.True(t, service.initialized)
	assert.NotNil(t, service.httpClient)
	assert.Equal(t, 30*time.Second, service.timeout)

	// Test already initialized
	err = service.Initialize(ctx)
	assert.NoError(t, err)
}

// TestAPIService_Initialize_WithConfig tests initialization with configuration.
func TestAPIService_Initialize_WithConfig(t *testing.T) {
	tests := []struct {
		name            string
		contextVars     map[string]string
		envVars         map[string]string
		expectedTimeout time.Duration
	}{
		{
			name:            "default timeout",
			contextVars:     map[string]string{},
			envVars:         map[string]string{},
			expectedTimeout: 30 * time.Second,
		},
		{
			name:            "context timeout",
			contextVars:     map[string]string{"@api_timeout": "60"},
			envVars:         map[string]string{},
			expectedTimeout: 60 * time.Second,
		},
		{
			name:            "env timeout",
			contextVars:     map[string]string{},
			envVars:         map[string]string{"API_TIMEOUT": "45"},
			expectedTimeout: 45 * time.Second,
		},
		{
			name:            "context overrides env",
			contextVars:     map[string]string{"@api_timeout": "120"},
			envVars:         map[string]string{"API_TIMEOUT": "90"},
			expectedTimeout: 120 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables
			for key, value := range tt.envVars {
				_ = os.Setenv(key, value)
			}
			defer func() {
				for key := range tt.envVars {
					_ = os.Unsetenv(key)
				}
			}()

			// Create context with variables
			ctx := testutils.NewMockContext()
			for key, value := range tt.contextVars {
				_ = ctx.SetVariable(key, value)
			}

			service := NewAPIService()
			err := service.Initialize(ctx)

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedTimeout, service.timeout)
		})
	}
}

// TestAPIService_GetTimeout tests timeout getter.
func TestAPIService_GetTimeout(t *testing.T) {
	service := NewAPIService()
	timeout := 45 * time.Second
	service.SetTimeout(timeout)
	assert.Equal(t, timeout, service.GetTimeout())
}

// TestAPIService_SetTimeout tests timeout setter.
func TestAPIService_SetTimeout(t *testing.T) {
	service := NewAPIService()
	ctx := testutils.NewMockContext()
	_ = service.Initialize(ctx)

	newTimeout := 60 * time.Second
	service.SetTimeout(newTimeout)

	assert.Equal(t, newTimeout, service.timeout)
	assert.Equal(t, newTimeout, service.httpClient.Timeout)
}

// TestAPIService_GetSupportedProviders tests supported providers.
func TestAPIService_GetSupportedProviders(t *testing.T) {
	service := NewAPIService()
	ctx := testutils.NewMockContext()
	_ = service.Initialize(ctx)

	providers := service.GetSupportedProviders()
	assert.Contains(t, providers, "openai")
	assert.Contains(t, providers, "anthropic")
	assert.Len(t, providers, 2)
}

// TestAPIService_GetAPIKey tests API key retrieval.
func TestAPIService_GetAPIKey(t *testing.T) {
	tests := []struct {
		name        string
		provider    string
		contextVars map[string]string
		envVars     map[string]string
		expectedKey string
		expectError bool
	}{
		{
			name:        "no key found",
			provider:    "openai",
			contextVars: map[string]string{},
			envVars:     map[string]string{},
			expectedKey: "",
			expectError: true,
		},
		{
			name:        "context key",
			provider:    "openai",
			contextVars: map[string]string{"@openai_api_key": "ctx-key-123"},
			envVars:     map[string]string{},
			expectedKey: "ctx-key-123",
			expectError: false,
		},
		{
			name:        "env key",
			provider:    "anthropic",
			contextVars: map[string]string{},
			envVars:     map[string]string{"ANTHROPIC_API_KEY": "env-key-456"},
			expectedKey: "env-key-456",
			expectError: false,
		},
		{
			name:        "context overrides env",
			provider:    "openai",
			contextVars: map[string]string{"@openai_api_key": "ctx-key-789"},
			envVars:     map[string]string{"OPENAI_API_KEY": "env-key-xyz"},
			expectedKey: "ctx-key-789",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables
			for key, value := range tt.envVars {
				_ = os.Setenv(key, value)
			}
			defer func() {
				for key := range tt.envVars {
					_ = os.Unsetenv(key)
				}
			}()

			// Create context with variables
			ctx := testutils.NewMockContext()
			for key, value := range tt.contextVars {
				_ = ctx.SetVariable(key, value)
			}

			// Set as global context for getAPIKey method
			oldCtx := neuroshellcontext.GetGlobalContext()
			neuroshellcontext.SetGlobalContext(ctx)
			defer neuroshellcontext.SetGlobalContext(oldCtx)

			service := NewAPIService()
			_ = service.Initialize(ctx)

			key, err := service.getAPIKey(tt.provider)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "no API key found")
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedKey, key)
			}
		})
	}
}

// TestAPIService_CheckConnectivity tests connectivity checking with mock server.
func TestAPIService_CheckConnectivity(t *testing.T) {
	tests := []struct {
		name         string
		provider     string
		statusCode   int
		expectError  bool
		errorMessage string
	}{
		{
			name:        "successful openai connection",
			provider:    "openai",
			statusCode:  200,
			expectError: false,
		},
		{
			name:        "successful anthropic connection",
			provider:    "anthropic",
			statusCode:  200,
			expectError: false,
		},
		{
			name:         "unauthorized",
			provider:     "openai",
			statusCode:   401,
			expectError:  true,
			errorMessage: "authentication failed: invalid API key",
		},
		{
			name:         "server error",
			provider:     "openai",
			statusCode:   500,
			expectError:  true,
			errorMessage: "API error: HTTP 500",
		},
		{
			name:         "unsupported provider",
			provider:     "unsupported",
			statusCode:   200,
			expectError:  true,
			errorMessage: "unsupported provider: unsupported",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify authentication headers
				switch tt.provider {
				case "openai":
					auth := r.Header.Get("Authorization")
					assert.Contains(t, auth, "Bearer")
				case "anthropic":
					apiKey := r.Header.Get("x-api-key")
					assert.NotEmpty(t, apiKey)
					version := r.Header.Get("anthropic-version")
					assert.Equal(t, "2023-06-01", version)
				}

				w.WriteHeader(tt.statusCode)
				_, _ = w.Write([]byte(`{"data": []}`))
			}))
			defer server.Close()

			// Create service and context
			service := NewAPIService()
			ctx := testutils.NewMockContext()

			// Set API key
			_ = ctx.SetVariable(fmt.Sprintf("@%s_api_key", tt.provider), "test-key-123")

			// Set global context
			oldCtx := neuroshellcontext.GetGlobalContext()
			neuroshellcontext.SetGlobalContext(ctx)
			defer neuroshellcontext.SetGlobalContext(oldCtx)

			// Initialize service
			_ = service.Initialize(ctx)

			// Override endpoint for supported providers
			if tt.provider == "openai" || tt.provider == "anthropic" {
				service.endpoints[tt.provider] = server.URL
			}

			// Test connectivity
			err := service.CheckConnectivity(tt.provider)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMessage != "" {
					assert.Contains(t, err.Error(), tt.errorMessage)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestAPIService_CheckConnectivity_NotInitialized tests error when service not initialized.
func TestAPIService_CheckConnectivity_NotInitialized(t *testing.T) {
	service := NewAPIService()
	err := service.CheckConnectivity("openai")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "api service not initialized")
}

// TestAPIService_CheckConnectivity_NoAPIKey tests error when no API key found.
func TestAPIService_CheckConnectivity_NoAPIKey(t *testing.T) {
	service := NewAPIService()
	ctx := testutils.NewMockContext()
	_ = service.Initialize(ctx)

	// Set global context without API key
	oldCtx := neuroshellcontext.GetGlobalContext()
	neuroshellcontext.SetGlobalContext(ctx)
	defer neuroshellcontext.SetGlobalContext(oldCtx)

	err := service.CheckConnectivity("openai")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "authentication failed")
}

// TestAPIService_CheckConnectivity_Timeout tests timeout handling.
func TestAPIService_CheckConnectivity_Timeout(t *testing.T) {
	// Create server that delays response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(100 * time.Millisecond) // Delay longer than timeout
		w.WriteHeader(200)
	}))
	defer server.Close()

	service := NewAPIService()
	ctx := testutils.NewMockContext()
	_ = ctx.SetVariable("@openai_api_key", "test-key")

	// Set global context
	oldCtx := neuroshellcontext.GetGlobalContext()
	neuroshellcontext.SetGlobalContext(ctx)
	defer neuroshellcontext.SetGlobalContext(oldCtx)

	_ = service.Initialize(ctx)
	service.SetTimeout(50 * time.Millisecond) // Short timeout
	service.endpoints["openai"] = server.URL

	err := service.CheckConnectivity("openai")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "connection timeout")
}

// TestAPIService_GetProviderStatus tests provider status checking.
func TestAPIService_GetProviderStatus(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"data": []}`))
	}))
	defer server.Close()

	service := NewAPIService()
	ctx := testutils.NewMockContext()
	_ = ctx.SetVariable("@openai_api_key", "test-key")

	// Set global context
	oldCtx := neuroshellcontext.GetGlobalContext()
	neuroshellcontext.SetGlobalContext(ctx)
	defer neuroshellcontext.SetGlobalContext(oldCtx)

	_ = service.Initialize(ctx)
	service.endpoints["openai"] = server.URL

	// Test successful status
	status, err := service.GetProviderStatus("openai")
	assert.NoError(t, err)
	assert.True(t, status)

	// Test failed status
	status, err = service.GetProviderStatus("nonexistent")
	assert.Error(t, err)
	assert.False(t, status)
}

// TestAPIService_TestAllProviders tests testing all providers.
func TestAPIService_TestAllProviders(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"data": []}`))
	}))
	defer server.Close()

	service := NewAPIService()
	ctx := testutils.NewMockContext()
	_ = ctx.SetVariable("@openai_api_key", "test-key")
	_ = ctx.SetVariable("@anthropic_api_key", "test-key")

	// Set global context
	oldCtx := neuroshellcontext.GetGlobalContext()
	neuroshellcontext.SetGlobalContext(ctx)
	defer neuroshellcontext.SetGlobalContext(oldCtx)

	_ = service.Initialize(ctx)
	service.endpoints["openai"] = server.URL
	service.endpoints["anthropic"] = server.URL

	results := service.TestAllProviders()

	assert.Len(t, results, 2)
	assert.Contains(t, results, "openai")
	assert.Contains(t, results, "anthropic")
	assert.NoError(t, results["openai"])
	assert.NoError(t, results["anthropic"])
}

// TestAPIService_GetEndpoints tests endpoint configuration.
func TestAPIService_GetEndpoints(t *testing.T) {
	tests := []struct {
		name              string
		contextVars       map[string]string
		envVars           map[string]string
		expectedOpenAI    string
		expectedAnthropic string
	}{
		{
			name:              "default endpoints",
			contextVars:       map[string]string{},
			envVars:           map[string]string{},
			expectedOpenAI:    "https://api.openai.com/v1/models",
			expectedAnthropic: "https://api.anthropic.com/v1/models",
		},
		{
			name: "context endpoints",
			contextVars: map[string]string{
				"@openai_endpoint":    "https://custom-openai.com/v1/models",
				"@anthropic_endpoint": "https://custom-anthropic.com/v1/models",
			},
			envVars:           map[string]string{},
			expectedOpenAI:    "https://custom-openai.com/v1/models",
			expectedAnthropic: "https://custom-anthropic.com/v1/models",
		},
		{
			name:        "env endpoints",
			contextVars: map[string]string{},
			envVars: map[string]string{
				"OPENAI_API_BASE_URL":    "https://env-openai.com",
				"ANTHROPIC_API_BASE_URL": "https://env-anthropic.com",
			},
			expectedOpenAI:    "https://env-openai.com/models",
			expectedAnthropic: "https://env-anthropic.com/models",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables
			for key, value := range tt.envVars {
				_ = os.Setenv(key, value)
			}
			defer func() {
				for key := range tt.envVars {
					_ = os.Unsetenv(key)
				}
			}()

			// Create context with variables
			ctx := testutils.NewMockContext()
			for key, value := range tt.contextVars {
				_ = ctx.SetVariable(key, value)
			}

			service := NewAPIService()
			endpoints := service.getEndpoints(ctx)

			assert.Equal(t, tt.expectedOpenAI, endpoints["openai"])
			assert.Equal(t, tt.expectedAnthropic, endpoints["anthropic"])
		})
	}
}
