package neurotypes

import (
	"errors"
	"net/http"
	"os"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/stretchr/testify/assert"
)

// TestCoreInterfaces tests the core interface definitions and their contracts
func TestCoreInterfaces(t *testing.T) {
	t.Run("Context interface method signatures", func(t *testing.T) {
		// This test ensures the Context interface has the expected method signatures
		// by creating a mock implementation that satisfies the interface
		var ctx Context = &mockContext{}
		assert.NotNil(t, ctx)
	})

	t.Run("Service interface method signatures", func(t *testing.T) {
		var svc Service = &mockService{}
		assert.NotNil(t, svc)
		assert.Equal(t, "mock-service", svc.Name())
	})

	t.Run("Command interface method signatures", func(t *testing.T) {
		var cmd Command = &mockCommand{}
		assert.NotNil(t, cmd)
		assert.Equal(t, "mock-command", cmd.Name())
		assert.Equal(t, ParseModeKeyValue, cmd.ParseMode())
	})

	t.Run("ServiceRegistry interface method signatures", func(t *testing.T) {
		var reg ServiceRegistry = &mockServiceRegistry{}
		assert.NotNil(t, reg)
	})
}

// TestLLMTypes tests LLM-related types and interfaces
func TestLLMTypes(t *testing.T) {
	t.Run("StructuredLLMResponse creation and access", func(t *testing.T) {
		response := &StructuredLLMResponse{
			TextContent: "Hello world",
			ThinkingBlocks: []ThinkingBlock{
				{Content: "Thinking about response", Provider: "anthropic", Type: "thinking"},
			},
			Error:    &LLMError{Code: "rate_limit", Message: "Rate limit exceeded", Type: "rate_limit"},
			Metadata: map[string]interface{}{"tokens": 42},
		}

		assert.Equal(t, "Hello world", response.TextContent)
		assert.Len(t, response.ThinkingBlocks, 1)
		assert.Equal(t, "Thinking about response", response.ThinkingBlocks[0].Content)
		assert.Equal(t, "rate_limit", response.Error.Code)
		assert.Equal(t, 42, response.Metadata["tokens"])
	})

	t.Run("LLMError methods", func(t *testing.T) {
		err := &LLMError{
			Code:    "invalid_request",
			Message: "Invalid API key",
			Type:    "invalid_request",
		}

		assert.Equal(t, "invalid_request", err.Code)
		assert.Equal(t, "Invalid API key", err.Message)
		assert.Equal(t, "invalid_request", err.Type)
	})

	t.Run("ThinkingBlock methods", func(t *testing.T) {
		block := ThinkingBlock{
			Content:  "Reasoning step",
			Provider: "openai",
			Type:     "reasoning",
		}

		assert.Equal(t, "Reasoning step", block.Content)
		assert.Equal(t, "openai", block.Provider)
		assert.Equal(t, "reasoning", block.Type)
	})

	t.Run("LLMClient interface method signatures", func(t *testing.T) {
		var client LLMClient = &mockLLMClient{}
		assert.NotNil(t, client)
		assert.Equal(t, "mock-provider", client.GetProviderName())
		assert.True(t, client.IsConfigured())
	})
}

// TestCommandTypes tests command-related types
func TestCommandTypes(t *testing.T) {
	t.Run("ParseMode values", func(t *testing.T) {
		assert.Equal(t, ParseMode(0), ParseModeKeyValue)
		assert.Equal(t, ParseMode(1), ParseModeRaw)
		assert.Equal(t, ParseMode(2), ParseModeWithOptions)
	})

	t.Run("CommandArgs creation", func(t *testing.T) {
		args := CommandArgs{
			Options: map[string]string{"key": "value"},
			Message: "test message",
		}

		assert.Equal(t, "value", args.Options["key"])
		assert.Equal(t, "test message", args.Message)
	})

	t.Run("HelpInfo creation", func(t *testing.T) {
		help := HelpInfo{
			Description: "Test command",
			Usage:       "\\test[key=value] message",
			Options: []HelpOption{
				{Name: "key", Description: "Test option", Required: true},
			},
			Examples: []HelpExample{
				{Command: "\\test[key=value] hello", Description: "Basic usage"},
			},
		}

		assert.Equal(t, "Test command", help.Description)
		assert.Len(t, help.Options, 1)
		assert.Len(t, help.Examples, 1)
	})
}

