// Package shell contains shell prompt related commands for configuring and managing the NeuroShell prompt display.
package shell

import (
	"fmt"
	"strconv"

	"neuroshell/internal/commands"
	neuroshellcontext "neuroshell/internal/context"
	"neuroshell/pkg/neurotypes"
)

// PromptCommand configures the shell prompt display
type PromptCommand struct{}

// Name returns the command name "shell-prompt" for registration and lookup.
func (c *PromptCommand) Name() string {
	return "shell-prompt"
}

// Description returns a brief description of what the shell-prompt command does.
func (c *PromptCommand) Description() string {
	return "Configure the shell prompt display"
}

// Usage returns the syntax and usage examples for the shell-prompt command.
func (c *PromptCommand) Usage() string {
	return `\shell-prompt[lines=N, line1="template", line2="template", ...]

Configure the shell prompt appearance with 1-5 lines, including colors and styles.

Options:
  lines=N          Set number of prompt lines (1-5)
  line1="template" Set template for first line
  line2="template" Set template for second line
  line3="template" Set template for third line
  line4="template" Set template for fourth line
  line5="template" Set template for fifth line

Color and Style Syntax:
  {{color:semantic}}text{{/color}}  - Use semantic colors (info, success, warning, error, etc.)
  {{color:#hex}}text{{/color}}      - Use hex colors (#ff0000, #00ff00, etc.)
  {{color:name}}text{{/color}}      - Use color names (red, blue, green, yellow, etc.)
  {{bold}}text{{/bold}}             - Bold text
  {{italic}}text{{/italic}}         - Italic text
  {{underline}}text{{/underline}}   - Underlined text
  
  Styles automatically adapt to terminal capabilities and user color preferences.

Examples:
  \shell-prompt[lines=1, line1="neuro> "]
  \shell-prompt[lines=1, line1="{{color:info}}neuro{{/color}}{{color:success}}>{{/color}} "]
  \shell-prompt[lines=2, line1="{{color:blue}}${@pwd}{{/color}} [{{color:yellow}}${#session_name}{{/color}}]", line2="{{bold}}❯{{/bold}} "]
  \shell-prompt[lines=3, line1="┌─[{{color:cyan}}${@time}{{/color}}]", line2="├─[{{color:blue}}${@pwd}{{/color}}]", line3="└➤ "]

Available variables:
  ${@pwd}           Current working directory
  ${@user}          Current username
  ${@date}          Current date
  ${@time}          Current time
  ${#session_name}  Active session name
  ${#message_count} Number of messages in session
  ${#active_model}  Currently active model

Available semantic colors:
  info, success, warning, error, command, variable, keyword, highlight`
}

// ParseMode returns ParseModeKeyValue for standard argument parsing.
func (c *PromptCommand) ParseMode() neurotypes.ParseMode {
	return neurotypes.ParseModeKeyValue
}

// HelpInfo returns structured help information for the shell-prompt command.
func (c *PromptCommand) HelpInfo() neurotypes.HelpInfo {
	return neurotypes.HelpInfo{
		Command:     c.Name(),
		Description: c.Description(),
		Usage:       c.Usage(),
		ParseMode:   c.ParseMode(),
		Options: []neurotypes.HelpOption{
			{
				Name:        "lines",
				Description: "Number of prompt lines (1-5)",
				Required:    false,
				Type:        "number",
			},
			{
				Name:        "line1",
				Description: "Template for first prompt line",
				Required:    false,
				Type:        "string",
			},
			{
				Name:        "line2",
				Description: "Template for second prompt line",
				Required:    false,
				Type:        "string",
			},
		},
		Examples: []neurotypes.HelpExample{
			{
				Command:     "\\shell-prompt[lines=2, line1=\"${@pwd}\", line2=\"❯ \"]",
				Description: "Set two-line prompt with path and arrow",
			},
			{
				Command:     "\\shell-prompt[lines=1, line1=\"{{color:info}}neuro{{/color}}{{color:success}}>{{/color}} \"]",
				Description: "Set colorized single-line prompt",
			},
			{
				Command:     "\\shell-prompt[lines=2, line1=\"{{color:blue}}${@pwd}{{/color}} [{{color:yellow}}${#session_name}{{/color}}]\", line2=\"{{bold}}❯{{/bold}} \"]",
				Description: "Set colorized two-line prompt with variables",
			},
			{
				Command:     "\\shell-prompt",
				Description: "Show current prompt configuration",
			},
		},
	}
}

// Execute configures the shell prompt display with specified options.
func (c *PromptCommand) Execute(options map[string]string, _ string) error {
	ctx := neuroshellcontext.GetGlobalContext().(*neuroshellcontext.NeuroContext)

	// Handle setting number of lines
	if lines, ok := options["lines"]; ok {
		if n, err := strconv.Atoi(lines); err == nil && n >= 1 && n <= 5 {
			if err := ctx.SetVariable("_prompt_lines_count", lines); err != nil {
				return fmt.Errorf("failed to set prompt lines count: %w", err)
			}
			fmt.Printf("Prompt lines set to %s\n", lines)
		} else {
			return fmt.Errorf("invalid lines value: must be 1-5, got: %s", lines)
		}
	}

	// Handle setting specific lines
	for i := 1; i <= 5; i++ {
		key := fmt.Sprintf("line%d", i)
		if template, ok := options[key]; ok {
			varName := fmt.Sprintf("_prompt_line%d", i)
			if err := ctx.SetVariable(varName, template); err != nil {
				return fmt.Errorf("failed to set prompt line %d: %w", i, err)
			}
			fmt.Printf("Prompt line %d set to: %s\n", i, template)
		}
	}

	// If no options provided, show current configuration
	if len(options) == 0 {
		return c.showCurrentConfiguration(ctx)
	}

	return nil
}

func (c *PromptCommand) showCurrentConfiguration(ctx *neuroshellcontext.NeuroContext) error {
	fmt.Println("Current shell prompt configuration:")

	// Get lines count
	linesCount, _ := ctx.GetVariable("_prompt_lines_count")
	if linesCount == "" {
		linesCount = "1" // default
	}
	fmt.Printf("  Lines: %s\n", linesCount)

	// Show configured lines
	count, _ := strconv.Atoi(linesCount)
	if count < 1 || count > 5 {
		count = 1
	}

	for i := 1; i <= count; i++ {
		varName := fmt.Sprintf("_prompt_line%d", i)
		template, _ := ctx.GetVariable(varName)
		if template == "" && i == 1 {
			template = "neuro> " // default
		}
		if template != "" {
			fmt.Printf("  Line %d: %s\n", i, template)
		}
	}

	return nil
}

// IsReadOnly returns false as this command modifies prompt configuration.
func (c *PromptCommand) IsReadOnly() bool {
	return false
}

func init() {
	if err := commands.GlobalRegistry.Register(&PromptCommand{}); err != nil {
		panic(fmt.Sprintf("failed to register shell-prompt command: %v", err))
	}
}
