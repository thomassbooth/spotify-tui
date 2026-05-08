package view

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/thomassbooth/spotify-tui/internal/assets"
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
	playbar    Component
	bus        *MessageBus
	width      int
	height     int
}

func NewPage(playlistService *service.PlaylistService, playbackService *service.PlaybackService) *Page {
	bus := NewMessageBus()
	sidebar := NewSidebar(bus, playlistService)
	sidebar.Focus()
	tracks := NewPlaylistTracks(bus, playlistService, playbackService)
	playbar := NewPlaybar(bus, playbackService)
	nav := NewNavigation(bus)
	return &Page{
		sidebar:    sidebar,
		navigation: nav,
		tracks:     tracks,
		playbar:    playbar,
		bus:        bus,
	}
}

func (p *Page) Init() tea.Cmd {
	return p.playbar.(*Playbar).Init()
}

func (p *Page) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch m := msg.(type) {
	case tea.KeyMsg:
		if m.String() == "q" {
			return p, tea.Quit
		}
		if m.String() == "tab" {
			p.cycleFocus()
			return p, nil
		}
		if m.String() == "Q" {
			cmds = append(cmds, p.bus.Publish(MsgToggleQueue, ToggleQueueMsg{}))
			return p, tea.Batch(cmds...)
		}
		if m.String() == "S" {
			cmds = append(cmds, p.bus.Publish(MsgToggleShuffle, ToggleShuffleMsg{}))
			return p, tea.Batch(cmds...)
		}

		if p.navigation.Focused() {
			p.navigation, cmd = p.navigation.Update(msg)
			cmds = append(cmds, cmd)
		} else if p.sidebar.Focused() {
			p.sidebar, cmd = p.sidebar.Update(msg)
			cmds = append(cmds, cmd)
		} else if p.tracks.Focused() {
			p.tracks, cmd = p.tracks.Update(msg)
			cmds = append(cmds, cmd)
		} else if p.playbar.Focused() {
			p.playbar, cmd = p.playbar.Update(msg)
			cmds = append(cmds, cmd)
		}

	case errMsg:
		// TODO: surface errors to the user (status bar, modal, etc.)
		_ = m

	case tea.WindowSizeMsg:
		p.width, p.height = m.Width, m.Height

	default:
		p.playbar, cmd = p.playbar.Update(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
		p.tracks, cmd = p.tracks.Update(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	return p, tea.Batch(cmds...)
}

func (p *Page) cycleFocus() {
	components := []Component{p.navigation, p.sidebar, p.tracks, p.playbar}

	for i, c := range components {
		if c.Focused() {
			c.Blur()
			components[(i+1)%len(components)].Focus()
			return
		}
	}
}

func (p *Page) View() string {
	if p.width == 0 || p.height == 0 {
		return "loading..."
	}

	height := p.height - 1
	width := p.width - 5

	logoLines := len(strings.Split(strings.Trim(assets.SpotifyLogo, "\n"), "\n"))
	navHeight := logoLines + 2
	const playbarHeight = 3

	const sidebarRatio = 0.35
	sidebarWidth := int(float64(width) * sidebarRatio)
	tracksWidth := width - sidebarWidth

	navBar := p.navigation.View(width+2, navHeight)
	contentHeight := height - navHeight - playbarHeight - 5

	sidebarView := p.sidebar.View(sidebarWidth, contentHeight)
	tracksView := p.tracks.View(tracksWidth, contentHeight)

	contentRow := lipgloss.JoinHorizontal(lipgloss.Top, sidebarView, tracksView)
	playbarView := p.playbar.View(width+2, playbarHeight)

	return lipgloss.JoinVertical(lipgloss.Left, navBar, contentRow, playbarView)
}