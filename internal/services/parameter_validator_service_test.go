package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"neuroshell/internal/stringprocessing"
	"neuroshell/pkg/neurotypes"
)

func TestParameterValidatorService_Name(t *testing.T) {
	service := NewParameterValidatorService()
	assert.Equal(t, "parameter_validator", service.Name())
}

func TestParameterValidatorService_Initialize(t *testing.T) {
	service := NewParameterValidatorService()
	err := service.Initialize()
	assert.NoError(t, err)
}

func TestParameterValidatorService_ValidateParameters_StringType(t *testing.T) {
	service := NewParameterValidatorService()
	require.NoError(t, service.Initialize())

	tests := []struct {
		name         string
		paramDef     neurotypes.ParameterDefinition
		inputValue   string
		expectError  bool
		expectedType interface{}
	}{
		{
			name: "valid string parameter",
			paramDef: neurotypes.ParameterDefinition{
				Name:        "model_name",
				Type:        "string",
				Required:    true,
				Description: "Model name",
			},
			inputValue:   "gpt-4o",
			expectError:  false,
			expectedType: "gpt-4o",
		},
		{
			name: "string with pattern constraint - valid",
			paramDef: neurotypes.ParameterDefinition{
				Name:        "api_key",
				Type:        "string",
				Required:    true,
				Description: "API key",
				Constraints: &neurotypes.ParameterConstraints{
					Pattern: stringprocessing.StringPtr("^sk-"),
				},
			},
			inputValue:   "sk-test123",
			expectError:  false,
			expectedType: "sk-test123",
		},
		{
			name: "string with pattern constraint - invalid",
			paramDef: neurotypes.ParameterDefinition{
				Name:        "api_key",
				Type:        "string",
				Required:    true,
				Description: "API key",
				Constraints: &neurotypes.ParameterConstraints{
					Pattern: stringprocessing.StringPtr("^sk-"),
				},
			},
			inputValue:  "invalid-key",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := map[string]string{tt.paramDef.Name: tt.inputValue}
			paramDefs := []neurotypes.ParameterDefinition{tt.paramDef}

			result, err := service.ValidateParameters(args, paramDefs)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedType, result[tt.paramDef.Name])
			}
		})
	}
}

func TestParameterValidatorService_ValidateParameters_IntType(t *testing.T) {
	service := NewParameterValidatorService()
	require.NoError(t, service.Initialize())

	tests := []struct {
		name         string
		paramDef     neurotypes.ParameterDefinition
		inputValue   string
		expectError  bool
		expectedType interface{}
	}{
		{
			name: "valid int parameter",
			paramDef: neurotypes.ParameterDefinition{
				Name:        "max_tokens",
				Type:        "int",
				Required:    false,
				Description: "Maximum tokens",
			},
			inputValue:   "1000",
			expectError:  false,
			expectedType: 1000,
		},
		{
			name: "int with min constraint - valid",
			paramDef: neurotypes.ParameterDefinition{
				Name:        "max_tokens",
				Type:        "int",
				Required:    false,
				Description: "Maximum tokens",
				Constraints: &neurotypes.ParameterConstraints{
					Min: stringprocessing.Float64Ptr(1),
				},
			},
			inputValue:   "500",
			expectError:  false,
			expectedType: 500,
		},
		{
			name: "int with min constraint - invalid",
			paramDef: neurotypes.ParameterDefinition{
				Name:        "max_tokens",
				Type:        "int",
				Required:    false,
				Description: "Maximum tokens",
				Constraints: &neurotypes.ParameterConstraints{
					Min: stringprocessing.Float64Ptr(1),
				},
			},
			inputValue:  "0",
			expectError: true,
		},
		{
			name: "int with max constraint - valid",
			paramDef: neurotypes.ParameterDefinition{
				Name:        "max_tokens",
				Type:        "int",
				Required:    false,
				Description: "Maximum tokens",
				Constraints: &neurotypes.ParameterConstraints{
					Max: stringprocessing.Float64Ptr(16384),
				},
			},
			inputValue:   "1000",
			expectError:  false,
			expectedType: 1000,
		},
		{
			name: "int with max constraint - invalid",
			paramDef: neurotypes.ParameterDefinition{
				Name:        "max_tokens",
				Type:        "int",
				Required:    false,
				Description: "Maximum tokens",
				Constraints: &neurotypes.ParameterConstraints{
					Max: stringprocessing.Float64Ptr(16384),
				},
			},
			inputValue:  "20000",
			expectError: true,
		},
		{
			name: "int with range constraint - valid",
			paramDef: neurotypes.ParameterDefinition{
				Name:        "max_tokens",
				Type:        "int",
				Required:    false,
				Description: "Maximum tokens",
				Constraints: &neurotypes.ParameterConstraints{
					Min: stringprocessing.Float64Ptr(1),
					Max: stringprocessing.Float64Ptr(16384),
				},
			},
			inputValue:   "8192",
			expectError:  false,
			expectedType: 8192,
		},
		{
			name: "invalid int format",
			paramDef: neurotypes.ParameterDefinition{
				Name:        "max_tokens",
				Type:        "int",
				Required:    false,
				Description: "Maximum tokens",
			},
			inputValue:  "not-a-number",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := map[string]string{tt.paramDef.Name: tt.inputValue}
			paramDefs := []neurotypes.ParameterDefinition{tt.paramDef}

			result, err := service.ValidateParameters(args, paramDefs)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedType, result[tt.paramDef.Name])
			}
		})
	}
}

