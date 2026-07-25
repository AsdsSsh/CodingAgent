package cli

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// View renders the TUI.
func (m Model) View() string {
	if m.showPermModal && m.pendingPermReq != nil {
		return m.renderPermissionModal()
	}

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
	permLevel := m.config.Permissions.DefaultLevel

	left := headerBold.Render("CodingAgent")
	mid := headerDim.Render("│ " + modelName + " │ " + permLevel)
	right := headerDim.Render(fmt.Sprintf("│ %d tok", m.estimatedTokens()))

	content := lipgloss.JoinHorizontal(lipgloss.Center, left, mid, right)
	return headerStyle.Width(m.width).Render(content)
}

func (m Model) renderViewport() string {
	var sb strings.Builder

	// Welcome screen when no messages
	if len(m.messages) == 0 {
		sb.WriteString(m.renderWelcome())
	} else {
		var lastRole string
		for i, msg := range m.messages {
			// Insert divider between turns (user→assistant transitions)
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
			}
			lastRole = msg.Role
		}
	}

	m.viewport.SetContent(sb.String())
	m.viewport.GotoBottom()
	return m.viewport.View()
}

func (m Model) renderWelcome() string {
	var sb strings.Builder
	sb.WriteString("\n")
	sb.WriteString("\n")

	logo := lipgloss.NewStyle().
		Foreground(lipgloss.Color("63")).
		Bold(true).
		Render("  ╔══════════════════════════════╗\n" +
			"  ║   🖥️  CodingAgent Go         ║\n" +
			"  ╚══════════════════════════════╝")
	sb.WriteString(logo)
	sb.WriteString("\n\n")

	info := statusDim.Render("  Model: " + m.config.Model.Model +
		"  │  Permission: " + m.config.Permissions.DefaultLevel +
		"  │  Workspace: " + m.config.Workspace)
	sb.WriteString(info)
	sb.WriteString("\n\n")

	tips := []string{
		"  Enter  → submit a single-line task",
		"  Ctrl+S → submit multi-line",
		"  /help  → see all commands",
		"  /exit  → quit",
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

func (m Model) renderStatusBar() string {
	if m.agentRunning {
		spinner := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
		frame := spinner[m.spinnerTick%len(spinner)]

		// Distinguish thinking vs executing
		var stateText string
		if strings.Contains(m.statusLine, "·") && !strings.Contains(m.statusLine, "Step") {
			stateText = m.statusLine
		} else if m.statusLine == "Thinking..." {
			stateText = spinnerStyle.Render("🧠 ") + m.statusLine
		} else {
			stateText = spinnerStyle.Render("🔧 ") + m.statusLine
		}

		return statusStyle.Render(spinnerStyle.Render(frame) + " " + stateText)
	}
	return statusStyle.Render(statusDim.Render("idle — " + m.statusLine))
}

func (m Model) renderInput() string {
	if m.agentRunning {
		return statusDim.Render("  Running: " + truncate(m.taskInFlight, 60))
	}
	return m.input.View()
}

func (m Model) renderHelpBar() string {
	help := lipgloss.JoinHorizontal(lipgloss.Center,
		"Enter: Submit (1 line)",
		helpDimStyle.Render(" │ "),
		"Ctrl+S: Submit (multi)",
		helpDimStyle.Render(" │ "),
		"Ctrl+C: Cancel/Quit",
		helpDimStyle.Render(" │ "),
		"/: Commands",
	)
	return helpBarStyle.Width(m.width).Render(help)
}

func (m Model) renderPermissionModal() string {
	req := m.pendingPermReq
	title := modalTitleStyle.Render("⚠️  Permission Required")
	body := fmt.Sprintf("\nTool: %s\n\nReason: %s\n\n", req.ToolName, req.Reason)
	buttons := lipgloss.JoinHorizontal(lipgloss.Center,
		modalActiveStyle.Render("[Y] Allow"),
		"  ",
		modalButtonStyle.Render("[N] Deny"),
		"  ",
		modalButtonStyle.Render("[A] Always Allow"),
	)
	help := statusDim.Render("\n\nPress Y/N/A or Enter/Esc")

	content := title + "\n" + body + "\n" + buttons + help
	box := modalBoxStyle.Render(content)

	return lipgloss.Place(m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		box,
	)
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
