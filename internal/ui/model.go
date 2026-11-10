// Package ui holds the app-specific terminal UI for rig-meter.
package ui

import (
	"fmt"
	"math"
	"sync/atomic"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	dropdown "github.com/isfopo/rig-components/dropdown"
	meter "github.com/isfopo/rig-components/meter"
	"github.com/isfopo/rig-core/audio"
)

// Model is a capture-level meter TUI with device and channel selection.
type Model struct {
	eng     *audio.Engine
	initErr error
	stream  *audio.Stream
	seq     *atomic.Int64

	inputs []audio.Device
	inDrop dropdown.Model

	chanDrop    dropdown.Model
	channels    uint32
	pendingPick int // 1=in, 2=chan

	peak   *atomic.Int64
	frames *atomic.Int64
	sumSq  *atomic.Int64

	meter *meter.Meter

	lastPeak   int64
	lastFrames int64
	lastRms    float64

	status string

	// currently focused dropdown (none when both closed)
}

type engineReadyMsg struct {
	eng *audio.Engine
	ins []audio.Device
	err error
}

type captureStartedMsg struct {
	stream *audio.Stream
	err    error
}

type tickMsg time.Time

// channel options
var channelCounts = []uint32{2, 1}
var channelLabels = []string{"2 channels", "1 channel"}

// New returns the meter model. Optional meter options configure the meter's
// scale, colors, width, ruler, etc. Defaults are used when none are given.
func New(opts ...meter.Option) Model {
	return Model{
		peak:     &atomic.Int64{},
		frames:   &atomic.Int64{},
		sumSq:    &atomic.Int64{},
		seq:      new(atomic.Int64),
		channels: 2,
		status:   "starting audio...",
		inDrop:   dropdown.New([]string{"system default"}, 0),
		chanDrop: dropdown.New(channelLabels, 0),
		meter:    meter.New(opts...),
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		func() tea.Msg {
			eng, err := audio.NewEngineWithName("rig-meter")
			if err != nil {
				return engineReadyMsg{err: err}
			}
			ins, err := eng.Inputs()
			return engineReadyMsg{eng: eng, ins: ins, err: err}
		},
		tickCmd(),
	)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case engineReadyMsg:
		if msg.err != nil {
			m.initErr = msg.err
			m.status = "audio error"
			return m, nil
		}
		m.eng = msg.eng
		m.inputs = msg.ins
		opts := []string{"system default"}
		for _, d := range msg.ins {
			opts = append(opts, d.Name)
		}
		m.inDrop = m.inDrop.SetOptions(opts)
		// open default capture
		return m, m.openCaptureCmd()

	case captureStartedMsg:
		if msg.err != nil {
			m.status = fmt.Sprintf("capture error: %v", msg.err)
			return m, nil
		}
		m.stream = msg.stream
		m.status = "capturing"
		return m, nil

	case dropdown.PickedMsg:
		which := m.pendingPick
		m.pendingPick = 0
		if msg.Changed {
			if which == 2 {
				m.channels = channelCounts[m.chanDrop.Current]
			}
			return m, m.openCaptureCmd()
		}
		return m, nil

	case dropdown.ClosedMsg:
		m.pendingPick = 0
		return m, nil

	case tickMsg:
		if m.peak != nil {
			m.lastPeak = m.peak.Load()
			m.lastFrames = m.frames.Load()
			if m.lastFrames > 0 {
				ss := m.sumSq.Load()
				n := m.lastFrames * int64(m.channels)
				if n > 0 {
					rms := math.Sqrt(float64(ss) / float64(n))
					m.lastRms = 20 * math.Log10(rms/32768.0)
					if rms == 0 {
						m.lastRms = -100
					}
				}
			}
			m.peak.Store(0)
			m.frames.Store(0)
			m.sumSq.Store(0)
		}
		return m, tickCmd()

	case tea.KeyMsg:
		// route to focused dropdown first
		if m.inDrop.Focus {
			m.pendingPick = 1
			nm, cmd := m.inDrop.Update(msg)
			m.inDrop = nm.(dropdown.Model)
			return m, cmd
		}
		if m.chanDrop.Focus {
			m.pendingPick = 2
			nm, cmd := m.chanDrop.Update(msg)
			m.chanDrop = nm.(dropdown.Model)
			return m, cmd
		}
		switch msg.String() {
		case "i":
			m.chanDrop = m.chanDrop.Close()
			m.inDrop = m.inDrop.Open()
			m.pendingPick = 1
		case "c":
			m.inDrop = m.inDrop.Close()
			m.chanDrop = m.chanDrop.Open()
			m.pendingPick = 2
		case "q", "ctrl+c":
			if m.stream != nil {
				m.stream.Stop()
			}
			if m.eng != nil {
				m.eng.Close()
			}
			return m, tea.Quit
		}
	}

	// update focused dropdowns via their own Update for non-key msgs already handled
	return m, nil
}

func (m Model) openCaptureCmd() tea.Cmd {
	if m.eng == nil {
		return nil
	}
	seq := m.seq.Add(1)
	opts := audio.Options{
		DeviceID:    deviceIDFor(m.inputs, m.inDrop.Current),
		Channels:    m.channels,
		SampleRate:  48000,
		LatencyMS:   10,
		CaptureName: "rig-meter:in",
	}
	peak, frames, sumSq := m.peak, m.frames, m.sumSq

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
		var acc int64
		for _, v := range samples {
			iv := int64(v)
			acc += iv * iv
		}
		sumSq.Add(acc)
		frames.Add(int64(len(samples) / int(opts.Channels)))
	}
	return func() tea.Msg {
		if m.stream != nil {
			m.stream.Stop()
		}
		stream, err := m.eng.OpenCapture(opts, deliver)
		if err != nil {
			return captureStartedMsg{err: err}
		}
		if m.seq.Load() != seq {
			stream.Stop()
			return nil
		}
		return captureStartedMsg{stream: stream}
	}
}

func deviceIDFor(inputs []audio.Device, sel int) []byte {
	if sel <= 0 || sel > len(inputs) {
		return nil
	}
	return inputs[sel-1].ID()
}

func (m Model) deviceName() string {
	if m.inDrop.Current > 0 && m.inDrop.Current-1 < len(m.inputs) {
		return m.inputs[m.inDrop.Current-1].Name
	}
	return "system default"
}

func (m Model) View() string {
	if m.initErr != nil {
		return fmt.Sprintf("rig-meter\n\naudio: %v\n\nq quit\n", m.initErr)
	}
	peakDb := audio.ToDbFS(int16(m.lastPeak))
	mtr := m.meter
	if mtr == nil {
		mtr = meter.New()
	}
	bar := mtr.Render(m.lastRms, peakDb, m.lastFrames, 0)
	deviceName := "system default"
	if m.inDrop.Current > 0 && m.inDrop.Current-1 < len(m.inputs) {
		deviceName = m.inputs[m.inDrop.Current-1].Name
	}
	channels := m.channels
	if channels == 0 {
		channels = 2
	}

	var b string
	b += "rig-meter\n\n"
	b += fmt.Sprintf("input: %s\n", deviceName)
	b += fmt.Sprintf("channels: %d\n", channels)
	b += fmt.Sprintf("raw peak: %d  RMS: %.1f dBFS\n\n", m.lastPeak, m.lastRms)
	b += bar + "\n\n"
	b += m.status + "\n\n"
	if m.inDrop.Focus {
		b += "input device:\n" + m.inDrop.View()
	} else if m.chanDrop.Focus {
		b += "channels:\n" + m.chanDrop.View()
	} else {
		b += "i: input device   c: channels   q: quit\n"
	}
	return b
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
