// Package embedded provides access to embedded model catalog data files.
package embedded

import _ "embed"

// O3ModelData contains the embedded O3 model YAML data.
//
//go:embed models/o3.yaml
var O3ModelData []byte

// O4MiniChatModelData contains the embedded O4-mini Chat model YAML data.
//
//go:embed models/o4-mini-chat.yaml
var O4MiniChatModelData []byte

// O4MiniReasoningModelData contains the embedded O4-mini Reasoning model YAML data.
//
//go:embed models/o4-mini-reasoning.yaml
var O4MiniReasoningModelData []byte

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

// ClaudeOpus41ModelData contains the embedded Claude Opus 4.1 model YAML data.
//
//go:embed models/claude-opus-4-1.yaml
var ClaudeOpus41ModelData []byte

// GPT41ModelData contains the embedded GPT-4.1 model YAML data.
//
//go:embed models/gpt-4-1-openai.yaml
var GPT41ModelData []byte

// O3ProModelData contains the embedded o3-pro model YAML data.
//
//go:embed models/o3-pro-openai.yaml
var O3ProModelData []byte

// O1ModelData contains the embedded o1 model YAML data.
//
//go:embed models/o1-openai.yaml
var O1ModelData []byte

// GPT5ChatModelData contains the embedded GPT-5 Chat model YAML data.
//
//go:embed models/gpt-5-chat.yaml
var GPT5ChatModelData []byte

// GPT5ResponsesModelData contains the embedded GPT-5 Responses model YAML data.
//
//go:embed models/gpt-5-responses.yaml
var GPT5ResponsesModelData []byte

// GPT5MiniChatModelData contains the embedded GPT-5 Mini Chat model YAML data.
//
//go:embed models/gpt-5-mini-chat.yaml
var GPT5MiniChatModelData []byte

// GPT5MiniResponsesModelData contains the embedded GPT-5 Mini Responses model YAML data.
//
//go:embed models/gpt-5-mini-responses.yaml
var GPT5MiniResponsesModelData []byte

// GPT5NanoChatModelData contains the embedded GPT-5 Nano Chat model YAML data.
//
//go:embed models/gpt-5-nano-chat.yaml
var GPT5NanoChatModelData []byte

// GPT5NanoResponsesModelData contains the embedded GPT-5 Nano Responses model YAML data.
//
//go:embed models/gpt-5-nano-responses.yaml
var GPT5NanoResponsesModelData []byte

// GPT4oModelData contains the embedded GPT-4o model YAML data.
//
//go:embed models/gpt-4o-openai.yaml
var GPT4oModelData []byte

// O1ProModelData contains the embedded o1-pro model YAML data.
//
//go:embed models/o1-pro-openai.yaml
var O1ProModelData []byte

// Gemini25ProModelData contains the embedded Gemini 2.5 Pro model YAML data.
//
//go:embed models/gemini-2-5-pro.yaml
var Gemini25ProModelData []byte

// Gemini25FlashModelData contains the embedded Gemini 2.5 Flash model YAML data.
//
//go:embed models/gemini-2-5-flash.yaml
var Gemini25FlashModelData []byte

// Gemini25FlashLiteModelData contains the embedded Gemini 2.5 Flash Lite model YAML data.
//
//go:embed models/gemini-2-5-flash-lite.yaml
var Gemini25FlashLiteModelData []byte

// Provider Catalog Data - embedded provider configuration YAML files

// OpenAIChatProviderData contains the embedded OpenAI chat provider YAML data.
//
//go:embed providers/openai-chat.yaml
var OpenAIChatProviderData []byte

// AnthropicChatProviderData contains the embedded Anthropic chat provider YAML data.
//
//go:embed providers/anthropic-chat.yaml
var AnthropicChatProviderData []byte

// GeminiChatProviderData contains the embedded Gemini chat provider YAML data.
//
//go:embed providers/gemini-chat.yaml
var GeminiChatProviderData []byte

// OpenAIResponsesProviderData contains the embedded OpenAI responses provider YAML data.
//
//go:embed providers/openai-responses.yaml
var OpenAIResponsesProviderData []byte

// Change Log Data - embedded change log YAML file

// ChangeLogData contains the embedded change log YAML data.
//
//go:embed change-logs/change-logs.yaml
var ChangeLogData []byte
