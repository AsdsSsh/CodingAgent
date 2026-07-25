package cli

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/codingagent/coding-agent/internal/sandbox"
)

// View renders the TUI.
func (m Model) View() string {
	return lipgloss.JoinVertical(lipgloss.Left,
		m.renderHeader(),
		m.renderViewport(),
		m.renderStatusBar(),
		m.renderInput(),
		m.renderHelpBar(),
	)
}

func (m Model) renderHeader() string {
	modelName := m.config.Model.Model
	permColor := permissionColor(m.permissionLevel)
	permText := lipgloss.NewStyle().
		Foreground(permColor).
		Bold(true).
		Render(m.permissionLevel.String())

	left := headerBold.Render("CodingAgent")
	mid := headerDim.Render("│ " + modelName + " │ ") + permText
	right := headerDim.Render(fmt.Sprintf("│ %d tok", m.estimatedTokens()))

	content := lipgloss.JoinHorizontal(lipgloss.Center, left, mid, right)
	return headerStyle.Width(m.width).Render(content)
}

func (m Model) renderViewport() string {
	var sb strings.Builder

	if len(m.messages) == 0 && m.pendingPermReq == nil {
		sb.WriteString(m.renderWelcome())
	} else {
		var lastRole string
		for i, msg := range m.messages {
			if i > 0 && msg.Role == "user" && lastRole != "user" {
				sb.WriteString(dividerStyle.Render(strings.Repeat("─", m.width-4)) + "\n")
			}

			switch msg.Role {
			case "user":
				sb.WriteString(m.renderUserMessage(msg))
			case "assistant":
				sb.WriteString(m.renderAssistantMessage(msg))
			case "tool":
				if msg.ToolOk {
					sb.WriteString(toolOkStyle.Render("  ✓ " + msg.Content) + "\n")
				} else {
					sb.WriteString(toolFailStyle.Render("  ✗ " + msg.Content) + "\n")
				}
			case "error":
				sb.WriteString(m.renderErrorMessage(msg))
			case "permission":
				sb.WriteString(m.renderPermissionHistory(msg))
			}
			lastRole = msg.Role
		}

		// Inline permission prompt (replaces modal)
		if m.pendingPermReq != nil {
			sb.WriteString("\n")
			sb.WriteString(m.renderPermissionInline())
		}
	}

	m.viewport.SetContent(sb.String())
	m.viewport.GotoBottom()
	return m.viewport.View()
}

func (m Model) renderWelcome() string {
	var sb strings.Builder
	sb.WriteString("\n\n")
	logo := lipgloss.NewStyle().
		Foreground(lipgloss.Color("63")).
		Bold(true).
		Render("  ╔══════════════════════════════╗\n" +
			"  ║   🖥️  CodingAgent Go         ║\n" +
			"  ╚══════════════════════════════╝")
	sb.WriteString(logo)
	sb.WriteString("\n\n")
	info := statusDim.Render("  Model: " + m.config.Model.Model +
		"  │  Permission: " + m.permissionLevel.String() +
		"  │  Workspace: " + m.config.Workspace)
	sb.WriteString(info)
	sb.WriteString("\n\n")
	tips := []string{
		"  Enter  → submit a single-line task",
		"  Ctrl+S → submit multi-line",
		"  Ctrl+P → cycle permission level",
		"  /help  → see all commands",
	}
	for _, tip := range tips {
		sb.WriteString(statusDim.Render(tip) + "\n")
	}
	sb.WriteString("\n")
	sb.WriteString(asstStyle.Render("  Type a programming task to begin..."))
	return sb.String()
}

func (m Model) renderUserMessage(msg ChatMessage) string {
	prefix := lipgloss.NewStyle().
		Foreground(lipgloss.Color("12")).
		Bold(true).
		Render("▌")
	return prefix + " " + userStyle.Render(msg.Content) + "\n"
}

func (m Model) renderAssistantMessage(msg ChatMessage) string {
	if msg.Content == "" {
		return ""
	}
	return asstStyle.Render(renderMarkdown(msg.Content)) + "\n\n"
}

