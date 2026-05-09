package view

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/thomassbooth/spotify-tui/internal/assets"
)

type Navigation struct {
	focused     bool
	searching   bool
	searchInput textinput.Model
	bus         *MessageBus
}

func NewNavigation(bus *MessageBus) *Navigation {
	ti := textinput.New()
	ti.Placeholder = "Search songs, artists..."
	ti.CharLimit = 100
	ti.Width = 30
	self := &Navigation{
		searchInput: ti,
		bus:         bus,
	}
	bus.Subscribe(MsgFocusSearch, self)
	return self
}

func (n *Navigation) OnMessage(t MsgType, msg tea.Msg) tea.Cmd {
	
	if t == MsgFocusSearch {
		return func() tea.Msg {
			return FocusSearchMsg{}
		}
	}

	return nil
}

func (n *Navigation) Update(msg tea.Msg) (Component, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case FocusSearchMsg:
        n.searching = true
        n.searchInput.Focus()
        return n, textinput.Blink

	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			if n.searching {
				n.searching = false
				n.searchInput.Blur()
				n.searchInput.SetValue("")
				return n, nil
			}
		case "enter":
			if n.searching {
				query := n.searchInput.Value()
				if query != "" {
					n.searching = false
					n.searchInput.Blur()
					cmd = n.bus.Publish(MsgSearch, SearchResultsMsg{Query: query})
					return n, cmd
				}
				return n, nil
			}
			// enter while focused but not yet searching — activate search
			n.searching = true
			n.searchInput.Focus()
			return n, textinput.Blink
		}
	}

	if n.searching {
		n.searchInput, cmd = n.searchInput.Update(msg)
	}

	return n, cmd
}

func (n *Navigation) Blur() {
	n.focused = false
	if n.searching {
		n.searching = false
		n.searchInput.Blur()
		n.searchInput.SetValue("")
	}
}

func (n *Navigation) Focus() {
	n.focused = true
}

func (n *Navigation) Focused() bool {
	return n.focused
}
func (n *Navigation) View(width, height int) string {
	halfWidth := width / 2

	logoStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#1db954"))
	logo := logoStyle.Width(halfWidth).Align(lipgloss.Right).PaddingRight(1).Render(assets.SpotifyLogo)

	inputStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Padding(0, 1).
		Width(halfWidth - 4)

	if n.searching {
		inputStyle = inputStyle.BorderForeground(lipgloss.Color("#1db954"))
	} else {
		inputStyle = inputStyle.BorderForeground(lipgloss.Color("#535353"))
	}

	searchInput := inputStyle.Render(n.searchInput.View())
	inputHeight := lipgloss.Height(searchInput)
	topPadding := (height - inputHeight) / 2

	left := lipgloss.NewStyle().
		PaddingLeft(1).
		PaddingTop(topPadding).
		Width(halfWidth).
		Render(searchInput)

	content := lipgloss.JoinHorizontal(lipgloss.Top, left, logo)

	border := borderStyle.Copy().
		Width(width).
		Height(height)

	if n.Focused() {
		border = border.BorderForeground(lipgloss.Color("#1db954"))
	}

	return border.Render(content)
}