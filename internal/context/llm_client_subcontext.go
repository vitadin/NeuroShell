package context

import (
	"sync"

	"neuroshell/pkg/neurotypes"
)

// LLMClientSubcontext defines the interface for LLM client caching functionality.
// This manages cached LLM client instances for different providers.
type LLMClientSubcontext interface {
	// Client operations
	StoreClient(clientID string, client neurotypes.LLMClient)
	GetClient(clientID string) (neurotypes.LLMClient, bool)
	RemoveClient(clientID string)
	GetAllClients() map[string]neurotypes.LLMClient
	ClearAllClients()
}

// llmClientSubcontext implements the LLMClientSubcontext interface.
type llmClientSubcontext struct {
	// LLM client storage
	clients      map[string]neurotypes.LLMClient // Client storage by client ID (provider:hash format)
	clientsMutex sync.RWMutex                    // Protects clients map
}

// NewLLMClientSubcontext creates a new LLMClientSubcontext instance.
func NewLLMClientSubcontext() LLMClientSubcontext {
	return &llmClientSubcontext{
		clients: make(map[string]neurotypes.LLMClient),
	}
}

// NewLLMClientSubcontextFromContext creates an LLMClientSubcontext from an existing NeuroContext.
// This is used by services to get a reference to the context's LLM client subcontext.
func NewLLMClientSubcontextFromContext(ctx *NeuroContext) LLMClientSubcontext {
	return ctx.llmClientCtx
}

// StoreClient stores an LLM client with the given client ID.
func (l *llmClientSubcontext) StoreClient(clientID string, client neurotypes.LLMClient) {
	l.clientsMutex.Lock()
	defer l.clientsMutex.Unlock()

	l.clients[clientID] = client
}

// GetClient retrieves an LLM client by client ID.
// Returns the client and true if found, nil and false otherwise.
func (l *llmClientSubcontext) GetClient(clientID string) (neurotypes.LLMClient, bool) {
	l.clientsMutex.RLock()
	defer l.clientsMutex.RUnlock()

	client, exists := l.clients[clientID]
	return client, exists
}

// RemoveClient removes an LLM client by client ID.
func (l *llmClientSubcontext) RemoveClient(clientID string) {
	l.clientsMutex.Lock()
	defer l.clientsMutex.Unlock()

	delete(l.clients, clientID)
}

// GetAllClients returns a copy of all cached LLM clients (thread-safe read).
func (l *llmClientSubcontext) GetAllClients() map[string]neurotypes.LLMClient {
	l.clientsMutex.RLock()
	defer l.clientsMutex.RUnlock()

	clients := make(map[string]neurotypes.LLMClient)
	for id, client := range l.clients {
		clients[id] = client
	}
	return clients
}

// ClearAllClients removes all cached LLM clients.
func (l *llmClientSubcontext) ClearAllClients() {
	l.clientsMutex.Lock()
	defer l.clientsMutex.Unlock()

	l.clients = make(map[string]neurotypes.LLMClient)
}
