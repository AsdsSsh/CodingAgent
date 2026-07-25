package cli

import (
	"regexp"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	boldStyle      = lipgloss.NewStyle().Bold(true)
	italicStyle    = lipgloss.NewStyle().Italic(true)
	codeStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("6"))
	linkStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("4")).Underline(true)
	heading1Style  = lipgloss.NewStyle().Bold(true).Underline(true)
	heading2Style  = lipgloss.NewStyle().Bold(true)
	listItemStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
)

// renderMarkdown renders basic markdown to ANSI-styled text.
func renderMarkdown(md string) string {
	if md == "" {
		return ""
	}

	var out strings.Builder
	lines := strings.Split(md, "\n")
	inCodeBlock := false

	for _, line := range lines {
		// Code block toggle
		if strings.HasPrefix(strings.TrimSpace(line), "```") {
			inCodeBlock = !inCodeBlock
			if inCodeBlock {
				out.WriteString(statusDim.Render("+-- " + strings.TrimPrefix(strings.TrimSpace(line), "```") + " "))
				out.WriteString(strings.Repeat("─", 40))
				out.WriteByte('\n')
			} else {
				out.WriteString(statusDim.Render("+" + strings.Repeat("─", 50)))
				out.WriteByte('\n')
			}
			continue
		}

		if inCodeBlock {
			out.WriteString(statusDim.Render(" | " + line))
			out.WriteByte('\n')
			continue
		}

		// Headings
		if match, _ := regexp.MatchString(`^#{1,4}\s+`, line); match {
			level := 0
			for level < len(line) && line[level] == '#' {
				level++
			}
			text := strings.TrimSpace(line[level:])
			if level <= 2 {
				out.WriteString(heading1Style.Render(text))
			} else {
				out.WriteString(heading2Style.Render(text))
			}
			out.WriteByte('\n')
			continue
		}

		// Horizontal rule
		if match, _ := regexp.MatchString(`^[-*=]{3,}$`, line); match {
			out.WriteString(dividerStyle.Render(strings.Repeat("─", 60)))
			out.WriteByte('\n')
			continue
		}

		// List items
		if match, _ := regexp.MatchString(`^\s*[-*+]\s+`, line); match {
			re := regexp.MustCompile(`^\s*[-*+]\s+`)
			text := re.ReplaceAllString(line, "")
			out.WriteString("  " + listItemStyle.Render("• ") + renderInline(text))
			out.WriteByte('\n')
			continue
		}

		// Ordered list
		if match, _ := regexp.MatchString(`^\s*\d+\.\s+`, line); match {
			re := regexp.MustCompile(`^\s*\d+\.\s+`)
			text := re.ReplaceAllString(line, "")
			out.WriteString("  " + listItemStyle.Render("- ") + renderInline(text))
			out.WriteByte('\n')
			continue
		}

		// Blockquote
		if strings.HasPrefix(line, "> ") {
			out.WriteString(statusDim.Render("| " + renderInline(line[2:])))
			out.WriteByte('\n')
			continue
		}

		// Normal line
		out.WriteString(renderInline(line))
		out.WriteByte('\n')
	}

	return out.String()
}

var (
	boldRe      = regexp.MustCompile(`\*\*(.+?)\*\*`)
	boldRe2     = regexp.MustCompile(`__(.+?)__`)
	italicRe    = regexp.MustCompile(`\*([^*]+)\*`)
	italicRe2   = regexp.MustCompile(`_([^_]+)_`)
	codeRe      = regexp.MustCompile("`([^`]+)`")
	linkRe      = regexp.MustCompile(`\[([^\]]+)]\([^)]+\)`)
)

// renderInline renders inline markdown formatting.
func renderInline(text string) string {
	text = boldRe.ReplaceAllString(text, boldStyle.Render("$1"))
	text = boldRe2.ReplaceAllString(text, boldStyle.Render("$1"))
	text = italicRe.ReplaceAllString(text, italicStyle.Render("$1"))
	text = italicRe2.ReplaceAllString(text, italicStyle.Render("$1"))
	text = codeRe.ReplaceAllString(text, codeStyle.Render("$1"))
	text = linkRe.ReplaceAllString(text, linkStyle.Render("$1"))
	return text
}
