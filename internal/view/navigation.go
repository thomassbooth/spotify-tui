package view

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/thomassbooth/spotify-tui/internal/assets"
)

type Navigation struct {
	selected    int
	items       []string
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

	return &Navigation{
		selected:    1,
		items:       []string{"🔍 Search", "🏠 Home", "📚 Browse"},
		focused:     false,
		searchInput: ti,
		bus:         bus,
	}
}

func (n *Navigation) Update(msg tea.Msg) (Component, tea.Cmd) {
	var cmd tea.Cmd

	if n.searching {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "esc":
				n.searching = false
				n.searchInput.Blur()
				n.searchInput.SetValue("")
				return n, nil
			case "enter":
				query := n.searchInput.Value()
				if query != "" {
					n.searching = false
					n.searchInput.Blur()
					cmd = n.bus.Publish(MsgSearch, SearchMsg{Query: query})
					return n, cmd
				}
				return n, nil
			}
		}
		n.searchInput, cmd = n.searchInput.Update(msg)
		return n, cmd
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "left", "h":
			if n.selected > 0 {
				n.selected--
			}
		case "right", "l":
			if n.selected < len(n.items)-1 {
				n.selected++
			}
		case "1":
			n.selected = 0
		case "2":
			n.selected = 1
		case "3":
			n.selected = 2
		case "enter":
			if n.selected == 0 {
				n.searching = true
				n.searchInput.Focus()
				return n, textinput.Blink
			}
		}
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

	var parts []string
	for i, txt := range n.items {
		style := ItemNavStyle.MarginRight(2)
		if i == n.selected {
			style = SelectedNavStyle.MarginRight(2)
		}
		parts = append(parts, style.Render(txt))
	}
	nav := lipgloss.JoinHorizontal(lipgloss.Top, parts...)

	logoStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#1db954"))
	logo := logoStyle.Width(halfWidth).Align(lipgloss.Right).Render(assets.SpotifyLogo)

	var left string
	if n.searching {
		inputStyle := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#1db954")).
			Padding(0, 1).
			Width(halfWidth - 4)
		searchInput := inputStyle.Render(n.searchInput.View())
		left = lipgloss.NewStyle().PaddingLeft(1).Width(halfWidth).Render(
			lipgloss.JoinVertical(lipgloss.Left, nav, searchInput),
		)
	} else {
		left = lipgloss.NewStyle().PaddingLeft(1).Width(halfWidth).Render(nav)
	}
	
	content := lipgloss.JoinHorizontal(lipgloss.Top, left, logo)

	border := borderStyle.Copy().
		Width(width).
		Height(height)

	if n.Focused() {
		border = border.BorderForeground(lipgloss.Color("#1db954"))
	}

	return border.Render(content)
}