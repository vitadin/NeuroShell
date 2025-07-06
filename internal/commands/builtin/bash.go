package builtin

import (
	"fmt"
	"strings"

	"neuroshell/internal/commands"
	"neuroshell/internal/services"
	"neuroshell/pkg/neurotypes"
)

// BashCommand implements the \bash command for executing system commands.
// It uses exec.Command("bash", "-c", command) for simple, safe command execution.
type BashCommand struct{}

// Name returns the command name "bash" for registration and lookup.
func (c *BashCommand) Name() string {
	return "bash"
}

// ParseMode returns ParseModeRaw to pass entire input directly to bash.
func (c *BashCommand) ParseMode() neurotypes.ParseMode {
	return neurotypes.ParseModeRaw
}

// Description returns a brief description of what the bash command does.
func (c *BashCommand) Description() string {
	return "Execute system commands via bash"
}

// Usage returns the syntax and usage examples for the bash command.
func (c *BashCommand) Usage() string {
	return "\\bash command_to_execute"
}

// HelpInfo returns structured help information for the bash command.
func (c *BashCommand) HelpInfo() neurotypes.HelpInfo {
	return neurotypes.HelpInfo{
		Command:     c.Name(),
		Description: c.Description(),
		Usage:       c.Usage(),
		ParseMode:   c.ParseMode(),
		Examples: []neurotypes.HelpExample{
			{
				Command:     "\\bash ls -la",
				Description: "List directory contents with details",
			},
			{
				Command:     "\\bash pwd",
				Description: "Show current working directory",
			},
			{
				Command:     "\\bash echo \"Hello, ${name}!\"",
				Description: "Execute command with variable interpolation",
			},
			{
				Command:     "\\bash cat file.txt | grep pattern",
				Description: "Use pipes and command chaining",
			},
			{
				Command:     "\\bash python -c \"print('hello')\"",
				Description: "Execute Python or other language commands",
			},
		},
		Notes: []string{
			"Commands are executed using bash -c for full shell capabilities",
			"Variables are interpolated before execution",
			"Output is stored in ${_output}, errors in ${_error}, exit code in ${_status}",
			"Supports pipes, redirection, and all bash features",
			"Use quotes to protect special characters from shell expansion",
			"Commands run with a configurable timeout (default: 2 minutes)",
		},
	}
}

// Execute runs system commands using bash and sets _output, _error, and _status variables.
func (c *BashCommand) Execute(_ map[string]string, input string) error {
	// Get the command to execute
	command := strings.TrimSpace(input)
	if command == "" {
		return fmt.Errorf("Usage: %s", c.Usage())
	}

	// Get bash service from global registry
	bashService, err := services.GetGlobalBashService()
	if err != nil {
		return fmt.Errorf("bash service not available: %w", err)
	}

	// Execute the command
	stdout, stderr, exitCode, err := bashService.Execute(command)
	if err != nil {
		return fmt.Errorf("failed to execute command: %w", err)
	}

	// Display output to user
	if stdout != "" {
		fmt.Print(stdout)
		if !strings.HasSuffix(stdout, "\n") {
			fmt.Println() // Add newline if output doesn't end with one
		}
	}

	if stderr != "" {
		fmt.Printf("Error: %s\n", stderr)
	}

	// Display exit status if non-zero
	if exitCode != 0 {
		fmt.Printf("Exit status: %d\n", exitCode)
	}

	return nil
}

func init() {
	if err := commands.GlobalRegistry.Register(&BashCommand{}); err != nil {
		panic(fmt.Sprintf("failed to register bash command: %v", err))
	}
}
