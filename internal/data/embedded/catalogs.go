// Package embedded provides access to embedded model catalog data files.
package embedded

import _ "embed"

// AnthropicCatalogData contains the embedded Anthropic model catalog YAML data.
//
//go:embed anthropic.yaml
var AnthropicCatalogData []byte

// OpenaiCatalogData contains the embedded OpenAI model catalog YAML data.
//
//go:embed openai.yaml
var OpenaiCatalogData []byte
