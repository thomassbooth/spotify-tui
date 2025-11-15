package view

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ---------------------------------------------------------------------
// Dashboard – the *only* full tea.Model we pass to bubbletea
// ---------------------------------------------------------------------
type dashboard struct {
	sidebar    Sidebar
	navigation Navigation
	width      int
	height     int
}

func NewDashboard() dashboard {
	return dashboard{
		sidebar:    NewSidebar(),
		navigation: NewNavigation(),
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

	// Forward messages to both sidebar and navigation
	var cmd1, cmd2 tea.Cmd
	d.sidebar, cmd1 = d.sidebar.Update(msg)
	d.navigation, cmd2 = d.navigation.Update(msg)

	return d, tea.Batch(cmd1, cmd2)
}

func (d dashboard) View() string {
	if d.width == 0 || d.height == 0 {
		return "loading..."
	}

	// Navigation bar spans full width at the top
	navBar := d.navigation.View(d.width)

	// Below nav: sidebar on left, content on right
	sidebarView := d.sidebar.View(22, d.height-4) // subtract nav height
	mainContent := d.navigation.ViewContent(d.width-22, d.height-4)

	// Sidebar and content side by side
	belowNav := lipgloss.JoinHorizontal(lipgloss.Top, sidebarView, mainContent)

	// Stack nav bar on top of everything
	return lipgloss.JoinVertical(lipgloss.Left, navBar, belowNav)
}
