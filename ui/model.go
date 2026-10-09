package ui

import (
	"fmt"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/ElshadHu/systui/internal/insight"
	"github.com/ElshadHu/systui/internal/metrics"
	"github.com/ElshadHu/systui/internal/units"
	"github.com/ElshadHu/systui/ui/theme"
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
	theme    theme.Theme
	sampler  *metrics.CPUSampler
	cpu      metrics.CPUStats
	mem      metrics.MemStats
	pressure metrics.PressureLevel
	insights []insight.Insight
	err      error
	barWidth int
	width    int
}

// New returns a Model ready to be passed to tea.NewProgram
func New(t theme.Theme) (Model, error) {
	sampler, err := metrics.NewCPUSampler()
	if err != nil {
		return Model{}, err
	}
	return Model{
		theme:    t,
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

// refresh reads new CPU, memory and pressure stats into the model
func (m *Model) refresh() error {
	cpuStats, err := m.sampler.Sample()
	if err != nil {
		return err
	}
	memStats, err := metrics.ReadMem()
	if err != nil {
		return err
	}
	pressure, err := metrics.ReadPressure()
	if err != nil {
		return err
	}
	m.cpu = cpuStats
	m.mem = memStats
	m.pressure = pressure
	m.insights = insight.Top(insight.System(cpuStats.Usage, pressure, memStats.SwapUsed), maxInsights)
	return nil
}

func (m Model) View() tea.View {
	var b strings.Builder

	for _, line := range m.topBar() {
		b.WriteString(line)
		b.WriteString("\n")
	}
	b.WriteString("\n")

	if m.err != nil {
		b.WriteString(m.theme.Critical.Render(fmt.Sprintf("%s error: %v", m.theme.Glyphs.Critical, m.err)))
		b.WriteString("\n")
	} else {
		for _, line := range fitColumns("│", m.width, m.usagePanel(), m.cpuPanel(), m.memPanel()) {
			b.WriteString(line)
			b.WriteString("\n")
		}
	}

	v := tea.NewView(b.String())
	v.AltScreen = true
	v.BackgroundColor = m.theme.Colors.Background
	v.ForegroundColor = m.theme.Colors.Primary
	return v
}

// usagePanel renders the CPU and MEM bars
func (m Model) usagePanel() []string {
	return []string{
		m.theme.Bright.Render("% Usage"),
		m.usageRow("CPU", m.cpu.Usage),
		m.usageRow("MEM", m.mem.UsedPercent),
	}
}

func (m Model) usageRow(label string, pct float64) string {
	return m.theme.Bright.Render(label+":") + " " +
		renderBar(pct, m.barWidth, m.theme) + " " +
		m.theme.Bright.Render(fmt.Sprintf("%5.1f%%", pct))
}

// cpuPanel renders the per-counter CPU breakdown
func (m Model) cpuPanel() []string {
	cells := []string{
		m.pct("user", m.cpu.User), m.pct("sys", m.cpu.System), m.pct("idle", m.cpu.Idle),
		m.pct("nice", m.cpu.Nice), m.pct("iowait", m.cpu.Iowait), m.pct("irq", m.cpu.Irq),
		m.pct("softirq", m.cpu.Softirq), m.pct("steal", m.cpu.Steal), m.pct("guest", m.cpu.Guest),
	}
	return append([]string{m.theme.Bright.Render("CPU")}, grid(cells, 3, cpuColWidth)...)
}

func (m Model) memPanel() []string {
	cells := []string{
		m.kv("total", units.Bytes(m.mem.Total)),
		m.kv("used", units.Bytes(m.mem.Used)),
		m.kv("free", units.Bytes(m.mem.Free)),
		m.kv("active", units.Bytes(m.mem.Active)),
		m.kv("buffers", units.Bytes(m.mem.Buffers)),
		m.kv("cached", units.Bytes(m.mem.Cached)),
	}
	return append([]string{m.theme.Bright.Render("MEM")}, grid(cells, 3, memColWidth)...)
}

func (m Model) pct(label string, v float64) string {
	return m.theme.Bright.Render(fmt.Sprintf("%s: %.1f%%", label, v))
}

func (m Model) kv(label, v string) string {
	return m.theme.Bright.Render(label + ": " + v)
}
