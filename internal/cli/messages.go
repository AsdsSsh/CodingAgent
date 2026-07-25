package cli

import "github.com/codingagent/coding-agent/internal/agent"

// ─── User input ───

// TaskSubmittedMsg is sent when the user submits a task via Ctrl+Enter.
type TaskSubmittedMsg string

// ─── Agent progress ───

// AgentProgressMsg carries a real-time agent state update.
type AgentProgressMsg struct {
	State agent.AgentState
}

// AgentResultMsg carries the final result of an agent run.
type AgentResultMsg struct {
	Result agent.AgentResult
}

// ─── Permission prompts ───

// PermissionPromptMsg triggers a blocking permission modal.
type PermissionPromptMsg struct {
	Request agent.PermissionRequest
}

// PermissionResponseMsg carries the user's response to a permission prompt.
type PermissionResponseMsg struct {
	Response agent.PermissionResponse
}

// ─── UI navigation ───

// FocusChangedMsg is sent when focus toggles between input and viewport.
type FocusChangedMsg struct {
	Focus Focus
}

// Focus represents which UI element has keyboard focus.
type Focus int

const (
	FocusInput    Focus = iota
	FocusViewport
)

// ─── Slash commands ───

// CommandMsg is sent when a slash command is executed.
type CommandMsg struct {
	Command string
	Args    string
}

// ─── Tick ───

// TickMsg is sent periodically for spinner animation.
type TickMsg struct{}
