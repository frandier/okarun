package cli

import "github.com/charmbracelet/lipgloss"

var (
	// TitleStyle defines the style for titles
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFB703")).
			MarginLeft(2)

	// DocStyle defines the style for the document
	DocStyle = lipgloss.NewStyle().Margin(1, 2)

	// HighlightStyle defines the style for highlighted elements
	HighlightStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FB8500")).
			Bold(true)

	// LoadingStyle defines the style for the loading spinner
	LoadingStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("205"))
)
