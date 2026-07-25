package cli

import (
	"fmt"
	"strings"

	"github.com/codingagent/coding-agent/internal/agent"
)

// Formatter renders agent results for display.
type Formatter struct{}

// NewFormatter creates a new formatter.
func NewFormatter() *Formatter {
	return &Formatter{}
}

// Format renders an agent result with markdown and styling.
func (f *Formatter) Format(result agent.AgentResult) string {
	var sb strings.Builder

	sb.WriteString("\n")
	sb.WriteString(dividerStyle.Render(strings.Repeat("─", 60)))
	sb.WriteString("\n")
	sb.WriteString(renderMarkdown(result.Answer))
	sb.WriteString("\n")
	sb.WriteString(dividerStyle.Render(strings.Repeat("─", 60)))
	sb.WriteString("\n")

	sb.WriteString(statusDim.Render(
		fmt.Sprintf("Steps: %d | Tools: %d", result.Steps, len(result.ToolCalls)),
	))
	sb.WriteString("\n")

	for _, tc := range result.ToolCalls {
		icon := "✓"
		style := toolOkStyle
		if !tc.Success {
			icon = "✗"
			style = toolFailStyle
		}
		sb.WriteString(style.Render(fmt.Sprintf("  %s %s", icon, tc.Name)))
		sb.WriteString("\n")
	}

	return sb.String()
}
