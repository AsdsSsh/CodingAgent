package llm

import (
	"os"
	"strings"
)

// ProviderPreset describes a built-in model provider with auto-detection rules.
type ProviderPreset struct {
	Name          string   // "DeepSeek" / "OpenAI" / "Anthropic"
	APIFormat     string   // "anthropic" / "openai"
	BaseURL       string   // API endpoint
	DefaultModel  string   // Recommended default model for this provider
	ModelPrefixes []string // Prefixes for auto-matching model names
	EnvVar        string   // API key environment variable name
	ContextWindow int
	MaxTokens     int
}

// Built-in provider presets, ordered by priority (first exact match wins).
var presets = []ProviderPreset{
	{
		Name: "Anthropic", APIFormat: "anthropic",
		BaseURL: "https://api.anthropic.com/v1", DefaultModel: "claude-sonnet-4-20250514",
		ModelPrefixes: []string{"claude-"},
		EnvVar:        "ANTHROPIC_API_KEY", ContextWindow: 200_000, MaxTokens: 8192,
	},
	{
		Name: "OpenAI", APIFormat: "openai",
		BaseURL: "https://api.openai.com/v1", DefaultModel: "gpt-4o",
		ModelPrefixes: []string{"gpt-", "o1", "o3", "o4"},
		EnvVar:        "OPENAI_API_KEY", ContextWindow: 128_000, MaxTokens: 4096,
	},
	{
		Name: "DeepSeek", APIFormat: "openai",
		BaseURL: "https://api.deepseek.com/v1", DefaultModel: "deepseek-v4-pro",
		ModelPrefixes: []string{"deepseek-"},
		EnvVar:        "DEEPSEEK_API_KEY", ContextWindow: 1_000_000, MaxTokens: 32768,
	},
	{
		Name: "Ollama", APIFormat: "openai",
		BaseURL: "http://localhost:11434/v1", DefaultModel: "qwen2.5:7b",
		ModelPrefixes: []string{},
		EnvVar:        "", ContextWindow: 32768, MaxTokens: 4096,
	},
	{
		Name: "Gemini", APIFormat: "openai",
		BaseURL: "https://generativelanguage.googleapis.com/v1beta/openai", DefaultModel: "gemini-2.0-flash",
		ModelPrefixes: []string{"gemini-"},
		EnvVar:        "GEMINI_API_KEY", ContextWindow: 1_048_576, MaxTokens: 4096,
	},
	{
		Name: "Moonshot", APIFormat: "openai",
		BaseURL: "https://api.moonshot.cn/v1", DefaultModel: "moonshot-v1-8k",
		ModelPrefixes: []string{"moonshot-", "kimi-"},
		EnvVar:        "MOONSHOT_API_KEY", ContextWindow: 128_000, MaxTokens: 4096,
	},
	{
		Name: "Qwen", APIFormat: "openai",
		BaseURL: "https://dashscope.aliyuncs.com/compatible-mode/v1", DefaultModel: "qwen-max",
		ModelPrefixes: []string{"qwen-"},
		EnvVar:        "DASHSCOPE_API_KEY", ContextWindow: 131_072, MaxTokens: 4096,
	},
	{
		Name: "Zhipu", APIFormat: "openai",
		BaseURL: "https://open.bigmodel.cn/api/paas/v4", DefaultModel: "glm-4",
		ModelPrefixes: []string{"glm-"},
		EnvVar:        "ZHIPU_API_KEY", ContextWindow: 128_000, MaxTokens: 4096,
	},
	{
		Name: "vLLM", APIFormat: "openai",
		BaseURL: "http://localhost:8000/v1", DefaultModel: "",
		ModelPrefixes: []string{},
		EnvVar:        "", ContextWindow: 32768, MaxTokens: 4096,
	},
}

// DetectPreset auto-matches a model name to a built-in preset.
// Rules: exact match on DefaultModel → prefix match on ModelPrefixes → nil.
func DetectPreset(modelName string) *ProviderPreset {
	if modelName == "" {
		return nil
	}
	m := strings.ToLower(modelName)

	// 1. Exact match on default model
	for i := range presets {
		if m == strings.ToLower(presets[i].DefaultModel) {
			return &presets[i]
		}
	}

	// 2. Prefix match
	for i := range presets {
		for _, prefix := range presets[i].ModelPrefixes {
			if strings.HasPrefix(m, strings.ToLower(prefix)) {
				return &presets[i]
			}
		}
	}

	return nil
}

// FindPresetByName finds a preset by its display name.
func FindPresetByName(name string) *ProviderPreset {
	for i := range presets {
		if strings.EqualFold(presets[i].Name, name) {
			return &presets[i]
		}
	}
	return nil
}

// AllPresets returns all registered presets.
func AllPresets() []ProviderPreset {
	return presets
}

// ResolveAPIKey resolves an API key with priority: explicit → preset env var → OPENAI_API_KEY → ANTHROPIC_API_KEY.
func ResolveAPIKey(preset *ProviderPreset, explicitKey string) string {
	if explicitKey != "" {
		return explicitKey
	}
	if preset != nil && preset.EnvVar != "" {
		if val := os.Getenv(preset.EnvVar); val != "" {
			return val
		}
	}
	if val := os.Getenv("OPENAI_API_KEY"); val != "" {
		return val
	}
	return os.Getenv("ANTHROPIC_API_KEY")
}