func TestParameterValidatorService_ValidateParameters_FloatType(t *testing.T) {
	service := NewParameterValidatorService()
	require.NoError(t, service.Initialize())

	tests := []struct {
		name         string
		paramDef     neurotypes.ParameterDefinition
		inputValue   string
		expectError  bool
		expectedType interface{}
	}{
		{
			name: "valid float parameter",
			paramDef: neurotypes.ParameterDefinition{
				Name:        "temperature",
				Type:        "float",
				Required:    false,
				Description: "Temperature",
			},
			inputValue:   "0.7",
			expectError:  false,
			expectedType: 0.7,
		},
		{
			name: "float as integer string",
			paramDef: neurotypes.ParameterDefinition{
				Name:        "temperature",
				Type:        "float",
				Required:    false,
				Description: "Temperature",
			},
			inputValue:   "1",
			expectError:  false,
			expectedType: 1.0,
		},
		{
			name: "float with min constraint - valid",
			paramDef: neurotypes.ParameterDefinition{
				Name:        "temperature",
				Type:        "float",
				Required:    false,
				Description: "Temperature",
				Constraints: &neurotypes.ParameterConstraints{
					Min: stringprocessing.Float64Ptr(0.0),
				},
			},
			inputValue:   "0.5",
			expectError:  false,
			expectedType: 0.5,
		},
		{
			name: "float with min constraint - invalid",
			paramDef: neurotypes.ParameterDefinition{
				Name:        "temperature",
				Type:        "float",
				Required:    false,
				Description: "Temperature",
				Constraints: &neurotypes.ParameterConstraints{
					Min: stringprocessing.Float64Ptr(0.0),
				},
			},
			inputValue:  "-0.1",
			expectError: true,
		},
		{
			name: "float with max constraint - valid",
			paramDef: neurotypes.ParameterDefinition{
				Name:        "temperature",
				Type:        "float",
				Required:    false,
				Description: "Temperature",
				Constraints: &neurotypes.ParameterConstraints{
					Max: stringprocessing.Float64Ptr(2.0),
				},
			},
			inputValue:   "1.5",
			expectError:  false,
			expectedType: 1.5,
		},
		{
			name: "float with max constraint - invalid",
			paramDef: neurotypes.ParameterDefinition{
				Name:        "temperature",
				Type:        "float",
				Required:    false,
				Description: "Temperature",
				Constraints: &neurotypes.ParameterConstraints{
					Max: stringprocessing.Float64Ptr(2.0),
				},
			},
			inputValue:  "2.5",
			expectError: true,
		},
		{
			name: "invalid float format",
			paramDef: neurotypes.ParameterDefinition{
				Name:        "temperature",
				Type:        "float",
				Required:    false,
				Description: "Temperature",
			},
			inputValue:  "not-a-float",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := map[string]string{tt.paramDef.Name: tt.inputValue}
			paramDefs := []neurotypes.ParameterDefinition{tt.paramDef}

			result, err := service.ValidateParameters(args, paramDefs)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedType, result[tt.paramDef.Name])
			}
		})
	}
}

