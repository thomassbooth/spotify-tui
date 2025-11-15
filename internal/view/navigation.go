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

	searchBoxStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#242424")).
			Foreground(lipgloss.Color("#ffffff")).
			Padding(0, 2).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#535353"))

	contentStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#ffffff")).
			Padding(2, 2)
)

func NewNavigation() Navigation {
	return Navigation{
		selected: 1, // Start with Home selected
		items:    []string{"🔍 Search", "🏠 Home", "📚 Browse"},
	}
}

// Update handles messages for the navigation component
func (n Navigation) Update(msg tea.Msg) (Navigation, tea.Cmd) {
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

// View renders the navigation bar - should span full width at top
func (n Navigation) View(width int) string {
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

// ViewContent renders the content area based on selected navigation
func (n Navigation) ViewContent(width, height int) string {
	switch n.selected {
	case 0: // Search
		searchBox := searchBoxStyle.Width(width - 4).Render("What do you want to listen to?")
		return contentStyle.Width(width).Height(height).Render(searchBox)
	case 1: // Home
		return contentStyle.Width(width).Height(height).Render("🏠 Home View\n\nGood afternoon\nRecently played tracks would appear here...")
	case 2: // Browse
		return contentStyle.Width(width).Height(height).Render("📚 Browse View\n\nGenres & Moods\nNew Releases\nPodcasts...")
	default:
		return ""
	}
}

// GetSelected returns the currently selected navigation index
func (n Navigation) GetSelected() int {
	return n.selected
}

