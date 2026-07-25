package cli

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/codingagent/coding-agent/internal/agent"
	"github.com/codingagent/coding-agent/internal/config"
	"github.com/codingagent/coding-agent/internal/sandbox"
)

// ChatMessage is a rendered message in the conversation history.
type ChatMessage struct {
	Role    string // "user", "assistant", "tool", "error", "permission"
	Content string
	ToolOk  bool
}

// Model is the top-level bubbletea model for the TUI.
type Model struct {
	// Dimensions
	width  int
	height int

	// Components
	input    textarea.Model
	viewport viewport.Model

	// Data
	config          *config.Config
	messages        []ChatMessage
	agentRunning    bool
	statusLine      string
	focus           Focus
	permissionLevel sandbox.PermissionLevel // current active permission level

	// Agent interaction
	orchestrator   *agent.Orchestrator
	progressChan   chan agent.AgentState
	resultChan     chan agent.AgentResult
	permReqChan    chan agent.PermissionRequest
	permRespChan   chan agent.PermissionResponse
	pendingPermReq *agent.PermissionRequest // nil = no pending request

	// Styles
	styles Styles

	// Spinner counter
	spinnerTick int
}

// NewModel creates a TUI model with the given configuration.
func NewModel(cfg *config.Config) Model {
	ta := textarea.New()
	ta.Placeholder = "Enter your task... (Enter to submit, Ctrl+S for multi-line)"
	ta.ShowLineNumbers = false
	ta.CharLimit = 0
	ta.SetHeight(3)
	ta.Focus()

	vp := viewport.New(80, 20)

	return Model{
		input:           ta,
		viewport:        vp,
		config:          cfg,
		messages:        []ChatMessage{},
		statusLine:      "Ready",
		focus:           FocusInput,
		permissionLevel: sandbox.FromString(cfg.Permissions.DefaultLevel),
		styles:          NewStyles(),
	}
}

// Init initializes the model.
func (m Model) Init() tea.Cmd {
	return textarea.Blink
}