func TestParameterValidatorService_ValidateParameters_BoolType(t *testing.T) {
	service := NewParameterValidatorService()
	require.NoError(t, service.Initialize())

	tests := []struct {
		name         string
		paramDef     neurotypes.ParameterDefinition
		inputValue   string
		expectError  bool
		expectedType interface{}
	}{
		{
			name: "bool true",
			paramDef: neurotypes.ParameterDefinition{
				Name:        "stream",
				Type:        "bool",
				Required:    false,
				Description: "Enable streaming",
			},
			inputValue:   "true",
			expectError:  false,
			expectedType: true,
		},
		{
			name: "bool false",
			paramDef: neurotypes.ParameterDefinition{
				Name:        "stream",
				Type:        "bool",
				Required:    false,
				Description: "Enable streaming",
			},
			inputValue:   "false",
			expectError:  false,
			expectedType: false,
		},
		{
			name: "bool 1",
			paramDef: neurotypes.ParameterDefinition{
				Name:        "stream",
				Type:        "bool",
				Required:    false,
				Description: "Enable streaming",
			},
			inputValue:   "1",
			expectError:  false,
			expectedType: true,
		},
		{
			name: "bool 0",
			paramDef: neurotypes.ParameterDefinition{
				Name:        "stream",
				Type:        "bool",
				Required:    false,
				Description: "Enable streaming",
			},
			inputValue:   "0",
			expectError:  false,
			expectedType: false,
		},
		{
			name: "invalid bool",
			paramDef: neurotypes.ParameterDefinition{
				Name:        "stream",
				Type:        "bool",
				Required:    false,
				Description: "Enable streaming",
			},
			inputValue:  "maybe",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := map[string]string{tt.paramDef.Name: tt.inputValue}
			paramDefs := []neurotypes.ParameterDefinition{tt.paramDef}

			result, err := service.ValidateParameters(args, paramDefs)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedType, result[tt.paramDef.Name])
			}
		})
	}
}

func TestParameterValidatorService_ValidateParameters_EnumType(t *testing.T) {
	service := NewParameterValidatorService()
	require.NoError(t, service.Initialize())

	tests := []struct {
		name         string
		paramDef     neurotypes.ParameterDefinition
		inputValue   string
		expectError  bool
		expectedType interface{}
	}{
		{
			name: "valid enum value",
			paramDef: neurotypes.ParameterDefinition{
				Name:        "reasoning_effort",
				Type:        "enum",
				Required:    false,
				Description: "Reasoning effort level",
				Constraints: &neurotypes.ParameterConstraints{
					EnumValues: []string{"low", "medium", "high"},
				},
			},
			inputValue:   "medium",
			expectError:  false,
			expectedType: "medium",
		},
		{
			name: "invalid enum value",
			paramDef: neurotypes.ParameterDefinition{
				Name:        "reasoning_effort",
				Type:        "enum",
				Required:    false,
				Description: "Reasoning effort level",
				Constraints: &neurotypes.ParameterConstraints{
					EnumValues: []string{"low", "medium", "high"},
				},
			},
			inputValue:  "extreme",
			expectError: true,
		},
		{
			name: "enum with case sensitivity",
			paramDef: neurotypes.ParameterDefinition{
				Name:        "reasoning_effort",
				Type:        "enum",
				Required:    false,
				Description: "Reasoning effort level",
				Constraints: &neurotypes.ParameterConstraints{
					EnumValues: []string{"low", "medium", "high"},
				},
			},
			inputValue:  "HIGH",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := map[string]string{tt.paramDef.Name: tt.inputValue}
			paramDefs := []neurotypes.ParameterDefinition{tt.paramDef}

			result, err := service.ValidateParameters(args, paramDefs)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedType, result[tt.paramDef.Name])
			}
		})
	}
}

