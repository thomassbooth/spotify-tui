package view

import (
	"fmt"
	"io"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/thomassbooth/spotify-tui/internal/service"
)

type playlistItem struct {
	name    string
	artists []string
	id      string
}

func (i playlistItem) Title() string       { return i.name }
func (i playlistItem) Description() string { return i.id }
func (i playlistItem) FilterValue() string { return i.name }

type playlistDelegate struct {
	list.DefaultDelegate
}

func (d playlistDelegate) Height() int { return 2 }

func (d playlistDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	i, ok := item.(playlistItem)
	if !ok {
		return
	}

	var (
		title = i.name

		isSelected  = index == m.Index()
		s           = lipgloss.NewStyle().Padding(0, 0, 0, 2)
		selectedStr = " "
	)

	if isSelected {
		selectedStr = ">"
		title = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#1db954")).
			Bold(true).
			Render(title)

	} else {
		title = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FAFAFA")).
			Render(title)

	}

	fmt.Fprintf(w, s.Render(selectedStr+" "+title))
}

// ---------------------------------------------------------------------
// 3. Sidebar component
// ---------------------------------------------------------------------
type PlaylistTracks struct {
	tracks          list.Model
	focused         bool
	bus             *MessageBus
	playlistService *service.PlaylistService
}

func NewPlaylistTracks(bus *MessageBus, playlistService *service.PlaylistService) *PlaylistTracks {
	const defaultWidth = 30

	// ←←← CREATE THE LIST HERE
	delegate := playlistDelegate{}
	l := list.New([]list.Item{}, delegate, defaultWidth, 0)
	l.SetShowTitle(false)
	l.SetShowHelp(false)
	l.SetShowStatusBar(false)
	l.SetShowPagination(false)

	self := &PlaylistTracks{
		tracks:          l,
		focused:         false,
		bus:             bus,
		playlistService: playlistService,
	}

	bus.Subscribe(MsgPlaylistSelected, self) // pass pointer
	return self
}

func (s *PlaylistTracks) OnMessage(t MsgType, msg tea.Msg) tea.Cmd {
	// We only care about playlist selection
	if t != MsgPlaylistSelected {
		return nil
	}

	// Payload should be the playlist ID (string)
	playlistID, ok := msg.(string)
	if !ok {
		// optional: publish an error
		return nil
	}

	// -----------------------------------------------------------------
	// Fetch tracks from Spotify
	// -----------------------------------------------------------------
	tracks, _ := s.playlistService.GetPlaylistTracks(playlistID)

	// -----------------------------------------------------------------
	// Convert to list items
	// -----------------------------------------------------------------
	items := make([]list.Item, len(tracks))
	for i, tr := range tracks {
		artistNames := make([]string, len(tr.Artists))
		for j, a := range tr.Artists {
			artistNames[j] = a.Name
		}

		items[i] = playlistItem{
			name:    tr.Name,
			artists: artistNames,
			id:      tr.ID,
		}
	}

	// Update UI
	s.tracks.SetItems(items)

	// No extra command needed, but you could return a custom cmd here
	return nil
}

func (s *PlaylistTracks) Deselect() {
	s.tracks.Select(-1)
}

func (s *PlaylistTracks) Update(msg tea.Msg) (Component, tea.Cmd) {
	var cmd tea.Cmd
	if !s.focused {
		s.tracks.Select(-1)
		return s, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, defaultKeyMap.Tab),
			key.Matches(msg, defaultKeyMap.ShiftTab):
			s.tracks.Select(-1)
			return s, nil
		}
	}

	s.tracks, cmd = s.tracks.Update(msg)
	return s, cmd
}

func (s *PlaylistTracks) Blur() {
	s.focused = false
}

func (s *PlaylistTracks) Focus() {
	s.focused = true
}

func (s *PlaylistTracks) Focused() bool {
	return s.focused
}

func (s *PlaylistTracks) View(width, height int) string {
	s.tracks.SetSize(width, height)
	border := borderStyle.Copy().
		Width(width).
		Height(height)

	if s.Focused() {
		border = border.BorderForeground(lipgloss.Color("#1db954"))
	}

	return border.Render(s.tracks.View())
}
