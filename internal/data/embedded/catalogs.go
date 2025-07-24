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

// KimiK2FreeOpenRouterModelData contains the embedded Kimi K2 Free (OpenRouter) model YAML data.
//
//go:embed models/kimi-k2-free-openrouter.yaml
var KimiK2FreeOpenRouterModelData []byte

// KimiK2OpenRouterModelData contains the embedded Kimi K2 (OpenRouter) model YAML data.
//
//go:embed models/kimi-k2-openrouter.yaml
var KimiK2OpenRouterModelData []byte

// KimiK2MoonshotModelData contains the embedded Kimi K2 (Moonshot) model YAML data.
//
//go:embed models/kimi-k2-moonshot.yaml
var KimiK2MoonshotModelData []byte

// Qwen3235BOpenRouterModelData contains the embedded Qwen3-235B (OpenRouter) model YAML data.
//
//go:embed models/qwen3-235b-openrouter.yaml
var Qwen3235BOpenRouterModelData []byte

// Grok4OpenRouterModelData contains the embedded Grok-4 (OpenRouter) model YAML data.
//
//go:embed models/grok-4-openrouter.yaml
var Grok4OpenRouterModelData []byte

// GPT41ModelData contains the embedded GPT-4.1 model YAML data.
//
//go:embed models/gpt-4-1-openai.yaml
var GPT41ModelData []byte

// O3ProModelData contains the embedded o3-pro model YAML data.
//
//go:embed models/o3-pro-openai.yaml
var O3ProModelData []byte
