package tui

import "github.com/charmbracelet/lipgloss"

var (
	// Colors
	primaryColor   = lipgloss.Color("#7C3AED") // Purple
	secondaryColor = lipgloss.Color("#10B981") // Green
	warningColor   = lipgloss.Color("#F59E0B") // Amber
	errorColor     = lipgloss.Color("#EF4444") // Red
	mutedColor     = lipgloss.Color("#6B7280") // Gray
	bgColor        = lipgloss.Color("#1F2937") // Dark gray

	// Title style
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(primaryColor).
			MarginBottom(1)

	// List item styles
	selectedItemStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFFFFF")).
				Background(primaryColor).
				Padding(0, 1)

	normalItemStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Padding(0, 1)

	// Status styles
	statusUnreadStyle = lipgloss.NewStyle().
				Foreground(mutedColor).
				Italic(true)

	statusReadingStyle = lipgloss.NewStyle().
				Foreground(warningColor).
				Bold(true)

	statusFinishedStyle = lipgloss.NewStyle().
				Foreground(secondaryColor).
				Bold(true)

	// Progress bar styles
	progressFullStyle = lipgloss.NewStyle().
				Foreground(secondaryColor)

	progressEmptyStyle = lipgloss.NewStyle().
				Foreground(mutedColor)

	// Detail view styles
	labelStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(primaryColor).
			Width(12)

	valueStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF"))

	// Help style
	helpStyle = lipgloss.NewStyle().
			Foreground(mutedColor).
			MarginTop(1)

	// Box styles
	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(primaryColor).
			Padding(1, 2)

	// Input style
	inputStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(mutedColor).
			Padding(0, 1)

	focusedInputStyle = lipgloss.NewStyle().
				Border(lipgloss.NormalBorder()).
				BorderForeground(primaryColor).
				Padding(0, 1)

	// Error style
	errorStyle = lipgloss.NewStyle().
			Foreground(errorColor).
			Bold(true)

	// Success style
	successStyle = lipgloss.NewStyle().
			Foreground(secondaryColor).
			Bold(true)
)

// RenderProgress renders a progress bar
func RenderProgress(current, total int, width int) string {
	if total == 0 {
		return progressEmptyStyle.Render(repeat("─", width))
	}

	progress := float64(current) / float64(total)
	filled := int(progress * float64(width))
	empty := width - filled

	return progressFullStyle.Render(repeat("█", filled)) +
		progressEmptyStyle.Render(repeat("░", empty))
}

func repeat(s string, n int) string {
	if n <= 0 {
		return ""
	}
	result := ""
	for i := 0; i < n; i++ {
		result += s
	}
	return result
}
