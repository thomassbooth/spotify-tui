// navigation.go
package view

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Navigation component (not a full tea.Model)
type Navigation struct {
	selected int
	items    []string
	focused  bool
}

var (
	navBarStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#000000")).
			Foreground(lipgloss.Color("#b3b3b3")).
			Padding(1, 2)

	selectedNavStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#1db954")).
				Bold(true)

	itemNavStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#b3b3b3")).
			Padding(0, 2)
)

func NewNavigation() *Navigation {
	return &Navigation{
		selected: 1, // Start with Home selected
		items:    []string{"🔍 Search", "🏠 Home", "📚 Browse"},
		focused:  false,
	}
}

// Update handles messages for the navigation component
func (n *Navigation) Update(msg tea.Msg) (Component, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "left", "h":
			if n.selected > 0 {
				n.selected--
			}
		case "right", "l":
			if n.selected < len(n.items)-1 {
				n.selected++
			}
		case "1":
			n.selected = 0
		case "2":
			n.selected = 1
		case "3":
			n.selected = 2
		}
	}
	return n, nil
}

func (n *Navigation) Blur() {
	n.focused = false
}
func (n *Navigation) Focus() {
	n.focused = true
}

func (n *Navigation) Focused() bool {
	return n.focused
}

// View renders the navigation bar - should span full width at top
func (n *Navigation) View(width, height int) string {
	var navItems []string

	for i, item := range n.items {
		style := itemNavStyle
		if i == n.selected {
			style = selectedNavStyle
		}
		navItems = append(navItems, style.Render(item))
	}

	nav := lipgloss.JoinHorizontal(lipgloss.Left, navItems...)
	return navBarStyle.Width(width).Render(nav) // Full width
}
