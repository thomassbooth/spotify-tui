package view

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Component interface {
	Update(tea.Msg) (Component, tea.Cmd)
	View(width, height int) string
	Focus()
	Blur()
	Focused() bool
}

type Page struct {
	sidebar    Component
	navigation Component
	bus        *MessageBus
	width      int
	height     int
}

func NewPage() Page {

	bus := NewMessageBus()
	sidebar := NewSidebar(bus)
	sidebar.Focus()

	return Page{
		sidebar:    sidebar,
		navigation: NewNavigation(),
		bus:        bus,
	}
}

// ---- tea.Model interface ------------------------------------------------
func (p Page) Init() tea.Cmd { return nil }

func (p Page) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd // Declare once

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "q" {
			return p, tea.Quit
		}
		if msg.String() == "tab" {
			p.cycleFocus()
			return p, nil
		}

		// Now you can use = instead of :=
		if p.navigation.Focused() {
			p.navigation, cmd = p.navigation.Update(msg)
			cmds = append(cmds, cmd)
		} else if p.sidebar.Focused() {
			p.sidebar, cmd = p.sidebar.Update(msg)
			cmds = append(cmds, cmd)
		}

	case tea.WindowSizeMsg:
		p.width, p.height = msg.Width, msg.Height
	}

	return p, tea.Batch(cmds...)
}

func (p *Page) cycleFocus() {
	components := []Component{p.navigation, p.sidebar}

	// Find currently focused
	for i, c := range components {
		if c.Focused() {
			c.Blur()
			next := (i + 1) % len(components)
			components[next].Focus()
			return
		}
	}
}

func (p Page) View() string {
	if p.width == 0 || p.height == 0 {
		return "loading..."
	}

	// Navigation bar spans full width at the top
	navBar := p.navigation.View(p.width, p.height)

	// Below nav: sidebar on left, content on right
	sidebarView := p.sidebar.View(22, p.height-4) // subtract nav height

	// Sidebar and content side by side
	belowNav := lipgloss.JoinHorizontal(lipgloss.Top, sidebarView)

	// Stack nav bar on top of everything
	return lipgloss.JoinVertical(lipgloss.Left, navBar, belowNav)
}
