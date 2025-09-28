// Package context provides an optimized LRU cache implementation specifically tailored
// for NeuroShell variable storage. This cache is designed to handle hundreds of thousands
// of variables efficiently while maintaining thread safety and supporting pinned entries.
package context

import (
	"sync"
)

// VariableLRUCache is an optimized LRU cache specifically for variable storage.
// It provides thread-safe operations, pinned entries for frequently accessed variables,
// and efficient memory usage for variable data.
type VariableLRUCache struct {
	maxSize    int
	cache      map[string]*cacheNode
	head       *cacheNode
	tail       *cacheNode
	pinnedVars map[string]bool // Variables that bypass LRU eviction
	mutex      sync.RWMutex
}

// cacheNode represents a node in the doubly-linked list used by the LRU cache.
type cacheNode struct {
	key    string
	value  string
	prev   *cacheNode
	next   *cacheNode
	pinned bool // Whether this node is pinned (won't be evicted)
}

// NewVariableLRUCache creates a new LRU cache with the specified maximum size.
// Frequently accessed variables like _output, _error, etc., are automatically pinned.
func NewVariableLRUCache(maxSize int) *VariableLRUCache {
	if maxSize <= 0 {
		maxSize = 10000 // Default cache size
	}

	// Create sentinel nodes for head and tail
	head := &cacheNode{}
	tail := &cacheNode{}
	head.next = tail
	tail.prev = head

	cache := &VariableLRUCache{
		maxSize:    maxSize,
		cache:      make(map[string]*cacheNode),
		head:       head,
		tail:       tail,
		pinnedVars: make(map[string]bool),
		mutex:      sync.RWMutex{},
	}

	// Pin frequently accessed variables that should bypass LRU eviction
	pinnedVariables := []string{
		"_output",
		"_error",
		"_status",
		"_elapsed",
		"_style",
		"_default_command",
		"_completion_mode",
	}

	for _, varName := range pinnedVariables {
		cache.pinnedVars[varName] = true
	}

	return cache
}

// Get retrieves a value from the cache and marks it as recently used.
// Returns the value and true if found, empty string and false if not found.
func (c *VariableLRUCache) Get(key string) (string, bool) {
	c.mutex.RLock()
	node, exists := c.cache[key]
	c.mutex.RUnlock()

	if !exists {
		return "", false
	}

	// Move to head (most recently used) if not pinned
	if !node.pinned {
		c.mutex.Lock()
		c.moveToHead(node)
		c.mutex.Unlock()
	}

	return node.value, true
}

// Set adds or updates a key-value pair in the cache.
// If the key is in the pinned list, it will be marked as pinned.
func (c *VariableLRUCache) Set(key, value string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if node, exists := c.cache[key]; exists {
		// Update existing node
		node.value = value
		if !node.pinned {
			c.moveToHead(node)
		}
		return
	}

	// Create new node
	isPinned := c.pinnedVars[key]
	newNode := &cacheNode{
		key:    key,
		value:  value,
		pinned: isPinned,
	}

	c.cache[key] = newNode
	c.addToHead(newNode)

	// Check if we need to evict (only if we exceeded capacity with non-pinned items)
	if len(c.cache) > c.maxSize {
		c.evictLRU()
	}
}

// Delete removes a key from the cache.
func (c *VariableLRUCache) Delete(key string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if node, exists := c.cache[key]; exists {
		c.removeNode(node)
		delete(c.cache, key)
	}
}

// GetAll returns all key-value pairs in the cache.
// This is used for implementing GetAllVariables functionality.
func (c *VariableLRUCache) GetAll() map[string]string {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	result := make(map[string]string, len(c.cache))
	for key, node := range c.cache {
		result[key] = node.value
	}
	return result
}

// Size returns the current number of items in the cache.
func (c *VariableLRUCache) Size() int {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return len(c.cache)
}

// Clear removes all items from the cache.
func (c *VariableLRUCache) Clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.cache = make(map[string]*cacheNode)
	c.head.next = c.tail
	c.tail.prev = c.head
}

// SetPinned marks a variable as pinned, preventing it from being evicted.
// This is useful for frequently accessed variables.
func (c *VariableLRUCache) SetPinned(key string, pinned bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.pinnedVars[key] = pinned

	// Update existing node if it exists
	if node, exists := c.cache[key]; exists {
		node.pinned = pinned
	}
}

// moveToHead moves a node to the head of the doubly-linked list.
// Must be called with mutex locked.
func (c *VariableLRUCache) moveToHead(node *cacheNode) {
	c.removeNode(node)
	c.addToHead(node)
}

// addToHead adds a node right after the head sentinel.
// Must be called with mutex locked.
func (c *VariableLRUCache) addToHead(node *cacheNode) {
	node.prev = c.head
	node.next = c.head.next
	c.head.next.prev = node
	c.head.next = node
}

// removeNode removes a node from the doubly-linked list.
// Must be called with mutex locked.
func (c *VariableLRUCache) removeNode(node *cacheNode) {
	node.prev.next = node.next
	node.next.prev = node.prev
}

// evictLRU removes the least recently used non-pinned item from the cache.
// Must be called with mutex locked.
func (c *VariableLRUCache) evictLRU() {
	// Start from the tail and find the first non-pinned node
	current := c.tail.prev
	for current != c.head {
		if !current.pinned {
			// Found a non-pinned node to evict
			c.removeNode(current)
			delete(c.cache, current.key)
			return
		}
		current = current.prev
	}

	// If we reach here, all items are pinned.
	// This is unusual but we handle it gracefully by not evicting anything.
	// The cache may temporarily exceed maxSize in this case.
}

// GetStats returns cache statistics for monitoring and debugging.
func (c *VariableLRUCache) GetStats() CacheStats {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	pinnedCount := 0
	for _, node := range c.cache {
		if node.pinned {
			pinnedCount++
		}
	}

	return CacheStats{
		Size:        len(c.cache),
		MaxSize:     c.maxSize,
		PinnedCount: pinnedCount,
	}
}

// CacheStats provides information about cache performance and usage.
type CacheStats struct {
	Size        int // Current number of items
	MaxSize     int // Maximum cache size
	PinnedCount int // Number of pinned items
}