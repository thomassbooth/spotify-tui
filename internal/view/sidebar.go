// view/sidebar.go
package view

import (
	"fmt"
	"io"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/thomassbooth/spotify-tui/internal/entities"
	"github.com/thomassbooth/spotify-tui/internal/service"
)

var defaultKeyMap = struct {
	Tab      key.Binding
	ShiftTab key.Binding
}{
	Tab:      key.NewBinding(key.WithKeys("tab")),
	ShiftTab: key.NewBinding(key.WithKeys("shift+tab")),
}

type sidebarItem struct {
	name      string
	ownerName string
	ownerID   string
	plType    string
	id        string
	uri       string
}

func (i sidebarItem) Title() string       { return i.name }
func (i sidebarItem) Description() string { return i.ownerName }
func (i sidebarItem) FilterValue() string { return i.name }

type sidebarDelegate struct {
	list.DefaultDelegate
	currentUserID string
}

func (d sidebarDelegate) Height() int  { return 2 }
func (d sidebarDelegate) Spacing() int { return 1 }

func (d sidebarDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	i, ok := item.(sidebarItem)
	if !ok {
		return
	}

	var (
		isSelected  = index == m.Index()
		isOwned     = d.currentUserID != "" && i.ownerID == d.currentUserID
		s           = lipgloss.NewStyle().Padding(0, 0, 0, 2)
		selectedStr = " "
		lock        = ""
	)

	if !isOwned {
		lock = " 🔒"
	}

	title  := i.name + lock
	owner  := i.ownerName
	plType := i.plType

	if isSelected {
		selectedStr = ">"
		titleColor := lipgloss.Color("#1db954")
		if !isOwned {
			titleColor = lipgloss.Color("#888888")
		}
		title = lipgloss.NewStyle().Foreground(titleColor).Bold(true).Render(title)
		owner = lipgloss.NewStyle().Foreground(lipgloss.Color("#1db954")).Render(owner)
		plType = lipgloss.NewStyle().Foreground(lipgloss.Color("#1db954")).Render(plType)
	} else {
		titleColor := lipgloss.Color("#FAFAFA")
		if !isOwned {
			titleColor = lipgloss.Color("#626262")
		}
		title = lipgloss.NewStyle().Foreground(titleColor).Render(title)
		owner = lipgloss.NewStyle().Foreground(lipgloss.Color("#626262")).Render(owner)
		plType = lipgloss.NewStyle().Foreground(lipgloss.Color("#626262")).Render(plType)
	}

	fmt.Fprintf(w, s.Render(selectedStr+" "+title+"\n  "+plType+" - "+owner))
}

type Sidebar struct {
	list            list.Model
	focused         bool
	bus             *MessageBus
	playlistService *service.PlaylistService
	userService     *service.UserService
	currentUser     *entities.User
	delegate        *sidebarDelegate
}

func NewSidebar(bus *MessageBus, playlistService *service.PlaylistService, userService *service.UserService) *Sidebar {
	delegate := &sidebarDelegate{DefaultDelegate: list.NewDefaultDelegate()}
	l := list.New([]list.Item{}, delegate, 22, 0)
	l.Title = "Playlists"
	l.SetShowHelp(false)
	l.SetShowStatusBar(false)

	return &Sidebar{
		list:            l,
		focused:         false,
		bus:             bus,
		playlistService: playlistService,
		userService:     userService,
		delegate:        delegate,
	}
}

func (s *Sidebar) Init() tea.Cmd {
	return func() tea.Msg {
		playlists, err := s.playlistService.GetPlaylists()
		if err != nil {
			return errMsg{Err: err}
		}
		user, err := s.userService.GetCurrentUser()
		if err != nil {
			return errMsg{Err: err}
		}
		return playlistsLoadedMsg{playlists: playlists, user: *user}
	}
}

func (s *Sidebar) Deselect() {
	s.list.Select(-1)
}

func (s *Sidebar) Update(msg tea.Msg) (Component, tea.Cmd) {
	var cmd tea.Cmd

	switch m := msg.(type) {
	case playlistsLoadedMsg:
		s.currentUser = &m.user
		s.delegate.currentUserID = m.user.ID

		items := make([]list.Item, len(m.playlists))
		for i, p := range m.playlists {
			items[i] = sidebarItem{
				name:      p.Name,
				ownerName: p.OwnerName,
				ownerID:   p.OwnerID,
				plType:    p.Type,
				id:        p.ID,
				uri:       p.URI,
			}
		}
		s.list.SetItems(items)
		return s, nil
	}

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
					isOwned := s.currentUser != nil && item.ownerID == s.currentUser.ID
					s.Blur()
					cmd := s.bus.Publish(MsgPlaylistSelected, PlaylistSelectedMsg{
						ID:      item.id,
						Name:    item.name,
						URI:     item.uri,
						IsOwned: isOwned,
					})
					return s, cmd
				}
			}
			return s, nil
		}
	}

	s.list, cmd = s.list.Update(msg)
	return s, cmd
}

var borderStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("#626262"))

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