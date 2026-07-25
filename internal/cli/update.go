package cli

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/codingagent/coding-agent/internal/agent"
	"github.com/codingagent/coding-agent/internal/config"
)

// Update handles all bubbletea messages and returns the updated model and commands.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {

	// ─── Lifecycle ───
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.viewport.Width = msg.Width - 4
		m.viewport.Height = msg.Height - 8
		m.input.SetWidth(msg.Width - 4)

	// ─── Keyboard ───
	case tea.KeyMsg:
		// If permission modal is showing, intercept all keys
		if m.showPermModal {
			switch msg.String() {
			case "y", "Y", "enter":
				m.showPermModal = false
				m.pendingPermReq = nil
				select {
				case m.permRespChan <- agent.PermissionResponse{Allowed: true}:
				default:
				}
				return m, m.listenProgress()
			case "n", "N", "esc":
				m.showPermModal = false
				m.pendingPermReq = nil
				select {
				case m.permRespChan <- agent.PermissionResponse{Allowed: false}:
				default:
				}
				return m, m.listenProgress()
			case "a", "A":
				m.showPermModal = false
				m.pendingPermReq = nil
				select {
				case m.permRespChan <- agent.PermissionResponse{Allowed: true, AlwaysAllow: true}:
				default:
				}
				return m, m.listenProgress()
			}
			return m, nil // block all other keys when modal is showing
		}

		switch msg.String() {
		case "ctrl+c":
			if m.agentRunning {
				m.agentRunning = false
				m.messages = append(m.messages, ChatMessage{Role: "error", Content: "Task cancelled"})
				m.statusLine = "Cancelled"
				return m, nil
			}
			return m, tea.Quit

		case "ctrl+s":
			// Ctrl+S: primary submit (works cross-platform)
			if m.focus == FocusInput && !m.agentRunning {
				task := strings.TrimSpace(m.input.Value())
				if task != "" {
					m.messages = append(m.messages, ChatMessage{Role: "user", Content: task})
					m.input.Reset()
					return m.handleTaskSubmit(task)
				}
			}

		case "enter":
			// Enter: submit if the input only has one line (not multi-line)
			if m.focus == FocusInput && !m.agentRunning {
				task := strings.TrimSpace(m.input.Value())
				if strings.HasPrefix(task, "/") {
					// Slash command on single line
					cmd, args := parseSlashCommand(task)
					if cmd != "" {
						m.input.Reset()
						return m, m.handleSlashCommand(cmd, args)
					}
				}
				if task != "" && !strings.Contains(task, "\n") {
					m.messages = append(m.messages, ChatMessage{Role: "user", Content: task})
					m.input.Reset()
					return m.handleTaskSubmit(task)
				}
			}
			// Multi-line: let textarea handle Enter for newline (falls through to sub-component update)

		case "tab":
			if m.focus == FocusInput {
				m.focus = FocusViewport
				m.input.Blur()
			} else {
				m.focus = FocusInput
				m.input.Focus()
			}
			return m, nil
		}

	// ─── Agent Progress ───
	case AgentProgressMsg:
		state := msg.State
		m.statusLine = fmt.Sprintf("Step %d · %s · %d tokens",
			state.Step, state.CurrentTool, state.ConversationTokens)
		if state.CurrentTool != "" {
			m.messages = append(m.messages, ChatMessage{
				Role:   "tool",
				Content: state.CurrentTool,
				ToolOk: true,
			})
		}
		// Continue listening for more progress
		cmds = append(cmds, m.listenProgress())

	// ─── Agent Result ───
	case AgentResultMsg:
		m.agentRunning = false
		result := msg.Result
		if result.Error != nil {
			m.messages = append(m.messages, ChatMessage{Role: "error", Content: result.Error.Error()})
		} else {
			m.messages = append(m.messages, ChatMessage{Role: "assistant", Content: result.Answer})
		}
		m.statusLine = fmt.Sprintf("Done — %d steps, %d tools", result.Steps, len(result.ToolCalls))

	// ─── Permission Prompt ───
	case PermissionPromptMsg:
		m.showPermModal = true
		m.pendingPermReq = &msg.Request

	case PermissionResponseMsg:
		if m.permRespChan != nil {
			select {
			case m.permRespChan <- msg.Response:
			default:
			}
		}
		m.showPermModal = false
		m.pendingPermReq = nil

	// ─── Tick (spinner) ───
	case TickMsg:
		if m.agentRunning {
			m.spinnerTick++
			cmds = append(cmds, m.tickLater())
		}
	}

	// ─── Sub-component updates ───
	if m.focus == FocusInput && !m.agentRunning {
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		cmds = append(cmds, cmd)
	}

	var vpCmd tea.Cmd
	m.viewport, vpCmd = m.viewport.Update(msg)
	cmds = append(cmds, vpCmd)

	return m, tea.Batch(cmds...)
}

