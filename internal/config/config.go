package config

import (
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config is the root configuration object with layered loading:
// hardcoded defaults → global (~/.coding-agent/config.yaml) → project (.coding-agent/config.yaml) → env vars → CLI flags
type Config struct {
	Model        ModelConfig     `yaml:"model"`
	MCPServers   []MCPServerConfig `yaml:"mcp_servers"`
	Permissions  PermissionConfig `yaml:"permissions"`
	Workspace    string          `yaml:"workspace"`
	MaxIter      int             `yaml:"max_iterations"`
	MemoryPath   string          `yaml:"memory_path"`
	Sandbox      SandboxConfig   `yaml:"sandbox"`
}

// ModelConfig is an immutable-ish model configuration.
type ModelConfig struct {
	Provider      string   `yaml:"provider"`
	Model         string   `yaml:"model"`
	APIKey        string   `yaml:"api_key"`
	APIKeyEnv     string   `yaml:"api_key_env"`
	BaseURL       string   `yaml:"base_url"`
	MaxTokens     int      `yaml:"max_tokens"`
	ContextWindow int      `yaml:"context_window"`
	Temperature   *float64 `yaml:"temperature"`
}

// PermissionConfig is the permission policy configuration.
type PermissionConfig struct {
	DefaultLevel          string          `yaml:"default_level"`
	AutoApprove           map[string]bool `yaml:"auto_approve"`
	MaxFrequency          int             `yaml:"max_frequency"`
	FrequencyWindowSeconds int            `yaml:"frequency_window_seconds"`
}

// SandboxConfig is the sandbox execution configuration.
type SandboxConfig struct {
	Type            string   `yaml:"type"`
	TimeoutSeconds  int      `yaml:"timeout_seconds"`
	BlockedCommands []string `yaml:"blocked_commands"`
}

// MCPServerConfig is configuration for a single MCP server.
type MCPServerConfig struct {
	Name      string            `yaml:"name"`
	Transport string            `yaml:"transport"`
	Command   string            `yaml:"command"`
	Args      []string          `yaml:"args"`
	URL       string            `yaml:"url"`
	Env       map[string]string `yaml:"env"`
}

// Defaults returns the hardcoded default configuration.
func Defaults() *Config {
	home, _ := os.UserHomeDir()
	return &Config{
		Model: ModelConfig{
			Provider:      "anthropic",
			Model:         "claude-sonnet-4-20250514",
			MaxTokens:     8192,
			ContextWindow: 200_000,
		},
		MCPServers: []MCPServerConfig{},
		Permissions: PermissionConfig{
			DefaultLevel:           "READ",
			MaxFrequency:           10,
			FrequencyWindowSeconds: 300,
		},
		Workspace:  "",
		MaxIter:    50,
		MemoryPath: filepath.Join(home, ".coding-agent", "memory.jsonl"),
		Sandbox: SandboxConfig{
			Type:           "subprocess",
			TimeoutSeconds: 120,
			BlockedCommands: []string{
				"rm -rf /", "sudo", "dd if=", "mkfs",
				":(){ :|:& };:", "shutdown", "reboot",
			},
		},
	}
}

// Load loads config with layered merging: global → project → env vars.
func Load() (*Config, error) {
	cfg := Defaults()

	// Layer 1: global config
	home, err := os.UserHomeDir()
	if err == nil {
		globalPath := filepath.Join(home, ".coding-agent", "config.yaml")
		if data, err := os.ReadFile(globalPath); err == nil {
			var globalCfg Config
			if err := yaml.Unmarshal(data, &globalCfg); err == nil {
				cfg.merge(&globalCfg)
			}
		}
	}

	// Layer 2: project config
	if data, err := os.ReadFile(filepath.Join(".coding-agent", "config.yaml")); err == nil {
		var projectCfg Config
		if err := yaml.Unmarshal(data, &projectCfg); err == nil {
			cfg.merge(&projectCfg)
		}
	}

	// Layer 3: env vars
	cfg.ApplyEnv()

	return cfg, nil
}

// LoadFromPath loads config from a specific YAML file path.
func LoadFromPath(configPath string) (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return Defaults(), nil
	}
	var fileCfg Config
	if err := yaml.Unmarshal(data, &fileCfg); err != nil {
		return nil, err
	}
	cfg := Defaults()
	cfg.merge(&fileCfg)
	return cfg, nil
}

