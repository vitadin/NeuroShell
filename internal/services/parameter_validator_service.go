package services

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"neuroshell/pkg/neurotypes"
)

// ParameterValidatorService provides generic parameter validation based on model catalog definitions.
// It validates parameters according to their type, constraints, and requirements as defined in model YAML files.
type ParameterValidatorService struct {
	initialized bool
}

// NewParameterValidatorService creates a new ParameterValidatorService instance.
func NewParameterValidatorService() *ParameterValidatorService {
	return &ParameterValidatorService{
		initialized: false,
	}
}

// Name returns the service name "parameter_validator" for registration.
func (p *ParameterValidatorService) Name() string {
	return "parameter_validator"
}

// Initialize sets up the ParameterValidatorService for operation.
func (p *ParameterValidatorService) Initialize() error {
	p.initialized = true
	return nil
}

// ValidateParameters validates a map of parameter values against parameter definitions from a model.
// It returns a map of validated/coerced values and an error if validation fails.
func (p *ParameterValidatorService) ValidateParameters(args map[string]string, parameterDefs []neurotypes.ParameterDefinition) (map[string]any, error) {
	if !p.initialized {
		return nil, fmt.Errorf("parameter validator service not initialized")
	}

	result := make(map[string]any)
	paramDefMap := make(map[string]neurotypes.ParameterDefinition)

	// Create a lookup map for parameter definitions
	for _, paramDef := range parameterDefs {
		paramDefMap[paramDef.Name] = paramDef
	}

	// Validate provided parameters
	for paramName, paramValue := range args {
		// Skip non-parameter arguments (like catalog_id, description)
		if paramName == "catalog_id" || paramName == "description" {
			continue
		}

		paramDef, exists := paramDefMap[paramName]
		if !exists {
			return nil, fmt.Errorf("unknown parameter '%s'", paramName)
		}

		validatedValue, err := p.validateSingleParameter(paramValue, paramDef)
		if err != nil {
			return nil, fmt.Errorf("parameter '%s': %w", paramName, err)
		}

		result[paramName] = validatedValue
	}

	// Check for missing required parameters and set defaults
	for _, paramDef := range parameterDefs {
		if _, provided := args[paramDef.Name]; !provided {
			if paramDef.Required {
				return nil, fmt.Errorf("required parameter '%s' is missing", paramDef.Name)
			}
			// Set default value if provided
			if paramDef.Default != nil {
				result[paramDef.Name] = paramDef.Default
			}
		}
	}

	return result, nil
}

// validateSingleParameter validates a single parameter value against its definition.
func (p *ParameterValidatorService) validateSingleParameter(value string, paramDef neurotypes.ParameterDefinition) (any, error) {
	// Handle empty values
	if strings.TrimSpace(value) == "" {
		if paramDef.Required {
			return nil, fmt.Errorf("required parameter cannot be empty")
		}
		if paramDef.Default != nil {
			return paramDef.Default, nil
		}
		return nil, nil
	}

	switch paramDef.Type {
	case "string":
		return p.validateStringParameter(value, paramDef)
	case "int":
		return p.validateIntParameter(value, paramDef)
	case "float":
		return p.validateFloatParameter(value, paramDef)
	case "bool":
		return p.validateBoolParameter(value, paramDef)
	case "enum":
		return p.validateEnumParameter(value, paramDef)
	default:
		return nil, fmt.Errorf("unsupported parameter type '%s'", paramDef.Type)
	}
}

// validateStringParameter validates a string parameter.
func (p *ParameterValidatorService) validateStringParameter(value string, paramDef neurotypes.ParameterDefinition) (string, error) {
	if paramDef.Constraints != nil && paramDef.Constraints.Pattern != nil {
		matched, err := regexp.MatchString(*paramDef.Constraints.Pattern, value)
		if err != nil {
			return "", fmt.Errorf("invalid regex pattern in parameter definition: %w", err)
		}
		if !matched {
			return "", fmt.Errorf("value '%s' does not match required pattern '%s'", value, *paramDef.Constraints.Pattern)
		}
	}
	return value, nil
}