func (m Model) renderErrorMessage(msg ChatMessage) string {
	return errorStyle.Render("  ⚠  " + msg.Content) + "\n"
}

func (m Model) renderPermissionHistory(msg ChatMessage) string {
	permStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("11")).
		Bold(true)
	return permStyle.Render("  🔓 " + msg.Content) + "\n"
}

func (m Model) renderPermissionInline() string {
	req := m.pendingPermReq
	permColor := lipgloss.Color("11") // yellow border

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(permColor).
		Padding(1, 2).
		Width(m.width - 4)

	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(permColor).
		Render("⚠️  Permission Required")

	body := fmt.Sprintf("\n%s needs %s permission\nReason: %s\n",
		req.ToolName,
		lipgloss.NewStyle().Foreground(lipgloss.Color("3")).Render("higher"),
		req.Reason,
	)

	buttons := lipgloss.JoinHorizontal(lipgloss.Center,
		lipgloss.NewStyle().Background(lipgloss.Color("63")).Foreground(lipgloss.Color("15")).Padding(0, 1).Render("[Y] Allow"),
		"  ",
		lipgloss.NewStyle().Background(lipgloss.Color("63")).Foreground(lipgloss.Color("15")).Padding(0, 1).Render("[A] Always"),
		"  ",
		lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Padding(0, 1).Render("[N] Deny"),
	)

	return boxStyle.Render(title + body + "\n" + buttons) + "\n"
}

func (m Model) renderStatusBar() string {
	if m.pendingPermReq != nil {
		alertStyle := lipgloss.NewStyle().
			Background(lipgloss.Color("9")).
			Foreground(lipgloss.Color("15")).
			Bold(true).
			Padding(0, 1)
		return alertStyle.Render(" ⚠  Waiting for permission... Press Y/N/A ") + statusStyle.Render("")
	}

	if m.agentRunning {
		spinner := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
		frame := spinner[m.spinnerTick%len(spinner)]
		stateText := m.statusLine
		if strings.Contains(m.statusLine, "Thinking") {
			stateText = "🧠 " + m.statusLine
		} else if !strings.Contains(m.statusLine, "Done") {
			stateText = "🔧 " + m.statusLine
		}
		return statusStyle.Render(spinnerStyle.Render(frame) + " " + stateText)
	}

	permIndicator := lipgloss.NewStyle().
		Foreground(permissionColor(m.permissionLevel)).
		Render(" · " + m.permissionLevel.String())
	return statusStyle.Render(statusDim.Render("idle — " + m.statusLine + permIndicator))
}

func (m Model) renderInput() string {
	if m.pendingPermReq != nil {
		// Blocked: show permission prompt instead of textarea
		blocked := lipgloss.NewStyle().
			Background(lipgloss.Color("9")).
			Foreground(lipgloss.Color("15")).
			Bold(true).
			Padding(0, 2).
			Width(m.width - 4).
			Render("🔒  Input blocked — respond to the permission request above (Y/N/A)")
		return blocked
	}

	if m.agentRunning {
		return statusDim.Render("  Running...")
	}
	return m.input.View()
}

func (m Model) renderHelpBar() string {
	help := lipgloss.JoinHorizontal(lipgloss.Center,
		"Enter: Submit (1 line)",
		helpDimStyle.Render(" │ "),
		"Ctrl+S: Submit (multi)",
		helpDimStyle.Render(" │ "),
		"Ctrl+P: Permission",
		helpDimStyle.Render(" │ "),
		"Ctrl+C: Cancel/Quit",
		helpDimStyle.Render(" │ "),
		"/: Commands",
	)
	return helpBarStyle.Width(m.width).Render(help)
}

func (m Model) estimatedTokens() int {
	total := 0
	for _, msg := range m.messages {
		total += len(msg.Content) / 4
	}
	return total
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func permissionColor(level sandbox.PermissionLevel) lipgloss.Color {
	switch level {
	case sandbox.LevelDangerous:
		return lipgloss.Color("9") // red
	case sandbox.LevelExecute:
		return lipgloss.Color("208") // orange
	case sandbox.LevelWrite:
		return lipgloss.Color("3") // yellow
	default:
		return lipgloss.Color("8") // gray
	}
}