// merge overlays non-zero fields from override onto base.
func (c *Config) merge(override *Config) {
	if override == nil {
		return
	}

	// Model: merge non-zero fields
	if override.Model.Provider != "" {
		c.Model.Provider = override.Model.Provider
	}
	if override.Model.Model != "" {
		c.Model.Model = override.Model.Model
	}
	if override.Model.APIKey != "" {
		c.Model.APIKey = override.Model.APIKey
	}
	if override.Model.APIKeyEnv != "" {
		c.Model.APIKeyEnv = override.Model.APIKeyEnv
	}
	if override.Model.BaseURL != "" {
		c.Model.BaseURL = override.Model.BaseURL
	}
	if override.Model.MaxTokens > 0 {
		c.Model.MaxTokens = override.Model.MaxTokens
	}
	if override.Model.ContextWindow > 0 {
		c.Model.ContextWindow = override.Model.ContextWindow
	}
	if override.Model.Temperature != nil {
		c.Model.Temperature = override.Model.Temperature
	}

	// MCP servers: replace if non-empty
	if len(override.MCPServers) > 0 {
		c.MCPServers = override.MCPServers
	}

	// Permissions
	if override.Permissions.DefaultLevel != "" {
		c.Permissions.DefaultLevel = override.Permissions.DefaultLevel
	}
	if override.Permissions.MaxFrequency > 0 {
		c.Permissions.MaxFrequency = override.Permissions.MaxFrequency
	}
	if override.Permissions.FrequencyWindowSeconds > 0 {
		c.Permissions.FrequencyWindowSeconds = override.Permissions.FrequencyWindowSeconds
	}
	if override.Permissions.AutoApprove != nil {
		c.Permissions.AutoApprove = override.Permissions.AutoApprove
	}

	// Workspace
	if override.Workspace != "" {
		c.Workspace = override.Workspace
	}

	// MaxIter
	if override.MaxIter > 0 {
		c.MaxIter = override.MaxIter
	}

	// MemoryPath
	if override.MemoryPath != "" {
		c.MemoryPath = override.MemoryPath
	}

	// Sandbox
	if override.Sandbox.Type != "" {
		c.Sandbox.Type = override.Sandbox.Type
	}
	if override.Sandbox.TimeoutSeconds > 0 {
		c.Sandbox.TimeoutSeconds = override.Sandbox.TimeoutSeconds
	}
	if len(override.Sandbox.BlockedCommands) > 0 {
		c.Sandbox.BlockedCommands = override.Sandbox.BlockedCommands
	}
}

// ApplyEnv applies environment variable overrides.
func (c *Config) ApplyEnv() {
	// ANTHROPIC_API_KEY env var
	if c.Model.Provider == "anthropic" || c.Model.APIKey == "" {
		if key := os.Getenv("ANTHROPIC_API_KEY"); key != "" {
			c.Model.APIKey = key
		}
	}
}

// ResolveAPIKey resolves the effective API key with priority:
// explicit key → api_key_env → provider default env var.
func (c *Config) ResolveAPIKey() string {
	if c.Model.APIKey != "" {
		return c.Model.APIKey
	}

	// User-specified env var name
	if c.Model.APIKeyEnv != "" {
		if val := os.Getenv(c.Model.APIKeyEnv); val != "" {
			return val
		}
	}

	// Provider default env var
	prov := strings.ToLower(c.Model.Provider)
	switch prov {
	case "anthropic":
		return os.Getenv("ANTHROPIC_API_KEY")
	case "openai":
		return os.Getenv("OPENAI_API_KEY")
	default:
		if key := os.Getenv("OPENAI_API_KEY"); key != "" {
			return key
		}
		return os.Getenv("ANTHROPIC_API_KEY")
	}
}

// CLI override methods

// WithModel overrides the model name in config.
func (c *Config) WithModel(model string) *Config {
	c.Model.Model = model
	return c
}

// WithProvider overrides the provider in config.
func (c *Config) WithProvider(provider string) *Config {
	c.Model.Provider = provider
	return c
}

// WithAPIKey overrides the API key.
func (c *Config) WithAPIKey(key string) *Config {
	c.Model.APIKey = key
	return c
}

// WithBaseURL overrides the base URL.
func (c *Config) WithBaseURL(url string) *Config {
	c.Model.BaseURL = url
	return c
}

// WithWorkspace overrides the workspace path.
func (c *Config) WithWorkspace(ws string) *Config {
	c.Workspace = ws
	return c
}

// WithPermissionLevel overrides the default permission level.
func (c *Config) WithPermissionLevel(level string) *Config {
	c.Permissions.DefaultLevel = level
	return c
}

// WithModelConfig replaces the entire model config.
func (c *Config) WithModelConfig(mc ModelConfig) *Config {
	c.Model = mc
	return c
}