// TestSessionTypes tests session-related types
func TestSessionTypes(t *testing.T) {
	t.Run("Message creation", func(t *testing.T) {
		msg := Message{
			Role:    "user",
			Content: "Hello",
		}

		assert.Equal(t, "user", msg.Role)
		assert.Equal(t, "Hello", msg.Content)
	})

	t.Run("SessionState creation", func(t *testing.T) {
		state := SessionState{
			Variables: map[string]string{"test": "value"},
		}

		assert.Equal(t, "value", state.Variables["test"])
	})

	t.Run("ChatSession creation", func(t *testing.T) {
		session := &ChatSession{
			ID:           "test-session",
			Name:         "Test Session",
			SystemPrompt: "You are a helpful assistant",
			Messages:     []Message{{Role: "user", Content: "Hello"}},
		}

		assert.Equal(t, "test-session", session.ID)
		assert.Equal(t, "Test Session", session.Name)
		assert.Len(t, session.Messages, 1)
	})
}

// TestModelTypes tests model-related types
func TestModelTypes(t *testing.T) {
	t.Run("ModelConfig creation", func(t *testing.T) {
		config := &ModelConfig{
			ID:          "test-model",
			Name:        "Test Model",
			Provider:    "openai",
			CatalogID:   "gpt-4",
			Description: "Test model configuration",
			Parameters:  map[string]interface{}{"temperature": 0.7},
		}

		assert.Equal(t, "test-model", config.ID)
		assert.Equal(t, "Test Model", config.Name)
		assert.Equal(t, 0.7, config.Parameters["temperature"].(float64))
	})

	t.Run("ModelCatalogEntry creation", func(t *testing.T) {
		entry := ModelCatalogEntry{
			ID:          "gpt-4",
			Name:        "GPT-4",
			Provider:    "openai",
			Description: "OpenAI GPT-4 model",
		}

		assert.Equal(t, "gpt-4", entry.ID)
		assert.Equal(t, "GPT-4", entry.Name)
	})

	t.Run("StandardModelParameters nil values by default", func(t *testing.T) {
		params := StandardModelParameters{}
		assert.Nil(t, params.Temperature)
		assert.Nil(t, params.TopP)
		assert.Nil(t, params.MaxTokens)
		assert.Nil(t, params.TopK)
		assert.Nil(t, params.PresencePenalty)
		assert.Nil(t, params.FrequencyPenalty)
	})

	t.Run("ModelValidationError methods", func(t *testing.T) {
		err := ModelValidationError{
			Field:   "temperature",
			Value:   "2.5",
			Message: "Temperature must be between 0.0 and 1.0",
		}

		assert.Equal(t, "Temperature must be between 0.0 and 1.0", err.Error())
		assert.Equal(t, "temperature", err.Field)
		assert.Equal(t, "2.5", err.Value)
		assert.Equal(t, "Temperature must be between 0.0 and 1.0", err.Message)
	})
}

// TestThemeTypes tests theme-related types
func TestThemeTypes(t *testing.T) {
	t.Run("ThemeConfig creation", func(t *testing.T) {
		config := &ThemeConfig{
			Name:   "test-theme",
			Styles: ThemeStyles{},
		}

		assert.Equal(t, "test-theme", config.Name)
		assert.NotNil(t, config.Styles)
	})

	t.Run("StyleConfig creation", func(t *testing.T) {
		bold := true
		style := StyleConfig{
			Foreground: "white",
			Background: "black",
			Bold:       &bold,
		}

		assert.Equal(t, "white", style.Foreground)
		assert.Equal(t, "black", style.Background)
		assert.True(t, *style.Bold)
	})

	t.Run("AdaptiveColor methods", func(t *testing.T) {
		color := AdaptiveColor{
			Light: "white",
			Dark:  "black",
		}

		assert.Equal(t, "white", color.Light)
		assert.Equal(t, "black", color.Dark)
	})

	t.Run("ThemeValidationError methods", func(t *testing.T) {
		err := ThemeValidationError{
			Field:   "foreground",
			Value:   "invalid-color",
			Message: "Invalid color value",
		}

		assert.Equal(t, "Invalid color value", err.Error())
		assert.Equal(t, "foreground", err.Field)
		assert.Equal(t, "invalid-color", err.Value)
		assert.Equal(t, "Invalid color value", err.Message)
	})
}

