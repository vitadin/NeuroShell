// Package embedded provides access to embedded model catalog data files.
package embedded

import _ "embed"

//go:embed anthropic.yaml
var AnthropicCatalogData []byte

//go:embed openai.yaml
var OpenaiCatalogData []byte