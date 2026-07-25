package tool

import "sync"

// ToolRegistry is the central registry for all available tools.
// Thread-safe via sync.RWMutex.
type ToolRegistry struct {
	mu    sync.RWMutex
	tools map[string]Tool
}

// NewToolRegistry creates an empty tool registry.
func NewToolRegistry() *ToolRegistry {
	return &ToolRegistry{
		tools: make(map[string]Tool),
	}
}

// Register adds a tool to the registry.
func (tr *ToolRegistry) Register(t Tool) {
	tr.mu.Lock()
	defer tr.mu.Unlock()
	tr.tools[t.Name()] = t
}

// RegisterAll adds multiple tools at once.
func (tr *ToolRegistry) RegisterAll(tools []Tool) {
	tr.mu.Lock()
	defer tr.mu.Unlock()
	for _, t := range tools {
		tr.tools[t.Name()] = t
	}
}

// Resolve returns a tool by name. Returns nil if not found — caller should check with Has first.
func (tr *ToolRegistry) Resolve(name string) Tool {
	tr.mu.RLock()
	defer tr.mu.RUnlock()
	return tr.tools[name]
}

// Has checks whether a tool is registered.
func (tr *ToolRegistry) Has(name string) bool {
	tr.mu.RLock()
	defer tr.mu.RUnlock()
	_, ok := tr.tools[name]
	return ok
}

// ToLLMFormats exports all tools as provider-agnostic LLM tool definitions.
func (tr *ToolRegistry) ToLLMFormats() []map[string]any {
	tr.mu.RLock()
	defer tr.mu.RUnlock()
	formats := make([]map[string]any, 0, len(tr.tools))
	for _, t := range tr.tools {
		formats = append(formats, t.ToLLMFormat())
	}
	return formats
}

// Size returns the number of registered tools.
func (tr *ToolRegistry) Size() int {
	tr.mu.RLock()
	defer tr.mu.RUnlock()
	return len(tr.tools)
}