func TestParameterValidatorService_ValidateParameters_RequiredParameters(t *testing.T) {
	service := NewParameterValidatorService()
	require.NoError(t, service.Initialize())

	paramDefs := []neurotypes.ParameterDefinition{
		{
			Name:        "required_param",
			Type:        "string",
			Required:    true,
			Description: "Required parameter",
		},
		{
			Name:        "optional_param",
			Type:        "string",
			Required:    false,
			Description: "Optional parameter",
		},
	}

	t.Run("missing required parameter", func(t *testing.T) {
		args := map[string]string{
			"optional_param": "value",
		}

		_, err := service.ValidateParameters(args, paramDefs)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "required parameter 'required_param' is missing")
	})

	t.Run("all required parameters provided", func(t *testing.T) {
		args := map[string]string{
			"required_param": "value1",
			"optional_param": "value2",
		}

		result, err := service.ValidateParameters(args, paramDefs)
		assert.NoError(t, err)
		assert.Equal(t, "value1", result["required_param"])
		assert.Equal(t, "value2", result["optional_param"])
	})

	t.Run("only required parameters provided", func(t *testing.T) {
		args := map[string]string{
			"required_param": "value1",
		}

		result, err := service.ValidateParameters(args, paramDefs)
		assert.NoError(t, err)
		assert.Equal(t, "value1", result["required_param"])
		assert.NotContains(t, result, "optional_param")
	})
}

func TestParameterValidatorService_ValidateParameters_DefaultValues(t *testing.T) {
	service := NewParameterValidatorService()
	require.NoError(t, service.Initialize())

	paramDefs := []neurotypes.ParameterDefinition{
		{
			Name:        "temperature",
			Type:        "float",
			Required:    false,
			Default:     1.0,
			Description: "Temperature with default",
		},
		{
			Name:        "max_tokens",
			Type:        "int",
			Required:    false,
			Default:     1000,
			Description: "Max tokens with default",
		},
	}

	t.Run("default values applied when parameters not provided", func(t *testing.T) {
		args := map[string]string{}

		result, err := service.ValidateParameters(args, paramDefs)
		assert.NoError(t, err)
		assert.Equal(t, 1.0, result["temperature"])
		assert.Equal(t, 1000, result["max_tokens"])
	})

	t.Run("provided values override defaults", func(t *testing.T) {
		args := map[string]string{
			"temperature": "0.5",
			"max_tokens":  "2000",
		}

		result, err := service.ValidateParameters(args, paramDefs)
		assert.NoError(t, err)
		assert.Equal(t, 0.5, result["temperature"])
		assert.Equal(t, 2000, result["max_tokens"])
	})
}

