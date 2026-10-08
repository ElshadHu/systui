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
	defaultBarWidth = 20
	cpuColWidth     = 16
	memColWidth     = 18
)

// tickMsg is sent by the tick command once per refreshInterval
type tickMsg time.Time

// Model is the Bubble Tea model for the whole TUI
type Model struct {
	sampler    *metrics.CPUSampler
	cpu        metrics.CPUStats
	mem        metrics.MemStats
	err        error
	lastUpdate time.Time
	barWidth   int
	width      int
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
		m.lastUpdate = time.Time(msg)
		m.err = m.refresh()
		return m, tick()
	case tea.WindowSizeMsg:
		m.width = msg.Width
		return m, nil
	case tea.KeyPressMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	}
	return m, nil
}

// refresh reads new CPU and mem stats into the model
func (m *Model) refresh() error {
	cpuStats, err := m.sampler.Sample()
	if err != nil {
		return err
	}
	memStats, err := metrics.ReadMem()
	if err != nil {
		return err
	}
	m.cpu = cpuStats
	m.mem = memStats
	return nil
}

func (m Model) View() tea.View {
	var b strings.Builder

	b.WriteString("systui - press q to quit\n")
	if !m.lastUpdate.IsZero() {
		fmt.Fprintf(&b, "Last update: %s\n", m.lastUpdate.Format("15:04:05"))
	}

	b.WriteString("\n")

	if m.err != nil {
		fmt.Fprintf(&b, "CPU  error: %v\n", m.err)
	} else {
		for _, line := range fitColumns("│", m.width, m.usagePanel(), m.cpuPanel(), m.memPanel()) {
			b.WriteString(line)
			b.WriteString("\n")
		}
	}

	v := tea.NewView(b.String())
	v.AltScreen = true
	return v
}

// usagePanel renders the CPU and MEM bars
func (m Model) usagePanel() []string {
	return []string{
		boldWhite("% Usage"),
		usageRow("CPU", m.cpu.Usage, m.barWidth),
		usageRow("MEM", m.mem.UsedPercent, m.barWidth),
	}
}

// usageRow renders label with the text in bright white
func usageRow(label string, pct float64, width int) string {
	return boldWhite(label+":") + " " + renderBar(pct, width) + " " + boldWhite(fmt.Sprintf("%5.1f%%", pct))
}

// cpuPanel renders the per-counter CPU breakdown
func (m Model) cpuPanel() []string {
	cells := []string{
		pct("user", m.cpu.User), pct("sys", m.cpu.System), pct("idle", m.cpu.Idle),
		pct("nice", m.cpu.Nice), pct("iowait", m.cpu.Iowait), pct("irq", m.cpu.Irq),
		pct("softirq", m.cpu.Softirq), pct("steal", m.cpu.Steal), pct("guest", m.cpu.Guest),
	}
	return append([]string{boldWhite("CPU")}, grid(cells, 3, cpuColWidth)...)
}

func (m Model) memPanel() []string {
	cells := []string{
		kv("total", formatBytes(m.mem.Total)),
		kv("used", formatBytes(m.mem.Used)),
		kv("free", formatBytes(m.mem.Free)),
		kv("active", formatBytes(m.mem.Active)),
		kv("buffers", formatBytes(m.mem.Buffers)),
		kv("cached", formatBytes(m.mem.Cached)),
	}
	return append([]string{boldWhite("MEM")}, grid(cells, 3, memColWidth)...)
}

func pct(label string, v float64) string {
	return boldWhite(fmt.Sprintf("%s: %.1f%%", label, v))
}

func kv(label, v string) string {
	return boldWhite(label + ": " + v)
}
