package agent

// AgentState is an immutable snapshot of agent progress at a single step.
// Emitted via EventBus for real-time display.
type AgentState struct {
	Task               string
	Step               int
	ConversationTokens int
	CurrentTool        string
	PermissionLevel    string
}

// AgentResult is the final result of an agent run.
type AgentResult struct {
	Answer       string
	Steps        int
	ToolCalls    []ToolCallRecord
	Conversation []MessageRecord
	Error        error
}

// ToolCallRecord logs a single tool invocation during the run.
type ToolCallRecord struct {
	Name    string
	Args    map[string]any
	Success bool
}

// MessageRecord is a lightweight conversation entry for the UI.
type MessageRecord struct {
	Role    string
	Content string
}
