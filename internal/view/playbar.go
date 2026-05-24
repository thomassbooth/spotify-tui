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
	bus.Subscribe(MsgToggleRepeat, p)

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
			case "up":
				vol, err := p.playbackService.VolumeUp(p.playbackState.Device.VolumePercent, 10)
				if err != nil {
					return p, nil
				}
				p.playbackState.Device.VolumePercent = vol
				return p, nil
			case "down":
				vol, err := p.playbackService.VolumeDown(p.playbackState.Device.VolumePercent, 10)
				if err != nil {
					return p, nil
				}
				p.playbackState.Device.VolumePercent = vol
				return p, nil
		}

	case schedulePollMsg:
		return p, tea.Tick(m.delay, func(time.Time) tea.Msg {
			state, err := p.playbackService.GetCurrentPlaybackState()
			if err != nil || state == nil {
				return errMsg{Err: err}
			}
			return *state
		})

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
		return schedulePollMsg{delay: 300 * time.Millisecond}
	}
}

func (p *Playbar) nextCmd() tea.Cmd {
	return func() tea.Msg {
		err := p.playbackService.NextTrack()
		if err != nil {
			return errMsg{Err: err}
		}
		return schedulePollMsg{delay: 500 * time.Millisecond}
	}
}

func (p *Playbar) previousCmd() tea.Cmd {
	return func() tea.Msg {
		err := p.playbackService.PreviousTrack()
		if err != nil {
			return errMsg{Err: err}
		}
		return schedulePollMsg{delay: 500 * time.Millisecond}
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
				return schedulePollMsg{delay: 300 * time.Millisecond}
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
			return schedulePollMsg{delay: 300 * time.Millisecond}
		}
	}

	if t == MsgToggleRepeat {
		return func() tea.Msg {
			
			err := p.playbackService.ToggleRepeatPlayback(p.getNextRepeatState())
			if err != nil {
				return errMsg{Err: err}
			}
			return schedulePollMsg{delay: 300 * time.Millisecond}
		}
	}

	return nil
}

func (p *Playbar) getNextRepeatState() string {
	switch p.playbackState.RepeatState {
	case "off":
		return "context"
	case "context":
		return "track"
	case "track":
		return "off"
	default:
		return "context"
	}
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
			Height(height).PaddingLeft(1).
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

	//shuffle
	shuffleColor := lipgloss.Color("#535353")
	shuffleText := "unshuffled"
	shuffleBold := false
	if state.ShuffleState {
		shuffleColor = lipgloss.Color("#1db954")
		shuffleText = "shuffled"
		shuffleBold = true
	}
	shuffle := lipgloss.NewStyle().Bold(shuffleBold).PaddingLeft(1).Foreground(shuffleColor).Render(shuffleText)

	// repeat
	repeatColor := lipgloss.Color("#535353")
	if state.RepeatState != "off" {
		repeatColor = lipgloss.Color("#1db954")
	}
	repeat := lipgloss.NewStyle().PaddingLeft(1).Foreground(repeatColor).Render("repeat: " + state.RepeatState)

	volumeText := fmt.Sprintf("vol: %d%%", state.Device.VolumePercent)
	volumeColor := getSoundColor(state.Device.VolumePercent)
	volume := lipgloss.NewStyle().PaddingLeft(1).Foreground(volumeColor).Render(volumeText)

	lowerBar := fmt.Sprintf("%s  %s  %s  %s  %s", bar, times, shuffle, repeat, volume)
	content := lipgloss.JoinVertical(lipgloss.Left, song, artist, lowerBar)
	paddedContent := lipgloss.NewStyle().PaddingLeft(2).Render(content)

	b := borderStyle.Copy().Width(width).Height(height)
	if p.focused {
		b = b.BorderForeground(lipgloss.Color("#1db954"))
	}

	return b.Render(paddedContent)
}

func getSoundColor(volume int) lipgloss.Color {
	switch {
	case volume == 0:
		return lipgloss.Color("#808080")
	case volume == 100:
		return lipgloss.Color("#c10000")
	case volume < 40:
		return lipgloss.Color("#6BA3FF")
	case volume < 70:
		return lipgloss.Color("#f8c948")
	default:
		return lipgloss.Color("#f46a1b")
	}
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
	return tea.Tick(3*time.Second, func(time.Time) tea.Msg {
		state, err := svc.GetCurrentPlaybackState()
		if err != nil {
			return syncPollMsg{state: nil}
		}
		return syncPollMsg{state: state}
	})
}