func TestParameterValidatorService_ValidateParameters_MultipleParameters(t *testing.T) {
	service := NewParameterValidatorService()
	require.NoError(t, service.Initialize())

	paramDefs := []neurotypes.ParameterDefinition{
		{
			Name:        "temperature",
			Type:        "float",
			Required:    false,
			Default:     1.0,
			Description: "Temperature",
			Constraints: &neurotypes.ParameterConstraints{
				Min: stringprocessing.Float64Ptr(0.0),
				Max: stringprocessing.Float64Ptr(2.0),
			},
		},
		{
			Name:        "max_tokens",
			Type:        "int",
			Required:    false,
			Description: "Max tokens",
			Constraints: &neurotypes.ParameterConstraints{
				Min: stringprocessing.Float64Ptr(1),
				Max: stringprocessing.Float64Ptr(16384),
			},
		},
		{
			Name:        "reasoning_effort",
			Type:        "enum",
			Required:    false,
			Default:     "medium",
			Description: "Reasoning effort",
			Constraints: &neurotypes.ParameterConstraints{
				EnumValues: []string{"low", "medium", "high"},
			},
		},
		{
			Name:        "stream",
			Type:        "bool",
			Required:    false,
			Default:     false,
			Description: "Enable streaming",
		},
	}

	t.Run("all valid parameters", func(t *testing.T) {
		args := map[string]string{
			"temperature":      "0.7",
			"max_tokens":       "8192",
			"reasoning_effort": "high",
			"stream":           "true",
		}

		result, err := service.ValidateParameters(args, paramDefs)
		assert.NoError(t, err)
		assert.Equal(t, 0.7, result["temperature"])
		assert.Equal(t, 8192, result["max_tokens"])
		assert.Equal(t, "high", result["reasoning_effort"])
		assert.Equal(t, true, result["stream"])
	})

	t.Run("mixed valid and invalid parameters", func(t *testing.T) {
		args := map[string]string{
			"temperature":      "0.7",    // valid
			"max_tokens":       "50000",  // invalid - too high
			"reasoning_effort": "medium", // valid
			"stream":           "true",   // valid
		}

		_, err := service.ValidateParameters(args, paramDefs)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "max_tokens")
	})
}

func TestParameterValidatorService_ValidateParameters_UnsupportedType(t *testing.T) {
	service := NewParameterValidatorService()
	require.NoError(t, service.Initialize())

	paramDefs := []neurotypes.ParameterDefinition{
		{
			Name:        "unsupported_param",
			Type:        "array", // unsupported type
			Required:    false,
			Description: "Unsupported parameter type",
		},
	}

	args := map[string]string{
		"unsupported_param": "value",
	}

	_, err := service.ValidateParameters(args, paramDefs)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported parameter type")
}

func TestParameterValidatorService_ValidateParameters_EdgeCases(t *testing.T) {
	service := NewParameterValidatorService()
	require.NoError(t, service.Initialize())

	t.Run("empty parameter definitions with no args", func(t *testing.T) {
		args := map[string]string{}
		paramDefs := []neurotypes.ParameterDefinition{}

		result, err := service.ValidateParameters(args, paramDefs)
		assert.NoError(t, err)
		assert.Empty(t, result)
	})

	t.Run("empty parameter definitions with unknown args", func(t *testing.T) {
		args := map[string]string{
			"some_param": "value",
		}
		paramDefs := []neurotypes.ParameterDefinition{}

		_, err := service.ValidateParameters(args, paramDefs)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unknown parameter 'some_param'")
	})

	t.Run("empty args", func(t *testing.T) {
		args := map[string]string{}
		paramDefs := []neurotypes.ParameterDefinition{
			{
				Name:        "optional_param",
				Type:        "string",
				Required:    false,
				Description: "Optional parameter",
			},
		}

		result, err := service.ValidateParameters(args, paramDefs)
		assert.NoError(t, err)
		assert.Empty(t, result)
	})

	t.Run("nil constraints", func(t *testing.T) {
		args := map[string]string{
			"param": "value",
		}
		paramDefs := []neurotypes.ParameterDefinition{
			{
				Name:        "param",
				Type:        "string",
				Required:    false,
				Description: "Parameter without constraints",
				Constraints: nil,
			},
		}

		result, err := service.ValidateParameters(args, paramDefs)
		assert.NoError(t, err)
		assert.Equal(t, "value", result["param"])
	})
}

// Interface compliance check
var _ neurotypes.Service = (*ParameterValidatorService)(nil)