// TestProviderTypes tests provider-related types
func TestProviderTypes(t *testing.T) {
	t.Run("ModelProviderInfo creation", func(t *testing.T) {
		info := ModelProviderInfo{
			Name:                "OpenAI",
			SupportedModels:     []string{"gpt-4", "gpt-3.5-turbo"},
			SupportedParameters: []string{"temperature", "max_tokens"},
		}

		assert.Equal(t, "OpenAI", info.Name)
		assert.Len(t, info.SupportedModels, 2)
		assert.Len(t, info.SupportedParameters, 2)
	})
}

// TestStateMachineTypes tests state machine types
func TestStateMachineTypes(t *testing.T) {
	t.Run("StateMachineConfig creation", func(t *testing.T) {
		config := StateMachineConfig{
			EchoCommands:   true,
			MacroExpansion: false,
			RecursionLimit: 50,
		}

		assert.True(t, config.EchoCommands)
		assert.False(t, config.MacroExpansion)
		assert.Equal(t, 50, config.RecursionLimit)
	})

	t.Run("State methods", func(t *testing.T) {
		assert.Equal(t, "Received", StateReceived.String())
		assert.Equal(t, "Interpolating", StateInterpolating.String())
		assert.Equal(t, "Parsing", StateParsing.String())
		assert.Equal(t, "Resolving", StateResolving.String())
		assert.Equal(t, "TryResolving", StateTryResolving.String())
		assert.Equal(t, "Executing", StateExecuting.String())
		assert.Equal(t, "TryExecuting", StateTryExecuting.String())
		assert.Equal(t, "ScriptLoaded", StateScriptLoaded.String())
		assert.Equal(t, "ScriptExecuting", StateScriptExecuting.String())
		assert.Equal(t, "TryCompleted", StateTryCompleted.String())
		assert.Equal(t, "Completed", StateCompleted.String())
		assert.Equal(t, "Error", StateError.String())
		assert.Equal(t, "Unknown", State(999).String())
	})
}

// TestChangeLogTypes tests change log types
func TestChangeLogTypes(t *testing.T) {
	t.Run("ChangeLogEntry creation", func(t *testing.T) {
		entry := ChangeLogEntry{
			ID:          "CL001",
			Version:     "1.0.0",
			Date:        "2024-01-01",
			Type:        "feature",
			Title:       "Initial release",
			Description: "First version of NeuroShell",
		}

		assert.Equal(t, "CL001", entry.ID)
		assert.Equal(t, "1.0.0", entry.Version)
		assert.Equal(t, "feature", entry.Type)
	})

	t.Run("ChangeLogType methods", func(t *testing.T) {
		assert.Equal(t, "bugfix", ChangeLogTypeBugfix.String())
		assert.Equal(t, "feature", ChangeLogTypeFeature.String())
		assert.Equal(t, "enhancement", ChangeLogTypeEnhancement.String())
		assert.Equal(t, "testing", ChangeLogTypeTesting.String())
		assert.Equal(t, "refactor", ChangeLogTypeRefactor.String())
		assert.Equal(t, "docs", ChangeLogTypeDocs.String())
		assert.Equal(t, "chore", ChangeLogTypeChore.String())

		assert.True(t, ChangeLogTypeBugfix.IsValid())
		assert.True(t, ChangeLogTypeFeature.IsValid())
		assert.True(t, ChangeLogTypeEnhancement.IsValid())
		assert.True(t, ChangeLogTypeTesting.IsValid())
		assert.True(t, ChangeLogTypeRefactor.IsValid())
		assert.True(t, ChangeLogTypeDocs.IsValid())
		assert.True(t, ChangeLogTypeChore.IsValid())
		assert.False(t, ChangeLogType("invalid").IsValid())
	})

	t.Run("ChangeLogStats creation", func(t *testing.T) {
		stats := ChangeLogStats{
			TotalEntries:  10,
			TypeCounts:    map[string]int{"feature": 5, "bugfix": 3, "docs": 2},
			VersionCounts: map[string]int{"1.0.0": 3, "1.1.0": 7},
		}

		assert.Equal(t, 10, stats.TotalEntries)
		assert.Equal(t, 5, stats.TypeCounts["feature"])
		assert.Equal(t, 3, stats.VersionCounts["1.0.0"])
	})
}

