package context

import (
	"strings"
	"sync"
)

// ProviderRegistrySubcontext defines the interface for LLM provider registry functionality.
// This manages the central registry of supported LLM providers and their configurations.
type ProviderRegistrySubcontext interface {
	// Provider operations
	GetSupportedProviders() []string
	SetSupportedProviders(providers []string)
	IsProviderSupported(provider string) bool

	// Environment prefix operations
	GetProviderEnvPrefixes() []string
	SetProviderEnvPrefixes(prefixes []string)
	AddProviderEnvPrefix(prefix string)
}

// providerRegistrySubcontext implements the ProviderRegistrySubcontext interface.
type providerRegistrySubcontext struct {
	// Provider registry - central source of truth for supported providers
	supportedProviders  []string     // Supported LLM provider names (lowercase)
	providerEnvPrefixes []string     // Environment variable prefixes for provider detection
	providerMutex       sync.RWMutex // Protects provider configuration
}

// NewProviderRegistrySubcontext creates a new ProviderRegistrySubcontext instance.
func NewProviderRegistrySubcontext() ProviderRegistrySubcontext {
	return &providerRegistrySubcontext{
		// Initialize provider registry with default supported providers
		supportedProviders:  []string{"openai", "anthropic", "openrouter", "moonshot", "gemini"},
		providerEnvPrefixes: []string{"NEURO_", "OPENAI_", "ANTHROPIC_", "MOONSHOT_", "GOOGLE_"},
	}
}

// NewProviderRegistrySubcontextFromContext creates a ProviderRegistrySubcontext from an existing NeuroContext.
// This is used by services to get a reference to the context's provider registry subcontext.
func NewProviderRegistrySubcontextFromContext(ctx *NeuroContext) ProviderRegistrySubcontext {
	return ctx.providerRegistryCtx
}

// GetSupportedProviders returns a copy of the supported providers list (thread-safe read).
func (p *providerRegistrySubcontext) GetSupportedProviders() []string {
	p.providerMutex.RLock()
	defer p.providerMutex.RUnlock()

	providers := make([]string, len(p.supportedProviders))
	copy(providers, p.supportedProviders)
	return providers
}

// SetSupportedProviders sets the supported providers list (thread-safe write).
func (p *providerRegistrySubcontext) SetSupportedProviders(providers []string) {
	p.providerMutex.Lock()
	defer p.providerMutex.Unlock()

	p.supportedProviders = make([]string, len(providers))
	copy(p.supportedProviders, providers)
}

// IsProviderSupported checks if a provider is supported (case-insensitive).
func (p *providerRegistrySubcontext) IsProviderSupported(provider string) bool {
	p.providerMutex.RLock()
	defer p.providerMutex.RUnlock()

	provider = strings.ToLower(provider)
	for _, supported := range p.supportedProviders {
		if supported == provider {
			return true
		}
	}
	return false
}

// GetProviderEnvPrefixes returns a copy of the provider environment prefixes list (thread-safe read).
func (p *providerRegistrySubcontext) GetProviderEnvPrefixes() []string {
	p.providerMutex.RLock()
	defer p.providerMutex.RUnlock()

	prefixes := make([]string, len(p.providerEnvPrefixes))
	copy(prefixes, p.providerEnvPrefixes)
	return prefixes
}

// SetProviderEnvPrefixes sets the provider environment prefixes list (thread-safe write).
func (p *providerRegistrySubcontext) SetProviderEnvPrefixes(prefixes []string) {
	p.providerMutex.Lock()
	defer p.providerMutex.Unlock()

	p.providerEnvPrefixes = make([]string, len(prefixes))
	copy(p.providerEnvPrefixes, prefixes)
}

// AddProviderEnvPrefix adds a new environment prefix if it doesn't already exist.
func (p *providerRegistrySubcontext) AddProviderEnvPrefix(prefix string) {
	p.providerMutex.Lock()
	defer p.providerMutex.Unlock()

	// Check if prefix already exists
	for _, existing := range p.providerEnvPrefixes {
		if existing == prefix {
			return // Already exists, no need to add
		}
	}

	p.providerEnvPrefixes = append(p.providerEnvPrefixes, prefix)
}
