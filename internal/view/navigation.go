// navigation.go
package view

import (
	"strings"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var spotifyLogo string

func init() {
	const bt = "`"
	const logo = `
____                      __            ___             ______  __  __   ______     
/\  _BT                   /\ \__  __   /'___\           /\__  _\/\ \/\ \ /\__  _\    
\ \,\L\_\   _____     ___ \ \ ,_\/\_\ /\ \__/  __  __   \/_/\ \/\ \ \ \ \\/_/\ \/    
 \/_\__ \  /\ '__BT  / __BT\\\ \ \/\/\ \\ \ ,__\/\ \/\ \     \ \ \ \ \ \ \ \ \  \ \ \    
   /\ \L\ \\ \ \L\ \/\ \L\ \\ \ \_\ \ \\ \ \_/\ \ \_\ \     \ \ \ \ \ \_\ \  \_\ \__ 
   \ BT____\\ \ ,__/\ \____/ \ \__\\ \_\\ \_\  \/BT____ \     \ \_\ \ \_____\ /\_____\
    \/_____/ \ \ \/  \/___/   \/__/ \/_/ \/_/   BT___/> \     \/_/  \/_____/ \/_____/
              \ \_\                                /\___/                            
               \/_/                                \/__/                              
`
	spotifyLogo = strings.ReplaceAll(logo, "BT", bt)
}

// Navigation component (not a full tea.Model)
type Navigation struct {
	selected int
	items    []string
	focused  bool
}

func NewNavigation() *Navigation {
	return &Navigation{
		selected: 1, // Start with Home selected
		items:    []string{"🔍 Search", "🏠 Home", "📚 Browse"},
		focused:  false,
	}
}

// Update handles messages for the navigation component
func (n *Navigation) Update(msg tea.Msg) (Component, tea.Cmd) {
	var cmd tea.Cmd
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
		}
	}
	return n, cmd

}

func (n *Navigation) Blur() {
	n.focused = false
}
func (n *Navigation) Focus() {
	n.focused = true
}

func (n *Navigation) Focused() bool {
	return n.focused
}

func (n *Navigation) View(width, height int) string {
	var parts []string
	for i, txt := range n.items {
		style := ItemNavStyle
		if i == n.selected {
			style = SelectedNavStyle
		}
		parts = append(parts, style.Render(txt))
	}
	inner := lipgloss.JoinHorizontal(lipgloss.Left, parts...)
	inner = NavBarStyle.Render(inner)

	logoStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#1db954"))
	content := lipgloss.JoinVertical(lipgloss.Top, logoStyle.Render(spotifyLogo), inner)

	border := borderStyle.Copy().
		Width(width).
		Height(height)

	if n.Focused() {
		border = border.BorderForeground(lipgloss.Color("#1db954"))
	}

	return border.Render(content)
}
