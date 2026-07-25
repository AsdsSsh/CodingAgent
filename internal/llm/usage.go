package llm

// TokenUsage tracks token consumption for a single LLM API call.
type TokenUsage struct {
	InputTokens  int
	OutputTokens int
}

// Total returns the sum of input and output tokens.
func (tu TokenUsage) Total() int {
	return tu.InputTokens + tu.OutputTokens
}
