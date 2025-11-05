// view/sidebar.go
package view

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ---------------------------------------------------------------------
// 1. Item type – any struct that implements list.Item works
// ---------------------------------------------------------------------
type sidebarItem string

func (i sidebarItem) Title() string       { return string(i) }
func (i sidebarItem) Description() string { return "" }
func (i sidebarItem) FilterValue() string { return string(i) }

// ---------------------------------------------------------------------
// 2. Sidebar component (sub-component, NOT a full tea.Model)
// ---------------------------------------------------------------------
type Sidebar struct {
	list list.Model
}

// NewSidebar creates a ready-to-use sidebar
func NewSidebar() Sidebar {
	const width = 22 // fixed width you asked for

	items := []list.Item{
		sidebarItem("Library"),
		sidebarItem("Search"),
		sidebarItem("Playlists"),
		sidebarItem("Browse"),
		sidebarItem("Radio"),
		sidebarItem("Your Episodes"),
		sidebarItem("Liked Songs"),
		sidebarItem("Albums"),
		sidebarItem("Artists"),
		sidebarItem("Podcasts"),
		sidebarItem("Podcasts"),
		sidebarItem("Podcasts"),
		sidebarItem("Podcasts"),
		sidebarItem("Podcasts"),
		sidebarItem("Podcasts"),
		sidebarItem("Podcasts"),
		sidebarItem("Podcasts"),
		sidebarItem("Podcasts"),
		sidebarItem("Podcasts"),

		// add as many as you want – scrolling works automatically
	}

	// Default delegate already draws a nice scrollbar on the right
	delegate := list.NewDefaultDelegate()

	l := list.New(items, delegate, width, 0)
	l.Title = "Spotify"
	l.SetShowHelp(false)
	l.SetShowStatusBar(false)

	return Sidebar{list: l}
}

// Update is the only method a sub-component needs
func (s Sidebar) Update(msg tea.Msg) (Sidebar, tea.Cmd) {
	var cmd tea.Cmd
	s.list, cmd = s.list.Update(msg)
	return s, cmd
}

// View renders the sidebar with a border
var borderStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("#626262")).
	Padding(0, 1)

func (s Sidebar) View(width, height int) string {
	s.list.SetSize(width, height) // resize to the area we give it
	return borderStyle.Copy().
		Width(width).
		Height(height).
		Render(s.list.View())
}