// TestRenderingTypes tests rendering types
func TestRenderingTypes(t *testing.T) {
	t.Run("RenderConfig interface contract", func(t *testing.T) {
		// RenderConfig is an interface, so we test with a mock implementation
		var config RenderConfig = &mockRenderConfig{}
		assert.NotNil(t, config)

		// Test that the mock implements the interface methods
		assert.Equal(t, "mock-theme", config.GetTheme())
		assert.Equal(t, 80, config.GetMaxWidth())
		assert.True(t, config.ShowThinking())
	})
}

// TestCommandResolutionTypes tests command resolution types
func TestCommandResolutionTypes(t *testing.T) {
	t.Run("StateMachineResolvedCommand creation", func(t *testing.T) {
		result := StateMachineResolvedCommand{
			Name:          "echo",
			Type:          CommandTypeBuiltin,
			ScriptContent: "",
			ScriptPath:    "",
		}

		assert.Equal(t, "echo", result.Name)
		assert.Equal(t, CommandTypeBuiltin, result.Type)
		assert.Empty(t, result.ScriptContent)
		assert.Empty(t, result.ScriptPath)
	})

	t.Run("CommandType methods", func(t *testing.T) {
		assert.Equal(t, "builtin", CommandTypeBuiltin.String())
		assert.Equal(t, "stdlib", CommandTypeStdlib.String())
		assert.Equal(t, "user", CommandTypeUser.String())
		assert.Equal(t, "try", CommandTypeTry.String())
		assert.Equal(t, "unknown", CommandType(999).String())
	})
}

// Mock implementations for interface testing

type mockContext struct{}

func (m *mockContext) GetVariable(name string) (string, error) {
	if name == "test" {
		return "value", nil
	}
	return "", errors.New("variable not found")
}

func (m *mockContext) SetVariable(_ string, _ string) error {
	return nil
}

func (m *mockContext) GetMessageHistory(_ int) []Message {
	return []Message{}
}

func (m *mockContext) GetSessionState() SessionState {
	return SessionState{}
}

func (m *mockContext) SetTestMode(_ bool)                                 {}
func (m *mockContext) IsTestMode() bool                                   { return false }
func (m *mockContext) GetEnv(_ string) string                             { return "" }
func (m *mockContext) SetTestEnvOverride(_, _ string)                     {}
func (m *mockContext) ClearTestEnvOverride(_ string)                      {}
func (m *mockContext) ClearAllTestEnvOverrides()                          {}
func (m *mockContext) GetTestEnvOverrides() map[string]string             { return map[string]string{} }
func (m *mockContext) GetChatSessions() map[string]*ChatSession           { return map[string]*ChatSession{} }
func (m *mockContext) SetChatSessions(_ map[string]*ChatSession)          {}
func (m *mockContext) GetSessionNameToID() map[string]string              { return map[string]string{} }
func (m *mockContext) SetSessionNameToID(_ map[string]string)             {}
func (m *mockContext) GetActiveSessionID() string                         { return "" }
func (m *mockContext) SetActiveSessionID(_ string)                        {}
func (m *mockContext) GetActiveModelID() string                           { return "" }
func (m *mockContext) SetActiveModelID(_ string)                          {}
func (m *mockContext) GetModels() map[string]*ModelConfig                 { return map[string]*ModelConfig{} }
func (m *mockContext) SetModels(_ map[string]*ModelConfig)                {}
func (m *mockContext) GetModelNameToID() map[string]string                { return map[string]string{} }
func (m *mockContext) SetModelNameToID(_ map[string]string)               {}
func (m *mockContext) GetModelIDToName() map[string]string                { return map[string]string{} }
func (m *mockContext) SetModelIDToName(_ map[string]string)               {}
func (m *mockContext) ModelNameExists(_ string) bool                      { return false }
func (m *mockContext) ModelIDExists(_ string) bool                        { return false }
func (m *mockContext) GetLLMClient(_ string) (LLMClient, bool)            { return nil, false }
func (m *mockContext) SetLLMClient(_ string, _ LLMClient)                 {}
func (m *mockContext) GetAllLLMClients() map[string]LLMClient             { return map[string]LLMClient{} }
func (m *mockContext) GetLLMClientCount() int                             { return 0 }
func (m *mockContext) ClearLLMClients()                                   {}
func (m *mockContext) GetAllVariables() map[string]string                 { return map[string]string{} }
func (m *mockContext) SetVariableWithValidation(_ string, _ string) error { return nil }
func (m *mockContext) GetDefaultCommand() string                          { return "" }
func (m *mockContext) SetDefaultCommand(_ string)                         {}
func (m *mockContext) ReadFile(_ string) ([]byte, error)                  { return nil, nil }
func (m *mockContext) WriteFile(_ string, _ []byte, _ os.FileMode) error  { return nil }
func (m *mockContext) FileExists(_ string) bool                           { return false }
func (m *mockContext) GetUserConfigDir() (string, error)                  { return "", nil }
func (m *mockContext) GetWorkingDir() (string, error)                     { return "", nil }
func (m *mockContext) MkdirAll(_ string, _ os.FileMode) error             { return nil }
func (m *mockContext) GetConfigMap() map[string]string                    { return map[string]string{} }
func (m *mockContext) SetConfigMap(_ map[string]string)                   {}
func (m *mockContext) GetConfigValue(_ string) (string, bool)             { return "", false }
func (m *mockContext) SetConfigValue(_, _ string)                         {}
func (m *mockContext) LoadDefaults() error                                { return nil }
func (m *mockContext) LoadConfigDotEnv() error                            { return nil }
func (m *mockContext) LoadLocalDotEnv() error                             { return nil }
func (m *mockContext) LoadEnvironmentVariables(_ []string) error          { return nil }
func (m *mockContext) LoadEnvironmentVariablesWithPrefix(_ string) error  { return nil }
func (m *mockContext) LoadConfigDotEnvWithPrefix(_ string) error          { return nil }
func (m *mockContext) LoadLocalDotEnvWithPrefix(_ string) error           { return nil }
func (m *mockContext) GetSupportedProviders() []string                    { return []string{} }
func (m *mockContext) GetProviderEnvPrefixes() []string                   { return []string{} }
func (m *mockContext) IsValidProvider(_ string) bool                      { return false }
func (m *mockContext) SetCommandReadOnly(_ string, _ bool)                {}
func (m *mockContext) RemoveCommandReadOnlyOverride(_ string)             {}
func (m *mockContext) GetReadOnlyOverrides() map[string]bool              { return map[string]bool{} }
func (m *mockContext) IsCommandReadOnly(_ Command) bool                   { return false }

