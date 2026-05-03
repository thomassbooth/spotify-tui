// navigation.go
package view

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/thomassbooth/spotify-tui/internal/assets"
)

// Navigation component (not a full tea.Model)
type Navigation struct {
	selected int
	items    []string
	focused  bool
}

func NewNavigation() *Navigation {
	return &Navigation{
		selected: 1, // Start with Home selected
		items:    []string{"🔍 Search", "🏠 Home", "📚 Browse"},
		focused:  false,
	}
}

// Update handles messages for the navigation component
func (n *Navigation) Update(msg tea.Msg) (Component, tea.Cmd) {
	var cmd tea.Cmd
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
	return n, cmd

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

func (n *Navigation) View(width, height int) string {
	var parts []string
	for i, txt := range n.items {
		style := ItemNavStyle
		if i == n.selected {
			style = SelectedNavStyle
		}
		parts = append(parts, style.Render(txt))
	}
	nav := lipgloss.JoinHorizontal(lipgloss.Top, parts...)

	logoStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#1db954"))
	
	navWidth := lipgloss.Width(nav)
	logo := logoStyle.Width(width - navWidth - 2).Align(lipgloss.Right).Render(assets.SpotifyLogo)

	content := lipgloss.JoinHorizontal(lipgloss.Top, nav, logo)

	border := borderStyle.Copy().
		Width(width).
		Height(height)

	if n.Focused() {
		border = border.BorderForeground(lipgloss.Color("#1db954"))
	}

	return border.Render(content)
}
