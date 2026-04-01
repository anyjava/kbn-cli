package tui

import "github.com/charmbracelet/lipgloss"

var (
	// Column header
	HeaderStyle = lipgloss.NewStyle().
		Bold(true).
		Padding(0, 1).
		Foreground(lipgloss.Color("15")).
		Background(lipgloss.Color("62"))

	// Active column header
	ActiveHeaderStyle = lipgloss.NewStyle().
		Bold(true).
		Padding(0, 1).
		Foreground(lipgloss.Color("15")).
		Background(lipgloss.Color("33"))

	// Card (normal)
	CardStyle = lipgloss.NewStyle().
		Padding(0, 1)

	// Card (selected)
	SelectedCardStyle = lipgloss.NewStyle().
		Padding(0, 1).
		Bold(true).
		Foreground(lipgloss.Color("33"))

	// Column border
	ColumnStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(0, 0)

	// Active column border
	ActiveColumnStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("33")).
		Padding(0, 0)

	// Preview panel
	PreviewStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(0, 1)

	// Help bar
	HelpStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("241"))

	// Search input
	SearchStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("33"))
)
