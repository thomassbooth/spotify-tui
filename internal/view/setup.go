package view

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type SetupDoneMsg struct {
	ClientID     string
}

type Setup struct {
	inputs  []textinput.Model
	focused int
	done    bool
}

func NewSetup() *Setup {
	clientID := textinput.New()
	clientID.Placeholder = "Client ID"
	clientID.CharLimit = 100
	clientID.Width = 40
	clientID.Focus()

	return &Setup{
		inputs:  []textinput.Model{clientID},
		focused: 0,
	}
}

func (s *Setup) Init() tea.Cmd {
	return textinput.Blink
}

func (s *Setup) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m := msg.(type) {
	case tea.KeyMsg:
		switch m.String() {
		case "ctrl+c", "q":
			return s, tea.Quit

		case "tab", "down":
			s.inputs[s.focused].Blur()
			s.focused = (s.focused + 1) % len(s.inputs)
			s.inputs[s.focused].Focus()
			return s, textinput.Blink

		case "shift+tab", "up":
			s.inputs[s.focused].Blur()
			s.focused = (s.focused - 1 + len(s.inputs)) % len(s.inputs)
			s.inputs[s.focused].Focus()
			return s, textinput.Blink

		case "enter":
			if s.focused < len(s.inputs)-1 {
				s.inputs[s.focused].Blur()
				s.focused++
				s.inputs[s.focused].Focus()
				return s, textinput.Blink
			}
			// on last field, submit
			clientID := strings.TrimSpace(s.inputs[0].Value())
			if clientID == ""{
				return s, nil
			}
			return s, tea.Quit
		}
	}

	var cmd tea.Cmd
	s.inputs[s.focused], cmd = s.inputs[s.focused].Update(msg)
	return s, cmd
}

func (s *Setup) Credentials() string {
	return strings.TrimSpace(s.inputs[0].Value())
}

func (s *Setup) View() string {
	var sb strings.Builder

	title := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#1db954")).
		Bold(true).
		MarginBottom(1).
		Render("Spotify TUI — First Time Setup")

	instructions := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#b3b3b3")).
		Render(fmt.Sprintf(
			"1. Go to %s\n2. Create an app\n3. Add %s as a Redirect URI\n4. Enter your credentials below",
			"https://developer.spotify.com/dashboard",
			"http://127.0.0.1:8889/callback",
		))

	sb.WriteString(title + "\n")
	sb.WriteString(instructions + "\n\n")

	labels := []string{"Client ID"}
	for i, input := range s.inputs {
		label := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FAFAFA")).
			Render(labels[i])
		sb.WriteString(label + "\n")
		sb.WriteString(input.View() + "\n\n")
	}

	sb.WriteString(lipgloss.NewStyle().
		Foreground(lipgloss.Color("#626262")).
		Render("tab to switch fields • enter to confirm"))

	return lipgloss.NewStyle().Padding(2, 4).Render(sb.String())
}