// validateIntParameter validates an integer parameter.
func (p *ParameterValidatorService) validateIntParameter(value string, paramDef neurotypes.ParameterDefinition) (int, error) {
	intValue, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("invalid integer value '%s'", value)
	}

	if paramDef.Constraints != nil {
		if paramDef.Constraints.Min != nil && float64(intValue) < *paramDef.Constraints.Min {
			return 0, fmt.Errorf("value %d is below minimum %g", intValue, *paramDef.Constraints.Min)
		}
		if paramDef.Constraints.Max != nil && float64(intValue) > *paramDef.Constraints.Max {
			return 0, fmt.Errorf("value %d is above maximum %g", intValue, *paramDef.Constraints.Max)
		}
	}

	return intValue, nil
}

// validateFloatParameter validates a float parameter.
func (p *ParameterValidatorService) validateFloatParameter(value string, paramDef neurotypes.ParameterDefinition) (float64, error) {
	floatValue, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid float value '%s'", value)
	}

	if paramDef.Constraints != nil {
		if paramDef.Constraints.Min != nil && floatValue < *paramDef.Constraints.Min {
			return 0, fmt.Errorf("value %g is below minimum %g", floatValue, *paramDef.Constraints.Min)
		}
		if paramDef.Constraints.Max != nil && floatValue > *paramDef.Constraints.Max {
			return 0, fmt.Errorf("value %g is above maximum %g", floatValue, *paramDef.Constraints.Max)
		}
	}

	return floatValue, nil
}

// validateBoolParameter validates a boolean parameter.
func (p *ParameterValidatorService) validateBoolParameter(value string, _ neurotypes.ParameterDefinition) (bool, error) {
	lowerValue := strings.ToLower(strings.TrimSpace(value))
	switch lowerValue {
	case "true", "1", "yes", "on":
		return true, nil
	case "false", "0", "no", "off":
		return false, nil
	default:
		return false, fmt.Errorf("invalid boolean value '%s' (use true/false, 1/0, yes/no, on/off)", value)
	}
}

// validateEnumParameter validates an enumeration parameter.
func (p *ParameterValidatorService) validateEnumParameter(value string, paramDef neurotypes.ParameterDefinition) (string, error) {
	if paramDef.Constraints == nil || len(paramDef.Constraints.EnumValues) == 0 {
		return "", fmt.Errorf("enum parameter has no valid values defined")
	}

	for _, validValue := range paramDef.Constraints.EnumValues {
		if value == validValue {
			return value, nil
		}
	}

	return "", fmt.Errorf("invalid value '%s'. Valid values: %s", value, strings.Join(paramDef.Constraints.EnumValues, ", "))
}

// GetParameterHelp returns help information for a specific parameter.
func (p *ParameterValidatorService) GetParameterHelp(paramName string, parameterDefs []neurotypes.ParameterDefinition) (string, error) {
	for _, paramDef := range parameterDefs {
		if paramDef.Name == paramName {
			help := fmt.Sprintf("%s (%s)", paramDef.Description, paramDef.Type)
			if paramDef.Required {
				help += " [REQUIRED]"
			} else if paramDef.Default != nil {
				help += fmt.Sprintf(" [default: %v]", paramDef.Default)
			}

			if paramDef.Constraints != nil {
				switch {
				case paramDef.Type == "enum" && len(paramDef.Constraints.EnumValues) > 0:
					help += fmt.Sprintf(" (valid values: %s)", strings.Join(paramDef.Constraints.EnumValues, ", "))
				case paramDef.Constraints.Min != nil && paramDef.Constraints.Max != nil:
					help += fmt.Sprintf(" (range: %g-%g)", *paramDef.Constraints.Min, *paramDef.Constraints.Max)
				case paramDef.Constraints.Min != nil:
					help += fmt.Sprintf(" (min: %g)", *paramDef.Constraints.Min)
				case paramDef.Constraints.Max != nil:
					help += fmt.Sprintf(" (max: %g)", *paramDef.Constraints.Max)
				}
			}

			return help, nil
		}
	}
	return "", fmt.Errorf("parameter '%s' not found", paramName)
}

// GetGlobalParameterValidatorService returns the global ParameterValidatorService instance.
func GetGlobalParameterValidatorService() (*ParameterValidatorService, error) {
	service, err := GetGlobalRegistry().GetService("parameter_validator")
	if err != nil {
		return nil, err
	}

	paramValidatorService, ok := service.(*ParameterValidatorService)
	if !ok {
		return nil, fmt.Errorf("service 'parameter_validator' is not a ParameterValidatorService")
	}

	return paramValidatorService, nil
}

func init() {
	if err := GlobalRegistry.RegisterService(NewParameterValidatorService()); err != nil {
		panic(fmt.Sprintf("failed to register parameter validator service: %v", err))
	}
}
