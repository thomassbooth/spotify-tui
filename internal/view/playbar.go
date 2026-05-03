package view

import (
	"fmt"
	"strings"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/thomassbooth/spotify-tui/internal/entities"
	"github.com/thomassbooth/spotify-tui/internal/service"
)

type playbarTickMsg struct{}
type playbarSyncMsg struct {
	state entities.PlaybackState
}

type Playbar struct {
	bus             *MessageBus
	playbackService *service.PlaybackService
	playbackState   *entities.PlaybackState
	focused         bool
	width           int
	elapsedMs       int
	mu              sync.Mutex
	ticking         bool
}

func NewPlaybar(bus *MessageBus, playbackService *service.PlaybackService) *Playbar {
	p := &Playbar{
		bus:             bus,
		playbackService: playbackService,
	}

	bus.Subscribe(MsgPlaybackUpdate, p)
	bus.Subscribe(MsgPlayTrack, p)

	return p
}

func (p *Playbar) Init() tea.Cmd {
	return p.fetchPlayback()
}

func (p *Playbar) Update(msg tea.Msg) (Component, tea.Cmd) {
	switch m := msg.(type) {
	case tea.KeyMsg:
		if !p.focused {
			return p, nil
		}
		switch m.String() {
		case " ", "enter":
			return p, p.togglePlayCmd()
		case "n", "l", "right":
			return p, p.nextCmd()
		case "p", "h", "left":
			return p, p.previousCmd()
		}

	case playbarTickMsg:
		if p.playbackState == nil || !p.playbackState.IsPlaying {
			p.ticking = false
			return p, nil
		}

		if p.elapsedMs >= p.playbackState.Track.DurationMs {
			return p, p.fetchPlayback()
		}

		p.elapsedMs += 1000
		return p, p.tickCmd()

	case playbarSyncMsg:
		p.mu.Lock()
		p.playbackState = &m.state
		p.elapsedMs = m.state.ProgressMs
		p.mu.Unlock()
		if p.playbackState.IsPlaying && !p.ticking {
			p.ticking = true
			return p, p.tickCmd()
		}
		return p, nil

	case entities.PlaybackState:
		p.mu.Lock()
		p.playbackState = &m
		p.elapsedMs = m.ProgressMs
		p.mu.Unlock()
		if p.playbackState.IsPlaying && !p.ticking {
			p.ticking = true
			return p, p.tickCmd()
		}
		return p, nil

	}

	return p, nil
}

func (p *Playbar) tickCmd() tea.Cmd {
	return tea.Tick(1*time.Second, func(time.Time) tea.Msg {
		return playbarTickMsg{}
	})
}

func (p *Playbar) fetchPlayback() tea.Cmd {
	return func() tea.Msg {
		state, err := p.playbackService.GetCurrentPlaybackState()
		if err != nil || state == nil {
			return playbarTickMsg{}
		}
		return *state
	}
}

func (p *Playbar) togglePlayCmd() tea.Cmd {
	return func() tea.Msg {
		p.mu.Lock()
		isPlaying := p.playbackState != nil && p.playbackState.IsPlaying
		p.mu.Unlock()

		var err error
		if isPlaying {
			err = p.playbackService.PausePlayback()
		} else {
			err = p.playbackService.ResumePlayback()
		}

		if err != nil {
			return nil
		}

		time.Sleep(300 * time.Millisecond)
		state, _ := p.playbackService.GetCurrentPlaybackState()
		if state != nil {
			return *state
		}
		return nil
	}
}

func (p *Playbar) nextCmd() tea.Cmd {
	return func() tea.Msg {
		p.playbackService.NextTrack()
		time.Sleep(500 * time.Millisecond)
		state, _ := p.playbackService.GetCurrentPlaybackState()
		if state != nil {
			return *state
		}
		return nil
	}
}

func (p *Playbar) previousCmd() tea.Cmd {
	return func() tea.Msg {
		p.playbackService.PreviousTrack()
		time.Sleep(500 * time.Millisecond)
		state, _ := p.playbackService.GetCurrentPlaybackState()
		if state != nil {
			return *state
		}
		return nil
	}
}


func (p *Playbar) OnMessage(t MsgType, msg tea.Msg) tea.Cmd {
	if t == MsgPlayTrack {
		if playTrackMsg, ok := msg.(PlayTrackMsg); ok {
			// play then get the updated state to update
			return func() tea.Msg {
				err := p.playbackService.Play(playTrackMsg.TrackURI, playTrackMsg.PlaylistURI)
				time.Sleep(300 * time.Millisecond)
				state, err := p.playbackService.GetCurrentPlaybackState()
				if err != nil {
					return errMsg{Err: err}
				}
				return playbarSyncMsg{state: *state}
			}
		}
	}

	if t == MsgPlaybackUpdate {
		if state, ok := msg.(entities.PlaybackState); ok {
			return func() tea.Msg {
				return playbarSyncMsg{state: state}
			}
		}
	}
	return nil
}

func (p *Playbar) Blur() {
	p.focused = false
}

func (p *Playbar) Focus() {
	p.focused = true
}

func (p *Playbar) Focused() bool {
	return p.focused
}

func (p *Playbar) View(width, height int) string {
	p.width = width

	p.mu.Lock()
	state := p.playbackState
	elapsed := p.elapsedMs
	p.mu.Unlock()

	if state == nil || state.Track.ID == "" {
		return borderStyle.Copy().
			Width(width).
			Height(height).
			Render("Nothing playing right now")
	}

	track := state.Track
	artistNames := make([]string, len(track.Artists))
	for i, a := range track.Artists {
		artistNames[i] = a.Name
	}
	artistStr := strings.Join(artistNames, ", ")

	song := lipgloss.NewStyle().Foreground(lipgloss.Color("#FAFAFA")).Bold(true).Render(track.Name)
	artist := lipgloss.NewStyle().Foreground(lipgloss.Color("#b3b3b3")).Render(artistStr)

	progressWidth := width / 3
	bar := renderProgressBar(elapsed, track.DurationMs, progressWidth)
	times := fmt.Sprintf("%s / %s", formatDuration(elapsed), formatDuration(track.DurationMs))
	progress := fmt.Sprintf("%s %s", bar, times)

	content := lipgloss.JoinVertical(lipgloss.Left, song, artist, progress)
	paddedContent := lipgloss.NewStyle().PaddingLeft(2).Render(content)

	b := borderStyle.Copy().Width(width).Height(height)
	if p.focused {
		b = b.BorderForeground(lipgloss.Color("#1db954"))
	}

	return b.Render(paddedContent)
}

func renderProgressBar(currentMs int, totalMs int, width int) string {
	if totalMs == 0 {
		return ""
	}

	fraction := float64(currentMs) / float64(totalMs)
	barWidth := width

	filled := int(fraction * float64(barWidth))
	if filled > barWidth {
		filled = barWidth
	}
	empty := barWidth - filled

	return strings.Repeat("━", filled) + strings.Repeat("─", empty)
}

func formatDuration(ms int) string {
	seconds := ms / 1000
	minutes := seconds / 60
	seconds = seconds % 60
	return fmt.Sprintf("%d:%02d", minutes, seconds)
}

type syncPollMsg struct {
	state *entities.PlaybackState
}

func startSyncPoll(svc *service.PlaybackService) tea.Cmd {
	return tea.Tick(5*time.Second, func(time.Time) tea.Msg {
		state, err := svc.GetCurrentPlaybackState()
		if err != nil {
			return syncPollMsg{state: nil}
		}
		return syncPollMsg{state: state}
	})
}