type mockService struct{}

func (m *mockService) Name() string      { return "mock-service" }
func (m *mockService) Initialize() error { return nil }

type mockCommand struct{}

func (m *mockCommand) Name() string         { return "mock-command" }
func (m *mockCommand) ParseMode() ParseMode { return ParseModeKeyValue }
func (m *mockCommand) Description() string  { return "Mock command for testing" }
func (m *mockCommand) Usage() string        { return "\\mock[key=value] message" }
func (m *mockCommand) HelpInfo() HelpInfo {
	return HelpInfo{
		Description: "Mock command",
		Usage:       "\\mock[key=value] message",
		Options:     []HelpOption{},
		Examples:    []HelpExample{},
	}
}
func (m *mockCommand) Execute(_ map[string]string, _ string) error { return nil }
func (m *mockCommand) IsReadOnly() bool                            { return false }

type mockServiceRegistry struct{}

func (m *mockServiceRegistry) GetService(_ string) (Service, error) {
	return &mockService{}, nil
}
func (m *mockServiceRegistry) RegisterService(_ Service) error { return nil }

type mockLLMClient struct{}

func (m *mockLLMClient) SendChatCompletion(_ *ChatSession, _ *ModelConfig) (string, error) {
	return "mock response", nil
}
func (m *mockLLMClient) SendStructuredCompletion(_ *ChatSession, _ *ModelConfig) *StructuredLLMResponse {
	return &StructuredLLMResponse{
		TextContent:    "mock response",
		ThinkingBlocks: []ThinkingBlock{},
		Error:          nil,
		Metadata:       map[string]interface{}{},
	}
}
func (m *mockLLMClient) GetProviderName() string               { return "mock-provider" }
func (m *mockLLMClient) IsConfigured() bool                    { return true }
func (m *mockLLMClient) SetDebugTransport(_ http.RoundTripper) {}

type mockRenderConfig struct{}

func (m *mockRenderConfig) GetStyle(_ string) lipgloss.Style {
	return lipgloss.NewStyle()
}

func (m *mockRenderConfig) GetTheme() string {
	return "mock-theme"
}

func (m *mockRenderConfig) IsCompactMode() bool {
	return false
}

func (m *mockRenderConfig) GetMaxWidth() int {
	return 80
}

func (m *mockRenderConfig) ShowThinking() bool {
	return true
}

func (m *mockRenderConfig) GetThinkingStyle() string {
	return "full"
}
