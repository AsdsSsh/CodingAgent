package mcp

import (
	"fmt"
	"log"
	"sync"

	"github.com/codingagent/coding-agent/internal/config"
	"github.com/codingagent/coding-agent/internal/tool"
)

// LazyRegistry manages MCP tools with deferred connection.
// Tool metadata is cached at startup; actual server connections are deferred to first invocation.
type LazyRegistry struct {
	mu      sync.Mutex
	clients map[string]*Client     // server name → MCP client
	tools   []tool.Tool            // registered MCP tool adapters
	configs []config.MCPServerConfig
}

// NewLazyRegistry creates a lazy MCP tool registry from server configs.
func NewLazyRegistry(cfgs []config.MCPServerConfig) *LazyRegistry {
	return &LazyRegistry{
		clients: make(map[string]*Client),
		configs: cfgs,
	}
}

// Preload connects to all MCP servers and registers their tools.
// Connection failures are logged as warnings, not fatal errors.
func (lr *LazyRegistry) Preload(existing map[string]bool) []tool.Tool {
	lr.mu.Lock()
	defer lr.mu.Unlock()

	var allTools []tool.Tool

	for _, cfg := range lr.configs {
		client, err := NewClient(cfg)
		if err != nil {
			log.Printf("MCP warning: failed to create client for %s: %v", cfg.Name, err)
			continue
		}

		tools, err := client.ListTools()
		if err != nil {
			log.Printf("MCP warning: failed to list tools for %s: %v", cfg.Name, err)
			continue
		}

		lr.clients[cfg.Name] = client

		for _, td := range tools {
			// Skip duplicate tool names
			if existing[td.Name] {
				log.Printf("MCP warning: skipping duplicate tool '%s' from server '%s'", td.Name, cfg.Name)
				continue
			}
			adapter := NewToolAdapter(client, td)
			lr.tools = append(lr.tools, adapter)
			allTools = append(allTools, adapter)
		}
	}

	return allTools
}

// EnsureConnected checks if an MCP server is connected and connects if not (lazy).
func (lr *LazyRegistry) EnsureConnected(serverName string) error {
	lr.mu.Lock()
	defer lr.mu.Unlock()

	if _, ok := lr.clients[serverName]; ok {
		return nil // Already connected
	}

	// Find the config and connect
	for _, cfg := range lr.configs {
		if cfg.Name == serverName {
			client, err := NewClient(cfg)
			if err != nil {
				return fmt.Errorf("MCP connect to %s failed: %w", serverName, err)
			}

			tools, err := client.ListTools()
			if err != nil {
				return fmt.Errorf("MCP list tools for %s failed: %w", serverName, err)
			}

			lr.clients[serverName] = client

			for _, td := range tools {
				adapter := NewToolAdapter(client, td)
				lr.tools = append(lr.tools, adapter)
			}
			return nil
		}
	}

	return fmt.Errorf("MCP server '%s' not found in configuration", serverName)
}

// Close terminates all MCP connections.
func (lr *LazyRegistry) Close() {
	lr.mu.Lock()
	defer lr.mu.Unlock()

	for name, client := range lr.clients {
		if err := client.Close(); err != nil {
			log.Printf("MCP warning: error closing client %s: %v", name, err)
		}
	}
	lr.clients = make(map[string]*Client)
	lr.tools = nil
}
