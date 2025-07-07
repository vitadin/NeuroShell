// Package embedded provides access to embedded model catalog data files.
package embedded

import _ "embed"

// O3ModelData contains the embedded O3 model YAML data.
//
//go:embed models/o3.yaml
var O3ModelData []byte

// O4MiniModelData contains the embedded O4-mini model YAML data.
//
//go:embed models/o4-mini.yaml
var O4MiniModelData []byte

// Claude37SonnetModelData contains the embedded Claude 3.7 Sonnet model YAML data.
//
//go:embed models/claude-3-7-sonnet.yaml
var Claude37SonnetModelData []byte

// ClaudeSonnet4ModelData contains the embedded Claude Sonnet 4 model YAML data.
//
//go:embed models/claude-sonnet-4.yaml
var ClaudeSonnet4ModelData []byte

// Claude37OpusModelData contains the embedded Claude 3.7 Opus model YAML data.
//
//go:embed models/claude-3-7-opus.yaml
var Claude37OpusModelData []byte

// ClaudeOpus4ModelData contains the embedded Claude Opus 4 model YAML data.
//
//go:embed models/claude-opus-4.yaml
var ClaudeOpus4ModelData []byte
