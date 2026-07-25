package llm

import (
	"fmt"
	"strings"
)

// ProviderFactory creates LLM providers from configuration, with auto-detection.
type ProviderFactory struct {
	Model         string
	Provider      string
	APIKey        string
	APIKeyEnv     string
	BaseURL       string
	MaxTokens     int
	ContextWindow int
	Temperature   *float64
}

// NewProviderFactory creates a factory from model configuration parameters.
func NewProviderFactory(model, provider, apiKey, apiKeyEnv, baseURL string, maxTokens, contextWindow int, temperature *float64) *ProviderFactory {
	return &ProviderFactory{
		Model:         model,
		Provider:      provider,
		APIKey:        apiKey,
		APIKeyEnv:     apiKeyEnv,
		BaseURL:       baseURL,
		MaxTokens:     maxTokens,
		ContextWindow: contextWindow,
		Temperature:   temperature,
	}
}

// Create builds the appropriate Provider based on configuration.
// Detection priority:
//  1. Explicit base_url → OpenAI-compatible format with that URL
//  2. Model name matches a preset → use preset's base_url and API format
//  3. Provider explicitly set to "anthropic" → AnthropicProvider
//  4. None matches → error with suggestions
func (f *ProviderFactory) Create() (Provider, error) {
	preset := DetectPreset(f.Model)

	// 1. Explicit base_url takes priority
	if f.BaseURL != "" {
		return f.createOpenAI(f.BaseURL)
	}

	// 2. Model matched a preset
	if preset != nil {
		apiKey := ResolveAPIKey(preset, f.apiKey())
		if apiKey == "" {
			return nil, fmt.Errorf("no API key found for %s. Set %s environment variable or api_key in config", preset.Name, preset.EnvVar)
		}
		switch preset.APIFormat {
		case "anthropic":
			maxTok := f.maxTokens()
			ctxWin := f.contextWindow()
			return NewAnthropicProvider(apiKey, f.Model, maxTok, ctxWin), nil
		default:
			maxTok := f.maxTokens()
			ctxWin := f.contextWindow()
			return NewOpenAIProvider(apiKey, f.Model, maxTok, ctxWin, preset.BaseURL, f.Temperature), nil
		}
	}

	// 3. Provider field explicitly says anthropic
	if strings.EqualFold(f.Provider, "anthropic") {
		apiKey := ResolveAPIKey(nil, f.apiKey())
		if apiKey == "" {
			return nil, fmt.Errorf("no Anthropic API key found. Set ANTHROPIC_API_KEY environment variable or api_key in config")
		}
		maxTok := f.maxTokens()
		ctxWin := f.contextWindow()
		return NewAnthropicProvider(apiKey, f.Model, maxTok, ctxWin), nil
	}

	// 4. Cannot determine
	names := make([]string, len(presets))
	for i, p := range presets {
		names[i] = p.Name
	}
	return nil, fmt.Errorf(
		"cannot identify model '%s'.\n  Available presets: %s\n  Or set base_url in config for custom endpoints",
		f.Model, strings.Join(names, ", "),
	)
}

func (f *ProviderFactory) apiKey() string {
	if f.APIKey != "" {
		return f.APIKey
	}
	if f.APIKeyEnv != "" {
		return f.APIKeyEnv
	}
	return ""
}

func (f *ProviderFactory) maxTokens() int {
	if f.MaxTokens > 0 {
		return f.MaxTokens
	}
	return 8192
}

func (f *ProviderFactory) contextWindow() int {
	if f.ContextWindow > 0 {
		return f.ContextWindow
	}
	return 200_000
}

func (f *ProviderFactory) createOpenAI(baseURL string) (Provider, error) {
	apiKey := ResolveAPIKey(nil, f.apiKey())
	if apiKey == "" {
		return nil, fmt.Errorf("no API key found. Set OPENAI_API_KEY or ANTHROPIC_API_KEY environment variable")
	}
	maxTok := f.maxTokens()
	ctxWin := f.contextWindow()
	return NewOpenAIProvider(apiKey, f.Model, maxTok, ctxWin, baseURL, f.Temperature), nil
}
