package cli

import (
	"context"
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/codingagent/coding-agent/internal/agent"
	"github.com/codingagent/coding-agent/internal/config"
	"github.com/codingagent/coding-agent/internal/sandbox"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.viewport.Width = msg.Width - 4
		m.viewport.Height = msg.Height - 8
		m.input.SetWidth(msg.Width - 4)

	case tea.KeyMsg:
		// Permission pending: only respond to Y/N/A
		if m.pendingPermReq != nil {
			switch msg.String() {
			case "y", "Y", "enter":
				req := *m.pendingPermReq
				m.pendingPermReq = nil
				m.messages = append(m.messages, ChatMessage{
					Role:    "permission",
					Content: fmt.Sprintf("%s: ALLOWED (once)", req.ToolName),
				})
				select {
				case m.permRespChan <- agent.PermissionResponse{Allowed: true}:
				default:
				}
				return m, tea.Batch(m.listenProgress(), m.listenPermission())
			case "n", "N", "esc":
				req := *m.pendingPermReq
				m.pendingPermReq = nil
				m.messages = append(m.messages, ChatMessage{
					Role:    "permission",
					Content: fmt.Sprintf("%s: DENIED", req.ToolName),
				})
				select {
				case m.permRespChan <- agent.PermissionResponse{Allowed: false}:
				default:
				}
				return m, tea.Batch(m.listenProgress(), m.listenPermission())
			case "a", "A":
				req := *m.pendingPermReq
				m.pendingPermReq = nil
				m.messages = append(m.messages, ChatMessage{
					Role:    "permission",
					Content: fmt.Sprintf("%s: ALLOWED (always)", req.ToolName),
				})
				select {
				case m.permRespChan <- agent.PermissionResponse{Allowed: true, AlwaysAllow: true}:
				default:
				}
				return m, tea.Batch(m.listenProgress(), m.listenPermission())
			}
			// Block all other keys when permission is pending
			return m, nil
		}

		switch msg.String() {
		case "ctrl+c":
			if m.agentRunning {
				if m.cancelFunc != nil {
					m.cancelFunc() // cancel in-flight LLM request
				}
				m.agentRunning = false
				m.messages = append(m.messages, ChatMessage{Role: "error", Content: "Task cancelled"})
				m.statusLine = "Cancelled"
				return m, nil
			}
			return m, tea.Quit

		case "ctrl+p":
			// Cycle permission level
			m.cyclePermission()
			return m, nil

		case "ctrl+s":
			if m.focus == FocusInput && !m.agentRunning {
				task := strings.TrimSpace(m.input.Value())
				if task != "" {
					m.messages = append(m.messages, ChatMessage{Role: "user", Content: task})
					m.input.Reset()
					return m.handleTaskSubmit(task)
				}
			}

		case "enter":
			if m.focus == FocusInput && !m.agentRunning {
				task := strings.TrimSpace(m.input.Value())
				if strings.HasPrefix(task, "/") {
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

	case AgentProgressMsg:
		state := msg.State
		if state.CurrentTool != "" {
			m.statusLine = fmt.Sprintf("Step %d · %s · %d tokens",
				state.Step, state.CurrentTool, state.ConversationTokens)
			m.messages = append(m.messages, ChatMessage{
				Role:   "tool",
				Content: fmt.Sprintf("Step %d: %s", state.Step, state.CurrentTool),
				ToolOk: true,
			})
		} else {
			m.statusLine = fmt.Sprintf("Thinking... (%d tokens)", state.ConversationTokens)
		}
		cmds = append(cmds, m.listenProgress())

	case AgentResultMsg:
		m.agentRunning = false
		result := msg.Result
		if result.Error != nil {
			m.messages = append(m.messages, ChatMessage{Role: "error", Content: result.Error.Error()})
		} else {
			m.messages = append(m.messages, ChatMessage{Role: "assistant", Content: result.Answer})
		}
		m.statusLine = fmt.Sprintf("Done — %d steps, %d tools", result.Steps, len(result.ToolCalls))

	case PermissionPromptMsg:
		m.pendingPermReq = &msg.Request

	case PermissionResponseMsg:
		if m.permRespChan != nil {
			select {
			case m.permRespChan <- msg.Response:
			default:
			}
		}
		m.pendingPermReq = nil

	case TickMsg:
		if m.agentRunning {
			m.spinnerTick++
			cmds = append(cmds, m.tickLater())
		}
	}

	if m.focus == FocusInput && !m.agentRunning && m.pendingPermReq == nil {
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		cmds = append(cmds, cmd)
	}

	var vpCmd tea.Cmd
	m.viewport, vpCmd = m.viewport.Update(msg)
	cmds = append(cmds, vpCmd)

	return m, tea.Batch(cmds...)
}

// cyclePermission cycles through permission levels: READ → WRITE → EXECUTE → DANGEROUS → READ
func (m *Model) cyclePermission() {
	levels := []sandbox.PermissionLevel{
		sandbox.LevelRead,
		sandbox.LevelWrite,
		sandbox.LevelExecute,
		sandbox.LevelDangerous,
	}
	current := m.permissionLevel
	next := levels[(int(current)+1)%len(levels)]
	m.permissionLevel = next
	m.config.Permissions.DefaultLevel = next.String()

	// If agent is running, also escalate the running sandbox
	if m.orchestrator != nil {
		// The sandbox is created per-task; escalate current policy engine
		m.statusLine = fmt.Sprintf("Permission: %s → %s", current.String(), next.String())
	}

	m.messages = append(m.messages, ChatMessage{
		Role:    "permission",
		Content: fmt.Sprintf("Permission: %s → %s", current.String(), next.String()),
	})
}

func (m Model) handleTaskSubmit(task string) (tea.Model, tea.Cmd) {
	// Apply current permission level to config before creating orchestrator
	m.config.Permissions.DefaultLevel = m.permissionLevel.String()

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
	m.statusLine = "Thinking..."
	m.pendingPermReq = nil

	m.progressChan = make(chan agent.AgentState, 20)
	m.resultChan = make(chan agent.AgentResult, 1)
	m.permReqChan = make(chan agent.PermissionRequest, 5)
	m.permRespChan = make(chan agent.PermissionResponse, 5)

	ctx, cancel := context.WithCancel(context.Background())
	m.cancelFunc = cancel

	go func() {
		result := orch.RunWithChannels(ctx, task, m.progressChan, m.permReqChan, m.permRespChan)
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

func (m Model) listenProgress() tea.Cmd {
	return func() tea.Msg {
		state, ok := <-m.progressChan
		if !ok {
			return nil
		}
		return AgentProgressMsg{State: state}
	}
}

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

func (m Model) listenPermission() tea.Cmd {
	return func() tea.Msg {
		req, ok := <-m.permReqChan
		if !ok {
			return nil
		}
		return PermissionPromptMsg{Request: req}
	}
}

func (m Model) tickLater() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		return TickMsg{}
	})
}

func (m Model) handleSlashCommand(cmd, args string) tea.Cmd {
	switch cmd {
	case "help", "h":
		helpText := `Available commands:
  /model [name]     View or switch model
  /permission [lvl] View or cycle permission (READ/WRITE/EXECUTE/DANGEROUS)
  /clear            Clear conversation history
  /exit             Exit CodingAgent

  Ctrl+P            Cycle permission level`
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
			// No args: cycle (same behavior as Ctrl+P)
			m.cyclePermission()
			return nil
		}
		level := sandbox.FromString(args)
		old := m.permissionLevel
		m.permissionLevel = level
		m.config.Permissions.DefaultLevel = level.String()
		m.messages = append(m.messages, ChatMessage{
			Role:    "permission",
			Content: fmt.Sprintf("Permission: %s → %s", old.String(), level.String()),
		})
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
