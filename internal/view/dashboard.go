package view

import (
	tea "github.com/charmbracelet/bubbletea"
)

// ---------------------------------------------------------------------
// Dashboard – the *only* full tea.Model we pass to bubbletea
// ---------------------------------------------------------------------
type dashboard struct {
	sidebar Sidebar
	width   int
	height  int
}

func NewDashboard() dashboard {
	return dashboard{
		sidebar: NewSidebar(),
	}
}

// ---- tea.Model interface ------------------------------------------------
func (d dashboard) Init() tea.Cmd { return nil }

func (d dashboard) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "q" || msg.String() == "ctrl+c" {
			return d, tea.Quit
		}
	case tea.WindowSizeMsg:
		d.width, d.height = msg.Width, msg.Height
	}

	// forward every message to the sidebar
	var cmd tea.Cmd
	d.sidebar, cmd = d.sidebar.Update(msg)
	return d, cmd
}

func (d dashboard) View() string {
	if d.width == 0 || d.height == 0 {
		return "loading..."
	}
	// The sidebar is the *only* thing we render
	return d.sidebar.View(22, d.height)
}
