// Package builtin provides built-in NeuroShell commands that are available by default.
// Individual commands register themselves via their own init() functions.
// Build tags are used for conditional compilation during state machine transition.
package builtin

// NOTE: This package follows Go best practices where each command file
// contains its own init() function for registration. This provides:
// - Modular, self-contained command registration
// - Easy addition/removal of commands
// - Support for conditional compilation via build tags
// - Adherence to Go language idioms
//
// During state machine transition, build tags control which commands are included:
// - Normal build: `go build` (includes all commands)
// - Minimal build: `go build -tags minimal` (includes only essential commands)
//
// Each command file uses appropriate build tags:
// - Essential commands (echo): No build tag (always included)
// - Optional commands (set, get, etc.): `//go:build !minimal` (excluded in minimal builds)
