package ui

import (
	"fmt"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/ElshadHu/systui/internal/metrics"
)

const (
	refreshInterval = time.Second
	defaultBarWidth = 40
)

// tickMsg is sent by the tick command once per refreshInterval
type tickMsg time.Time

// Model is the Bubble Tea model for the whole TUI
type Model struct {
	sampler  *metrics.CPUSampler
	cpuPct   float64
	err      error
	barWidth int
}

// New returns a Model ready to be passed to tea.NewProgram
func New() (Model, error) {
	sampler, err := metrics.NewCPUSampler()
	if err != nil {
		return Model{}, err
	}
	return Model{
		sampler:  sampler,
		barWidth: defaultBarWidth,
	}, nil
}

// tick scheduled the next tickMsg
func tick() tea.Cmd {
	return tea.Tick(refreshInterval, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m Model) Init() tea.Cmd {
	return tick()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tickMsg:
		pct, err := m.sampler.Usage()
		m.err = err
		if err == nil {
			m.cpuPct = pct
		}
		return m, tick()
	case tea.KeyPressMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m Model) View() tea.View {
	var b strings.Builder

	b.WriteString("systui - press q to quit\n\n")

	if m.err != nil {
		fmt.Fprintf(&b, "CPU  error: %v\n", m.err)
	} else {
		fmt.Fprintf(&b, "CPU %s %5.1f%%\n", renderBar(m.cpuPct, m.barWidth), m.cpuPct)
	}

	v := tea.NewView(b.String())
	v.AltScreen = true
	return v
}
