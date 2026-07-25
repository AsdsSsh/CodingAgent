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

	for _, msg := range m.messages {
		switch msg.Role {
		case "user":
			sb.WriteString(userStyle.Render("\n> " + msg.Content) + "\n")
		case "assistant":
			sb.WriteString(asstStyle.Render(renderMarkdown(msg.Content)) + "\n")
		case "tool":
			if msg.ToolOk {
				sb.WriteString(toolOkStyle.Render("  ✓ " + msg.Content) + "\n")
			} else {
				sb.WriteString(toolFailStyle.Render("  ✗ " + msg.Content) + "\n")
			}
		case "error":
			sb.WriteString(errorStyle.Render("  ⚠ " + msg.Content) + "\n")
		}
	}

	if m.agentRunning {
		sb.WriteString(dividerStyle.Render(strings.Repeat("─", m.width-4)) + "\n")
	}

	m.viewport.SetContent(sb.String())
	return m.viewport.View()
}

func (m Model) renderStatusBar() string {
	if m.agentRunning {
		spinner := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
		frame := spinner[m.spinnerTick%len(spinner)]
		return statusStyle.Render(spinnerStyle.Render(frame) + " " + m.statusLine)
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
		"Ctrl+Enter: Submit",
		helpDimStyle.Render(" │ "),
		"Ctrl+C: Cancel/Quit",
		helpDimStyle.Render(" │ "),
		"Tab: Switch focus",
		helpDimStyle.Render(" │ "),
		"/: Commands",
	)
	return helpBarStyle.Width(m.width).Render(help)
}

func (m Model) renderPermissionModal() string {
	req := m.pendingPermReq
	title := modalTitleStyle.Render("⚠️ Permission Required")
	body := fmt.Sprintf("\nTool: %s\n\nReason: %s\n\n", req.ToolName, req.Reason)
	buttons := lipgloss.JoinHorizontal(lipgloss.Center,
		modalActiveStyle.Render("[ Allow ]"),
		"  ",
		modalButtonStyle.Render("[ Deny ]"),
		"  ",
		modalButtonStyle.Render("[ Always Allow ]"),
	)

	content := title + "\n" + body + "\n" + buttons
	box := modalBoxStyle.Render(content)

	return lipgloss.Place(m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		box,
	)
}

func (m Model) estimatedTokens() int {
	// Rough estimate from conversation messages
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
