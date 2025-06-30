// Package builtin provides built-in NeuroShell commands that are available by default.
// This file imports all builtin command packages for side effects (init functions).
package builtin

import (
	// Import bash package for command registration
	_ "neuroshell/internal/commands/builtin/bash"
)