// handleTaskSubmit starts an agent run in a goroutine and wires up progress/result channels.
func (m Model) handleTaskSubmit(task string) (tea.Model, tea.Cmd) {
	// Create orchestrator for this task
	orch, err := agent.NewOrchestrator(m.config)
	if err != nil {
		m.messages = append(m.messages,
			ChatMessage{Role: "error", Content: "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"},
			ChatMessage{Role: "error", Content: "⚠️  INITIALIZATION FAILED"},
			ChatMessage{Role: "error", Content: err.Error()},
			ChatMessage{Role: "error", Content: "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"},
			ChatMessage{Role: "assistant", Content: "Please set your API key:\n  export DEEPSEEK_API_KEY=\"sk-...\"\nOr edit .coding-agent/config.yaml"},
		)
		return m, nil
	}

	m.orchestrator = orch
	m.agentRunning = true
	m.taskInFlight = task
	m.statusLine = "Starting..."
	m.showPermModal = false

	// Create channels
	m.progressChan = make(chan agent.AgentState, 20)
	m.resultChan = make(chan agent.AgentResult, 1)
	m.permReqChan = make(chan agent.PermissionRequest, 5)
	m.permRespChan = make(chan agent.PermissionResponse, 5)

	// Run agent in goroutine
	go func() {
		result := orch.RunWithChannels(task, m.progressChan, m.permReqChan, m.permRespChan)
		// Signal completion by closing channels
		close(m.progressChan)
		m.resultChan <- result
	}()

	return m, tea.Batch(
		m.listenProgress(),
		m.waitForResult(),
		m.listenPermission(),
		m.tickLater(),
	)
}

// listenProgress returns a command that waits for the next progress update.
func (m Model) listenProgress() tea.Cmd {
	return func() tea.Msg {
		state, ok := <-m.progressChan
		if !ok {
			return nil // Channel closed
		}
		return AgentProgressMsg{State: state}
	}
}

// waitForResult returns a command that waits for the final agent result.
func (m Model) waitForResult() tea.Cmd {
	return func() tea.Msg {
		result, ok := <-m.resultChan
		if !ok {
			return AgentResultMsg{
				Result: agent.AgentResult{Error: fmt.Errorf("agent channel closed unexpectedly")},
			}
		}
		return AgentResultMsg{Result: result}
	}
}

// listenPermission returns a command that waits for permission requests from the agent.
func (m Model) listenPermission() tea.Cmd {
	return func() tea.Msg {
		req, ok := <-m.permReqChan
		if !ok {
			return nil
		}
		return PermissionPromptMsg{Request: req}
	}
}

// tickLater schedules a TickMsg after a short delay for spinner animation.
func (m Model) tickLater() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		return TickMsg{}
	})
}

// ─── Slash command handling ───

// handleSlashCommand processes a slash command and returns updated model.
func (m Model) handleSlashCommand(cmd, args string) tea.Cmd {
	switch cmd {
	case "help", "h":
		helpText := `Available commands:
  /model [name]     View or switch model
  /permission [lvl] View or switch permission level (READ/WRITE/EXECUTE/DANGEROUS)
  /clear            Clear conversation history
  /exit             Exit CodingAgent`
		m.messages = append(m.messages, ChatMessage{Role: "assistant", Content: helpText})
		return nil

	case "model", "m":
		if args == "" {
			m.messages = append(m.messages, ChatMessage{
				Role:    "assistant",
				Content: fmt.Sprintf("Current model: %s\nUse /model <name> to switch.", m.config.Model.Model),
			})
		} else {
			m.config.WithModel(args)
			m.messages = append(m.messages, ChatMessage{
				Role:    "assistant",
				Content: fmt.Sprintf("Model switched to: %s", args),
			})
		}
		return nil

	case "permission", "perm", "p":
		if args == "" {
			m.messages = append(m.messages, ChatMessage{
				Role:    "assistant",
				Content: fmt.Sprintf("Current permission: %s\nUse /permission <READ|WRITE|EXECUTE|DANGEROUS>.", m.config.Permissions.DefaultLevel),
			})
		} else {
			level := strings.ToUpper(args)
			m.config.WithPermissionLevel(level)
			m.messages = append(m.messages, ChatMessage{
				Role:    "assistant",
				Content: fmt.Sprintf("Permission switched to: %s", level),
			})
		}
		return nil

	case "clear":
		m.messages = []ChatMessage{}
		m.statusLine = "Conversation cleared"
		return nil

	case "exit", "quit", "q":
		return tea.Quit

	default:
		m.messages = append(m.messages, ChatMessage{
			Role:    "error",
			Content: fmt.Sprintf("Unknown command: %s. Type /help for available commands.", cmd),
		})
		return nil
	}
}

// parseSlashCommand extracts the command and args from a slash-prefixed string.
func parseSlashCommand(input string) (cmd, args string) {
	input = strings.TrimSpace(input)
	if !strings.HasPrefix(input, "/") {
		return "", ""
	}
	parts := strings.SplitN(input[1:], " ", 2)
	cmd = strings.ToLower(parts[0])
	if len(parts) > 1 {
		args = strings.TrimSpace(parts[1])
	}
	return
}

// NewConfigFromCLI creates a config from CLI-relevant parameters.
func NewConfigFromCLI(provider, model, apiKey, baseURL, workspace string) (*config.Config, error) {
	cfg, err := config.Load()
	if err != nil {
		cfg = config.Defaults()
	}
	if workspace != "" {
		cfg.WithWorkspace(workspace)
	}
	if provider != "" || model != "" || apiKey != "" || baseURL != "" {
		mc := cfg.Model
		if provider != "" {
			mc.Provider = provider
		}
		if model != "" {
			mc.Model = model
		}
		if apiKey != "" {
			mc.APIKey = apiKey
		}
		if baseURL != "" {
			mc.BaseURL = baseURL
		}
		cfg.WithModelConfig(mc)
	}
	return cfg, nil
}
