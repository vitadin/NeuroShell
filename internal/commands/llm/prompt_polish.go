// Package llm contains LLM-related commands for the NeuroShell CLI.
package llm

import (
	"fmt"
	"strings"

	"neuroshell/internal/commands"
	"neuroshell/internal/services"
	"neuroshell/pkg/neurotypes"
)

// PromptPolishCommand implements the \prompt-polish command as a delegation wrapper to the _prompt_polish neuro script.
// This provides clean help system integration while keeping the complex logic in the neuro script.
type PromptPolishCommand struct{}

// Name returns the command name "prompt-polish" for registration and lookup.
func (c *PromptPolishCommand) Name() string {
	return "prompt-polish"
}

// ParseMode returns ParseModeKeyValue for bracket parameter parsing.
func (c *PromptPolishCommand) ParseMode() neurotypes.ParseMode {
	return neurotypes.ParseModeKeyValue
}

// Description returns a brief description of what the prompt-polish command does.
func (c *PromptPolishCommand) Description() string {
	return "Optimize and correct English text for better LLM comprehension"
}

// Usage returns the syntax and usage examples for the prompt-polish command.
func (c *PromptPolishCommand) Usage() string {
	return `\prompt-polish[instruction="custom prompt", model="G5MR"] text to polish`
}

// HelpInfo returns structured help information for the prompt-polish command.
func (c *PromptPolishCommand) HelpInfo() neurotypes.HelpInfo {
	return neurotypes.HelpInfo{
		Command:     c.Name(),
		Description: c.Description(),
		Usage:       c.Usage(),
		ParseMode:   c.ParseMode(),
		Options: []neurotypes.HelpOption{
			{
				Name:        "instruction",
				Description: "Custom system prompt to override default polishing behavior",
				Required:    false,
				Type:        "string",
			},
			{
				Name:        "model",
				Description: "Model catalog ID to use for polishing",
				Required:    false,
				Type:        "string",
				Default:     "G5MR",
			},
		},
		Examples: []neurotypes.HelpExample{
			{
				Command:     `\prompt-polish Fix grammar and improve clarity`,
				Description: "Basic text polishing with default settings",
			},
			{
				Command:     `\prompt-polish[instruction="Make it formal and professional"] Hey, can you help me?`,
				Description: "Polish text with custom tone instruction",
			},
			{
				Command:     `\prompt-polish[model="G4OC"] Complex technical documentation to optimize`,
				Description: "Use different model for polishing",
			},
			{
				Command:     `\prompt-polish ${_output}`,
				Description: "Polish output from previous command",
			},
		},
		StoredVariables: []neurotypes.HelpStoredVariable{
			{
				Name:        "_output",
				Description: "The polished and optimized text",
				Type:        "command_output",
				Example:     "Please help me fix the grammar and improve clarity.",
			},
		},
		Notes: []string{
			"Creates temporary session for processing - no session state preserved",
			"Default system prompt optimizes for LLM comprehension while preserving meaning",
			"Requires OpenAI API key (OPENAI_API_KEY environment variable)",
			"Uses G5MR (gpt-5-mini-responses) model by default for efficiency",
			"Custom instruction completely replaces default polishing prompt",
			"Automatically cleans up temporary session after processing",
			"Leverages existing NeuroShell infrastructure via delegation to _prompt_polish script",
		},
	}
}

// Execute delegates to the _prompt_polish neuro script via stack service with options.
func (c *PromptPolishCommand) Execute(options map[string]string, input string) error {
	// Input validation
	if strings.TrimSpace(input) == "" {
		fmt.Printf("Usage: %s\n\n", c.Usage())
		return nil
	}

	// Get stack service for delegation
	stackService, err := services.GetGlobalStackService()
	if err != nil {
		return fmt.Errorf("stack service not available: %w", err)
	}

	// Build command with options for _prompt_polish neuro script
	command := "\\_prompt_polish"

	// Add options if provided
	if len(options) > 0 {
		command += "["
		first := true
		for key, value := range options {
			if !first {
				command += ","
			}
			command += key + "=" + value
			first = false
		}
		command += "]"
	}

	// Add the text input
	command += " " + input

	// Delegate to _prompt_polish neuro script
	stackService.PushCommand(command)

	return nil
}

// IsReadOnly returns false as the command sets variables.
func (c *PromptPolishCommand) IsReadOnly() bool {
	return false
}

// init registers the PromptPolishCommand with the global command registry.
func init() {
	if err := commands.GetGlobalRegistry().Register(&PromptPolishCommand{}); err != nil {
		panic(fmt.Sprintf("failed to register prompt-polish command: %v", err))
	}
}
