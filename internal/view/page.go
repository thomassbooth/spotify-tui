package view

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/thomassbooth/spotify-tui/internal/service"
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
	tracks     Component
	bus        *MessageBus
	width      int
	height     int
}

func NewPage(playlistService *service.PlaylistService) Page {

	bus := NewMessageBus()
	sidebar := NewSidebar(bus, playlistService)
	sidebar.Focus()
	tracks := NewPlaylistTracks(bus, playlistService)
	return Page{
		sidebar:    sidebar,
		navigation: NewNavigation(),
		tracks:     tracks,
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
		} else if p.tracks.Focused() {
			p.tracks, cmd = p.tracks.Update(msg)
			cmds = append(cmds, cmd)
		}

	case tea.WindowSizeMsg:
		p.width, p.height = msg.Width, msg.Height
	}

	return p, tea.Batch(cmds...)
}

func (p *Page) cycleFocus() {
	components := []Component{p.navigation, p.sidebar, p.tracks}

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

	p.height = p.height - 1
	p.width = p.width - 5
	// === 1. Navigation bar (full width, fixed height) ===
	const navHeight = 3
	navBar := p.navigation.View(p.width, navHeight)

	// === 2. Main content area (below nav) ===
	contentHeight := p.height - navHeight - 3

	// === 3. Sidebar + Tracks (side by side) ===
	const sidebarRatio = 0.35 // 35% of width for sidebar
	sidebarWidth := int(float64(p.width) * sidebarRatio)
	tracksWidth := p.width - sidebarWidth

	sidebarView := p.sidebar.View(sidebarWidth, contentHeight)
	tracksView := p.tracks.View(tracksWidth, contentHeight)

	// Join sidebar + tracks horizontally
	contentRow := lipgloss.JoinHorizontal(lipgloss.Top, sidebarView, tracksView)

	// === 4. Stack nav on top of content ===
	return lipgloss.JoinVertical(lipgloss.Left, navBar, contentRow)
}
