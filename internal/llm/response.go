package llm

// LlmResponse is the parsed LLM response. Either content (text) or tool calls, mutually exclusive.
// ReasoningContent is used for DeepSeek thinking mode passthrough.
type LlmResponse struct {
	Content          string
	ToolCalls        []ToolCall
	IsFinalAnswer    bool
	Usage            TokenUsage
	StopReason       string
	ReasoningContent string
}

// HasToolCalls returns true if the response contains tool call requests.
func (r LlmResponse) HasToolCalls() bool {
	return len(r.ToolCalls) > 0
}
