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
type syncPollMsg struct {
	state *entities.PlaybackState
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
	bus.Subscribe(MsgToggleShuffle, p)

	return p
}

func (p *Playbar) Init() tea.Cmd {
	return tea.Batch(p.fetchPlayback(), startSyncPoll(p.playbackService))
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
		p.mu.Lock()
		state := p.playbackState
		p.mu.Unlock()

		if state == nil || !state.IsPlaying {
			p.ticking = false
			return p, nil
		}

		if p.elapsedMs >= state.Track.DurationMs {
			p.ticking = false
			return p, p.fetchPlayback()
		}

		p.elapsedMs += 1000
		return p, p.tickCmd()

	case syncPollMsg:
		var cmds []tea.Cmd
		cmds = append(cmds, startSyncPoll(p.playbackService))

		if m.state == nil {
			return p, tea.Batch(cmds...)
		}

		p.mu.Lock()
		current := p.playbackState
		changed := current == nil ||
			current.Track.ID != m.state.Track.ID ||
			current.IsPlaying != m.state.IsPlaying ||
			current.ShuffleState != m.state.ShuffleState
		if changed {
			p.playbackState = m.state
			p.elapsedMs = m.state.ProgressMs
		}
		p.mu.Unlock()

		if changed && m.state.IsPlaying && !p.ticking {
			p.ticking = true
			cmds = append(cmds, p.tickCmd())
		}
		return p, tea.Batch(cmds...)

	case playbarSyncMsg:
		p.mu.Lock()
		p.playbackState = &m.state
		p.elapsedMs = m.state.ProgressMs
		p.mu.Unlock()
		if m.state.IsPlaying && !p.ticking {
			p.ticking = true
			return p, p.tickCmd()
		}
		return p, nil

	case entities.PlaybackState:
		p.mu.Lock()
		p.playbackState = &m
		p.elapsedMs = m.ProgressMs
		p.mu.Unlock()
		if m.IsPlaying && !p.ticking {
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
			return errMsg{Err: err}
		}

		time.Sleep(300 * time.Millisecond)
		state, err := p.playbackService.GetCurrentPlaybackState()
		if err != nil {
			return errMsg{Err: err}
		}
		if state != nil {
			return *state
		}
		return nil
	}
}

func (p *Playbar) nextCmd() tea.Cmd {
	return func() tea.Msg {
		err := p.playbackService.NextTrack()
		if err != nil {
			return errMsg{Err: err}
		}
		time.Sleep(500 * time.Millisecond)
		state, err := p.playbackService.GetCurrentPlaybackState()
		if err != nil {
			return errMsg{Err: err}
		}
		if state != nil {
			return *state
		}
		return nil
	}
}

func (p *Playbar) previousCmd() tea.Cmd {
	return func() tea.Msg {
		err := p.playbackService.PreviousTrack()
		if err != nil {
			return errMsg{Err: err}
		}
		time.Sleep(500 * time.Millisecond)
		state, err := p.playbackService.GetCurrentPlaybackState()
		if err != nil {
			return errMsg{Err: err}
		}
		if state != nil {
			return *state
		}
		return nil
	}
}

func (p *Playbar) OnMessage(t MsgType, msg tea.Msg) tea.Cmd {
	if t == MsgPlayTrack {
		if playTrackMsg, ok := msg.(PlayTrackMsg); ok {
			return func() tea.Msg {
				err := p.playbackService.Play(playTrackMsg.TrackURI, playTrackMsg.PlaylistURI)
				if err != nil {
					return errMsg{Err: err}
				}
				time.Sleep(300 * time.Millisecond)
				state, err := p.playbackService.GetCurrentPlaybackState()
				if err != nil {
					return errMsg{Err: err}
				}
				return playbarSyncMsg{state: *state}
			}
		}
	}

	if t == MsgToggleShuffle {
		return func() tea.Msg {
			p.mu.Lock()
			currentShuffle := p.playbackState != nil && p.playbackState.ShuffleState
			p.mu.Unlock()
			
			err := p.playbackService.ToggleShufflePlayback(!currentShuffle)
			if err != nil {
				return errMsg{Err: err}
			}
			time.Sleep(300 * time.Millisecond)
			state, err := p.playbackService.GetCurrentPlaybackState()
			if err != nil {
				return errMsg{Err: err}
			}
			if state != nil {
				return playbarSyncMsg{state: *state}
			}
			return nil
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

	shuffleColor := lipgloss.Color("#535353")
	if state.ShuffleState {
		shuffleColor = lipgloss.Color("#1db954")
	}
	shuffle := lipgloss.NewStyle().Foreground(shuffleColor).Render("⇄")

	progress := fmt.Sprintf("%s %s %s", bar, times, shuffle)

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
	filled := int(fraction * float64(width))
	if filled > width {
		filled = width
	}
	empty := width - filled

	return strings.Repeat("━", filled) + strings.Repeat("─", empty)
}

func formatDuration(ms int) string {
	seconds := ms / 1000
	minutes := seconds / 60
	seconds = seconds % 60
	return fmt.Sprintf("%d:%02d", minutes, seconds)
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