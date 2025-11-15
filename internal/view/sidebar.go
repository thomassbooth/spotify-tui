// view/sidebar.go
package view

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var defaultKeyMap = struct {
	Tab      key.Binding
	ShiftTab key.Binding
}{
	Tab:      key.NewBinding(key.WithKeys("tab")),
	ShiftTab: key.NewBinding(key.WithKeys("shift+tab")),
}

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
	list    list.Model
	focused bool
	bus     *MessageBus
}

// NewSidebar creates a ready-to-use sidebar
func NewSidebar(bus *MessageBus) *Sidebar {
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
		// add as many as you want – scrolling works automatically
	}

	// Default delegate already draws a nice scrollbar on the right
	delegate := list.NewDefaultDelegate()
	l := list.New(items, delegate, width, 0)
	l.Title = "Spotify"
	l.SetShowHelp(false)
	l.SetShowStatusBar(false)

	return &Sidebar{list: l, focused: false, bus: bus}
}

func (s *Sidebar) Deselect() {
	s.list.Select(-1)
}

func (s *Sidebar) Update(msg tea.Msg) (Component, tea.Cmd) {
	var cmd tea.Cmd

	if !s.focused {
		s.list.Select(-1)
		return s, cmd
	}
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, defaultKeyMap.Tab),
			key.Matches(msg, defaultKeyMap.ShiftTab):
			s.list.Select(-1) // Modify s directly
			return s, nil     // Return the modified copy
		}
	}
	s.list, cmd = s.list.Update(msg)
	return s, cmd
}

// View renders the sidebar with a border
var borderStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("#626262")).
	Padding(0, 1)

func (s *Sidebar) Blur() {
	s.focused = false
}

func (s *Sidebar) Focus() {
	s.focused = true
}

func (s *Sidebar) Focused() bool {
	return s.focused
}

func (s *Sidebar) View(width, height int) string {
	s.list.SetSize(width, height)

	border := borderStyle.Copy().
		Width(width).
		Height(height)

	if s.Focused() {
		border = border.BorderForeground(lipgloss.Color("#1db954")) // Spotify green
	}

	return border.Render(s.list.View())
}
