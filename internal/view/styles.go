// view/styles.go
package view

import "github.com/charmbracelet/lipgloss"

var (
	BorderStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("#626262")).
			Padding(0, 1)

	NavBarStyle = lipgloss.NewStyle().
			Padding(0, 0)

	SelectedNavStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#1db954")).
				Bold(true)

	ItemNavStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#b3b3b3")).
			Padding(0, 2)
)
