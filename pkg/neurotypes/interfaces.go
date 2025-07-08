// Package neurotypes defines core interfaces and data structures used throughout NeuroShell.
//
// This package contains the fundamental types that enable NeuroShell's modular architecture,
// providing the contracts and data structures that components use to interact with each other.
// The package is organized into logical groupings to improve maintainability and discoverability.
//
// # Architecture Overview
//
// NeuroShell follows a three-layer architecture:
//
//   - Context Layer: Holds ALL state and resources (variables, sessions, models, etc.)
//   - Service Layer: Stateless business logic that operates on the global context
//   - Command Layer: Orchestrates services to handle user input and commands
//
// # Package Organization
//
// The neurotypes package is organized into the following files:
//
// ## Core Interfaces (core_interfaces.go)
//
// Contains the fundamental interfaces that define the system's structure:
//
//   - Context: Session state management and variable interpolation
//   - Service: Services that provide specific functionality
//   - Command: Commands that handle user input and actions
//   - ServiceRegistry: Centralized service registration and retrieval
//
// ## Command System Types (command_types.go)
//
// Types for command parsing, execution, and help system:
//
//   - ParseMode: Defines how command arguments are parsed
//   - CommandArgs: Structured command arguments and message content
//   - HelpInfo, HelpOption, HelpExample: Rich help system data structures
//
// ## Session and Conversation Types (session_types.go)
//
// Types for managing conversation history and session state:
//
//   - Message: Individual messages in conversation history
//   - SessionState: Complete session state for save/restore operations
//   - ChatSession: LLM conversation sessions with metadata
//
// ## Model Configuration Types (model_types.go)
//
// Types for LLM model management and configuration:
//
//   - ModelConfig: User-defined model configurations with parameters
//   - ModelCatalogEntry: Catalog entries for available models
//   - ModelProviderInfo: Provider metadata and capabilities
//   - StandardModelParameters: Common parameters across providers
//
// ## Theme and Styling Types (theme_types.go)
//
// Types for theme configuration and visual styling:
//
//   - ThemeConfig: Complete theme configuration from YAML
//   - ThemeStyles: Styling for different semantic elements
//   - StyleConfig: Visual styling for individual elements
//   - AdaptiveColor: Colors that adapt to light/dark terminals
//
// # Usage Patterns
//
// The types in this package are designed to be used together in specific patterns:
//
// ## Service Pattern
//
// Services implement the Service interface and use the global context singleton:
//
//	type MyService struct {
//		initialized bool
//	}
//
//	func (s *MyService) Name() string { return "my" }
//	func (s *MyService) Initialize() error {
//		ctx := context.GetGlobalContext()
//		// Use context for state access
//		return nil
//	}
//
// ## Command Pattern
//
// Commands implement the Command interface and orchestrate services:
//
//	type MyCommand struct{}
//
//	func (c *MyCommand) Name() string { return "my" }
//	func (c *MyCommand) ParseMode() ParseMode { return ParseModeKeyValue }
//	func (c *MyCommand) Execute(args map[string]string, input string) error {
//		service, err := services.GetGlobalService("my")
//		// Use service to perform operations
//		return nil
//	}
//
// ## Context Pattern
//
// The Context interface provides access to all session state:
//
//	ctx := context.GetGlobalContext()
//	value, err := ctx.GetVariable("var_name")
//	err = ctx.SetVariable("var_name", "value")
//	sessions := ctx.GetChatSessions()
//
// # Design Principles
//
// The types in this package follow these design principles:
//
//   - Single Responsibility: Each type has a clear, focused purpose
//   - Interface Segregation: Interfaces are focused and composable
//   - Dependency Inversion: Components depend on abstractions, not implementations
//   - Immutability: Data structures favor immutable operations where possible
//   - Extensibility: New functionality can be added without breaking existing code
//
// # Backward Compatibility
//
// All types maintain backward compatibility. Existing code using:
//
//	import "neuroshell/pkg/neurotypes"
//
//	var ctx neurotypes.Context
//	var cmd neurotypes.Command
//	var svc neurotypes.Service
//
// Will continue to work without any changes.
package neurotypes
