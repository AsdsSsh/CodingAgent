package llm

// Provider is the abstraction for all LLM backends (Anthropic, OpenAI, local models).
// All methods are synchronous — concurrency is handled by the caller via goroutines.
type Provider interface {
	// ModelName returns a human-readable model identifier.
	ModelName() string

	// ContextWindow returns the total context window size in tokens.
	ContextWindow() int

	// Chat sends messages to the LLM and returns text or tool calls.
	// Tools are passed as provider-agnostic maps; each provider converts to its own API format.
	Chat(messages []Message, tools []map[string]any) (LlmResponse, error)

	// CountTokens returns a rough token count for a string.
	CountTokens(text string) int

	// CountMessagesTokens returns a rough token count for a full message list.
	CountMessagesTokens(messages []Message) int

	// Close releases any resources held by the provider.
	Close() error
}

// BaseProvider provides a default implementation of CountMessagesTokens
// that providers can embed.
type BaseProvider struct{}

// CountMessagesTokens estimates the token count for a list of messages.
func (BaseProvider) CountMessagesTokens(messages []Message, counter func(string) int) int {
	total := 0
	for _, msg := range messages {
		if msg.Content != "" {
			total += counter(msg.Content)
		}
		for _, tc := range msg.ToolCalls {
			total += counter(tc.Name)
			if tc.Arguments != nil {
				for k, v := range tc.Arguments {
					total += counter(k)
					if s, ok := v.(string); ok {
						total += counter(s)
					}
				}
			}
		}
	}
	return total
}
