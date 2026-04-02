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

	// Card box (normal)
	CardStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(0, 1).
		MarginBottom(0)

	// Card box (selected)
	SelectedCardStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("33")).
		Padding(0, 1).
		MarginBottom(0)

	// Card ID label
	CardIDStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("33")).
		Bold(true)

	// Card title
	CardTitleStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("15"))

	// Priority badge
	PriorityHighStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("1")).
		Bold(true)

	PriorityMediumStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("3"))

	PriorityLowStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("8"))

	// Type badge
	TypeStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("5"))

	// Status badge
	StatusStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("2"))

	// Card cursor indicator
	CursorStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("33")).
		Bold(true)

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
