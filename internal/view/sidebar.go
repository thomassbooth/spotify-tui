// view/sidebar.go
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

var defaultKeyMap = struct {
	Tab      key.Binding
	ShiftTab key.Binding
}{
	Tab:      key.NewBinding(key.WithKeys("tab")),
	ShiftTab: key.NewBinding(key.WithKeys("shift+tab")),
}

// ---------------------------------------------------------------------
// 1. Item type with Name and OwnerName
// ---------------------------------------------------------------------
type sidebarItem struct {
	name      string
	ownerName string
	plType    string
	id        string
}

func (i sidebarItem) Title() string       { return i.name }
func (i sidebarItem) Description() string { return i.ownerName }
func (i sidebarItem) FilterValue() string { return i.name }

// ---------------------------------------------------------------------
// 2. Custom Delegate to render name and owner
// ---------------------------------------------------------------------
type sidebarDelegate struct {
	list.DefaultDelegate
}

func (d sidebarDelegate) Height() int { return 2 }

func (d sidebarDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	i, ok := item.(sidebarItem)
	if !ok {
		return
	}

	var (
		title       = i.name
		owner       = i.ownerName
		plType      = i.plType
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
		owner = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#1db954")).
			Render(owner)
		plType = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#1db954")).
			Render(plType)
	} else {
		title = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FAFAFA")).
			Render(title)
		owner = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262")).
			Render(owner)
		plType = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262")).
			Render(plType)

	}

	fmt.Fprintf(w, s.Render(selectedStr+" "+title+"\n  "+plType+" - "+owner))
}

// ---------------------------------------------------------------------
// 3. Sidebar component
// ---------------------------------------------------------------------
type Sidebar struct {
	list            list.Model
	focused         bool
	bus             *MessageBus
	playlistService *service.PlaylistService
}

// NewSidebar creates a ready-to-use sidebar
func NewSidebar(bus *MessageBus, playlistService *service.PlaylistService) *Sidebar {
	const width = 22
	playlists, _ := playlistService.GetPlaylists()
	lists := make([]list.Item, len(playlists))
	for i, p := range playlists {
		lists[i] = sidebarItem{
			name:      p.Name,
			ownerName: p.OwnerName,
			plType:    p.Type,
			id:        p.ID,
		}
	}

	delegate := sidebarDelegate{list.NewDefaultDelegate()}
	l := list.New(lists, delegate, width, 0)
	l.Title = "Playlists"
	l.SetShowHelp(false)
	l.SetShowStatusBar(false)

	return &Sidebar{list: l, focused: false, bus: bus, playlistService: playlistService}
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

	switch m := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(m, defaultKeyMap.Tab),
			key.Matches(m, defaultKeyMap.ShiftTab):
			s.list.Select(-1)
			return s, nil

		case key.Matches(m, key.NewBinding(key.WithKeys("enter"))):
			if sel := s.list.SelectedItem(); sel != nil {
				if item, ok := sel.(sidebarItem); ok && item.id != "" {
					// Publish using your MessageBus API: (type, payload)
					s.bus.Publish(MsgPlaylistSelected, item.id)
				}
			}
			return s, nil
		}
	}

	// Let the list handle navigation, selection, etc.
	s.list, cmd = s.list.Update(msg)
	return s, cmd
}

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
		border = border.BorderForeground(lipgloss.Color("#1db954"))
	}

	return border.Render(s.list.View())
}
