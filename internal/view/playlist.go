package view

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/thomassbooth/spotify-tui/internal/service"
)

// --- playlistItem ---

type playlistItem struct {
	name       string
	artists    []string
	id         string
	resultType string // "track", "album", "playlist" — populated during search
}

func (i playlistItem) Title() string       { return i.name }
func (i playlistItem) Description() string { return i.id }
func (i playlistItem) FilterValue() string { return i.name }

// --- playlistDelegate ---

type playlistDelegate struct {
	list.DefaultDelegate
}

func (d playlistDelegate) Height() int  { return 2 }
func (d playlistDelegate) Spacing() int { return 1 }

func (d playlistDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	i, ok := item.(playlistItem)
	if !ok {
		return
	}

	var (
		title       = i.name
		artist      = strings.Join(i.artists, ", ")
		isSelected  = index == m.Index()
		s           = lipgloss.NewStyle().Padding(0, 0, 0, 2)
		selectedStr = " "
	)

	if isSelected {
		selectedStr = ">"
		title = lipgloss.NewStyle().Foreground(lipgloss.Color("#1db954")).Bold(true).Render(title)
		artist = lipgloss.NewStyle().Foreground(lipgloss.Color("#1db954")).Render(artist)
	} else {
		title = lipgloss.NewStyle().Foreground(lipgloss.Color("#FAFAFA")).Render(title)
		artist = lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Render(artist)
	}

	fmt.Fprintf(w, s.Render(selectedStr+" "+title+"\n  "+artist))
}

// --- search filter ---

type searchFilter int

const (
	filterAll       searchFilter = iota
	filterPlaylists
	filterAlbums
	filterSongs
)

var filterLabels = []string{"All", "Playlists", "Albums", "Songs"}

type search struct {
	active   bool
	filter   searchFilter // the applied filter (what the list currently shows)
	cursor   searchFilter // the highlighted filter (moves with ←/→)
	allItems []list.Item
}

func (s *search) filterItems() []list.Item {
	if s.filter == filterAll {
		return s.allItems
	}
	want := map[searchFilter]string{
		filterPlaylists: "playlist",
		filterAlbums:    "album",
		filterSongs:     "track",
	}[s.filter]

	var out []list.Item
	for _, it := range s.allItems {
		if pi, ok := it.(playlistItem); ok && pi.resultType == want {
			out = append(out, it)
		}
	}
	return out
}

func (s *search) renderFilterTabs() string {
	tabs := make([]string, len(filterLabels))
	for i, label := range filterLabels {
		f := searchFilter(i)
		switch {
		case f == s.cursor && f == s.filter:
			tabs[i] = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#1db954")).
				Bold(true).
				Underline(true).
				Render("[ " + label + " ]")
		case f == s.cursor:
			tabs[i] = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FAFAFA")).
				Bold(true).
				Render("[ " + label + " ]")
		case f == s.filter:
			tabs[i] = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#1db954")).
				Underline(true).
				Render(label)
		default:
			tabs[i] = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#888888")).
				Render(label)
		}
	}
	return lipgloss.NewStyle().
		Padding(0, 0, 0, 2).
		Render(strings.Join(tabs, "  "))
}

// --- PlaylistTracks ---

type PlaylistTracks struct {
	tracks          list.Model
	focused         bool
	bus             *MessageBus
	playlistService *service.PlaylistService
	playbackService *service.PlaybackService
	showingQueue    bool
	lastPlaylist    PlaylistSelectedMsg
	search          search
}

func NewPlaylistTracks(bus *MessageBus, playlistService *service.PlaylistService, playbackService *service.PlaybackService) *PlaylistTracks {
	const defaultWidth = 30

	delegate := playlistDelegate{}
	l := list.New([]list.Item{}, delegate, defaultWidth, 0)
	l.Title = "Playlist Tracks"
	l.SetShowHelp(false)
	l.SetShowStatusBar(false)

	self := &PlaylistTracks{
		tracks:          l,
		focused:         false,
		bus:             bus,
		playlistService: playlistService,
		playbackService: playbackService,
	}

	bus.Subscribe(MsgPlaylistSelected, self)
	bus.Subscribe(MsgToggleQueue, self)
	bus.Subscribe(MsgSearch, self)
	return self
}

