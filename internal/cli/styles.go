package cli

import "github.com/charmbracelet/lipgloss"

var (
	headerStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("63")).
			Foreground(lipgloss.Color("15")).
			Padding(0, 1)

	headerBold = lipgloss.NewStyle().Bold(true)

	headerDim = lipgloss.NewStyle().Foreground(lipgloss.Color("250"))

	statusStyle = lipgloss.NewStyle().Padding(0, 1)

	statusDim = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))

	userStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("12")).
			Bold(true)

	asstStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("15"))

	toolOkStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))

	toolFailStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))

	errorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))

	dividerStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))

	spinnerStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("6"))

	helpBarStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("236")).
			Foreground(lipgloss.Color("250")).
			Padding(0, 1)

	helpDimStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))

	modalBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("11")).
			Padding(1, 2)

	modalTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("11"))

	modalButtonStyle = lipgloss.NewStyle().Padding(0, 1)

	modalActiveStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("63")).
			Foreground(lipgloss.Color("15")).
			Padding(0, 1)
)

// Styles holds all lipgloss styles used in the TUI.
type Styles struct{}

// NewStyles creates the default TUI styles holder.
func NewStyles() Styles {
	return Styles{}
}
