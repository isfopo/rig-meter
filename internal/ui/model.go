// Package ui holds the app-specific terminal UI for rig-meter.
package ui

import (
	"fmt"
	"sync/atomic"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	meter "github.com/isfopo/rig-components/meter"
	"github.com/isfopo/rig-core/audio"
)

// Model is a minimal capture-level meter TUI.
type Model struct {
	eng     *audio.Engine
	initErr error
	stream  *audio.Stream

	peak   *atomic.Int64
	frames *atomic.Int64

	lastPeak   int64
	lastFrames int64

	status string
}

type streamStartedMsg struct {
	eng    *audio.Engine
	stream *audio.Stream
	err    error
}

type tickMsg time.Time

// New returns the meter model.
func New() Model {
	return Model{
		peak:   &atomic.Int64{},
		frames: &atomic.Int64{},
		status: "starting audio...",
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		func() tea.Msg {
			eng, err := audio.NewEngineWithName("rig-meter")
			if err != nil {
				return streamStartedMsg{err: err}
			}
			opts := audio.Options{Channels: 2, SampleRate: 48000, LatencyMS: 10, CaptureName: "rig-meter:in"}
			peak, frames := m.peak, m.frames
			deliver := func(samples []int16) {
				p := audio.Peak(samples)
				for {
					cur := peak.Load()
					if int64(p) <= cur {
						break
					}
					if peak.CompareAndSwap(cur, int64(p)) {
						break
					}
				}
				frames.Add(int64(len(samples) / 2))
			}
			stream, err := eng.OpenCapture(opts, deliver)
			return streamStartedMsg{eng: eng, stream: stream, err: err}
		},
		tickCmd(),
	)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case streamStartedMsg:
		if msg.err != nil {
			m.initErr = msg.err
			m.status = "audio error"
			return m, nil
		}
		m.eng = msg.eng
		m.stream = msg.stream
		m.status = "capturing default input"
		return m, nil

	case tickMsg:
		if m.peak != nil {
			m.lastPeak = m.peak.Load()
			m.lastFrames = m.frames.Load()
			m.peak.Store(0)
			m.frames.Store(0)
		}
		return m, tickCmd()

	case tea.KeyMsg:
		if msg.String() == "q" || msg.String() == "ctrl+c" {
			if m.stream != nil {
				m.stream.Stop()
			}
			if m.eng != nil {
				m.eng.Close()
			}
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m Model) View() string {
	if m.initErr != nil {
		return fmt.Sprintf("rig-meter\n\naudio: %v\n\nq quit\n", m.initErr)
	}
	bar := meter.Bar(int16(m.lastPeak), m.lastFrames, 0)
	return fmt.Sprintf("rig-meter\n\ninput: system default\n%s\n\n%s\n\nq quit\n", bar, m.status)
}

// Shutdown releases audio resources.
func (m Model) Shutdown() {
	if m.stream != nil {
		m.stream.Stop()
	}
	if m.eng != nil {
		m.eng.Close()
	}
}

func tickCmd() tea.Cmd {
	return tea.Tick(200*time.Millisecond, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

var _ = lipgloss.NewStyle
