package services

import (
	"fmt"

	"neuroshell/internal/context"
	"neuroshell/internal/logger"
	"neuroshell/internal/parser"
	"neuroshell/pkg/types"
)

type InterpolationService struct {
	initialized bool
}

func NewInterpolationService() *InterpolationService {
	return &InterpolationService{
		initialized: false,
	}
}

func (i *InterpolationService) Name() string {
	return "interpolation"
}

func (i *InterpolationService) Initialize(ctx types.Context) error {
	i.initialized = true
	logger.ServiceOperation("interpolation", "initialize", "service ready")
	return nil
}

// InterpolateString performs pure interpolation of a single string using context variables
func (i *InterpolationService) InterpolateString(text string, ctx types.Context) (string, error) {
	if !i.initialized {
		return "", fmt.Errorf("interpolation service not initialized")
	}

	neuroCtx, ok := ctx.(*context.NeuroContext)
	if !ok {
		return "", fmt.Errorf("context is not a NeuroContext")
	}

	result := neuroCtx.InterpolateVariables(text)
	if text != result {
		logger.Debug("String interpolation performed", "input", text, "output", result)
	}
	return result, nil
}

// InterpolateCommand interpolates all parts of a command structure and returns a new interpolated command
func (i *InterpolationService) InterpolateCommand(cmd *parser.Command, ctx types.Context) (*parser.Command, error) {
	if !i.initialized {
		return nil, fmt.Errorf("interpolation service not initialized")
	}

	neuroCtx, ok := ctx.(*context.NeuroContext)
	if !ok {
		return nil, fmt.Errorf("context is not a NeuroContext")
	}

	// Create new command with interpolated values
	interpolatedCmd := &parser.Command{
		Name:           cmd.Name, // Don't interpolate command name
		Message:        neuroCtx.InterpolateVariables(cmd.Message),
		BracketContent: neuroCtx.InterpolateVariables(cmd.BracketContent),
		Options:        make(map[string]string),
		ParseMode:      cmd.ParseMode,
	}

	// Interpolate option values
	for key, value := range cmd.Options {
		interpolatedCmd.Options[key] = neuroCtx.InterpolateVariables(value)
	}

	return interpolatedCmd, nil
}