func (s *PlaylistTracks) OnMessage(t MsgType, msg tea.Msg) tea.Cmd {

	if t == MsgPlaylistSelected {
		playlistMsg, ok := msg.(PlaylistSelectedMsg)
		if !ok {
			return nil
		}
		s.lastPlaylist = playlistMsg
		s.showingQueue = false
		s.search.active = false
		s.tracks.Title = playlistMsg.Name

		return func() tea.Msg {
			tracks, err := s.playlistService.GetPlaylistTracks(playlistMsg.ID)
			if err != nil {
				return errMsg{Err: err}
			}
			return tracksLoadedMsg{tracks: tracks}
		}
	}

	if t == MsgToggleQueue {
		s.showingQueue = !s.showingQueue
		s.search.active = false
		if s.showingQueue {
			s.tracks.Title = "Queued Songs"
			return func() tea.Msg {
				ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
				defer cancel()
				queueTracks, err := s.playbackService.GetQueue(ctx)
				if err != nil {
					return errMsg{Err: err}
				}
				return queueLoadedMsg{tracks: queueTracks}
			}
		} else if s.lastPlaylist.ID != "" {
			s.tracks.Title = s.lastPlaylist.Name
			return func() tea.Msg {
				tracks, err := s.playlistService.GetPlaylistTracks(s.lastPlaylist.ID)
				if err != nil {
					return errMsg{Err: err}
				}
				return tracksLoadedMsg{tracks: tracks}
			}
		}
	}

	if t == MsgSearch {
		if msgSearchQuery, ok := msg.(SearchResultsMsg); ok {
			s.search.active = true
			s.search.filter = filterAll
			s.search.cursor = filterAll
			s.tracks.Title = fmt.Sprintf("Results: %q", msgSearchQuery.Query)

			items := make([]list.Item, 0, len(msgSearchQuery.Tracks)+len(msgSearchQuery.Albums)+len(msgSearchQuery.Playlists))
			for _, tr := range msgSearchQuery.Tracks {
				names := make([]string, len(tr.Artists))
				for j, a := range tr.Artists {
					names[j] = a.Name
				}
				items = append(items, playlistItem{name: tr.Name, artists: names, id: tr.ID, resultType: "track"})
			}
			for _, al := range msgSearchQuery.Albums {
				items = append(items, playlistItem{name: al.Name, artists: []string{al.Name}, id: al.ID, resultType: "album"})
			}
			for _, pl := range msgSearchQuery.Playlists {
				items = append(items, playlistItem{name: pl.Name, id: pl.ID, resultType: "playlist"})
			}

			s.search.allItems = items
			s.tracks.SetItems(items)
			return nil
		}
	}

	return nil
}

func (s *PlaylistTracks) Deselect() {
	s.tracks.Select(-1)
}

func (s *PlaylistTracks) Update(msg tea.Msg) (Component, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tracksLoadedMsg:
		items := make([]list.Item, len(msg.tracks))
		for i, tr := range msg.tracks {
			artistNames := make([]string, len(tr.Artists))
			for j, a := range tr.Artists {
				artistNames[j] = a.Name
			}
			items[i] = playlistItem{name: tr.Name, artists: artistNames, id: tr.ID}
		}
		s.tracks.SetItems(items)
		return s, nil

	case queueLoadedMsg:
		items := make([]list.Item, len(msg.tracks))
		for i, tr := range msg.tracks {
			artistNames := make([]string, len(tr.Artists))
			for j, a := range tr.Artists {
				artistNames[j] = a.Name
			}
			items[i] = playlistItem{name: tr.Name, artists: artistNames, id: tr.ID}
		}
		s.tracks.SetItems(items)
		return s, nil

	case errMsg:
		return s, nil
	}

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

		case key.Matches(msg, key.NewBinding(key.WithKeys("left", "h"))):
			if s.search.active {
				if s.search.cursor > 0 {
					s.search.cursor--
				}
				return s, nil
			}

		case key.Matches(msg, key.NewBinding(key.WithKeys("right", "l"))):
			if s.search.active {
				if int(s.search.cursor) < len(filterLabels)-1 {
					s.search.cursor++
				}
				return s, nil
			}

		case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
			// If cursor is on a different filter, apply it
			if s.search.active && s.search.cursor != s.search.filter {
				s.search.filter = s.search.cursor
				s.tracks.SetItems(s.search.filterItems())
				return s, nil
			}
			// Otherwise play the selected track
			if len(s.tracks.Items()) > 0 {
				selectedTrack := s.tracks.SelectedItem()
				if selectedTrack != nil {
					if item, ok := selectedTrack.(playlistItem); ok {
						playlistURI := ""
						if !s.showingQueue && !s.search.active && s.lastPlaylist.ID != "" {
							playlistURI = s.lastPlaylist.URI
						}
						return s, s.bus.Publish(MsgPlayTrack, PlayTrackMsg{
							TrackURI:    "spotify:track:" + item.id,
							PlaylistURI: playlistURI,
						})
					}
				}
			}
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
	border := borderStyle.Copy().
		Width(width).
		Height(height)

	if s.Focused() {
		border = border.BorderForeground(lipgloss.Color("#1db954"))
	}

	if s.search.active {
		tabBarHeight := 1
		s.tracks.SetSize(width, height-tabBarHeight)
		listView := s.tracks.View()

		firstNL := strings.Index(listView, "\n")
		var inner string
		if firstNL == -1 {
			inner = lipgloss.JoinVertical(lipgloss.Left, listView, s.search.renderFilterTabs())
		} else {
			titleLine := listView[:firstNL]
			rest := listView[firstNL+1:]
			inner = lipgloss.JoinVertical(lipgloss.Left, titleLine, s.search.renderFilterTabs(), rest)
		}
		return border.Render(inner)
	}

	s.tracks.SetSize(width, height)
	return border.Render(s.tracks.View())
}