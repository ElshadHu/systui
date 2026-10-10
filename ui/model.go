package ui

import (
	"fmt"
	"os"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/ElshadHu/systui/internal/apps"
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
	topBarLines     = 2
	pageStep        = 10
)

// tickMsg is sent by the tick command once per refreshInterval
type tickMsg time.Time

// processTickMsg asks for the next process sample
type processTickMsg struct{}

// processesMsg carries one finished process sample
type processesMsg struct {
	stats []metrics.ProcessStats
	err   error
}

// Model is the Bubble Tea model for the whole TUI
type Model struct {
	theme    theme.Theme
	sampler  *metrics.CPUSampler
	procs    *metrics.ProcessSampler
	trends   *apps.Trends
	cpu      metrics.CPUStats
	mem      metrics.MemStats
	pressure metrics.PressureLevel
	insights []insight.Insight
	err      error
	procErr  error
	barWidth int
	width    int
	height   int
	view     View
	apps     appsState
	uid      uint32
	now      time.Time
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
		procs:    metrics.NewProcessSampler(),
		trends:   apps.NewTrends(),
		barWidth: defaultBarWidth,
		view:     AppsView,
		uid:      uint32(os.Getuid()),
		now:      time.Now(),
	}, nil
}

// tick scheduled the next tickMsg
func tick() tea.Cmd {
	return tea.Tick(refreshInterval, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func processTick() tea.Cmd {
	return tea.Tick(refreshInterval, func(time.Time) tea.Msg {
		return processTickMsg{}
	})
}

// sampleProcesses runs the slow process walk off the update loop
func sampleProcesses(s *metrics.ProcessSampler) tea.Cmd {
	return func() tea.Msg {
		stats, err := s.Sample()
		return processesMsg{stats: stats, err: err}
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(tick(), sampleProcesses(m.procs))
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tickMsg:
		m.now = time.Time(msg)
		m.err = m.refresh()
		return m, tick()
	case processTickMsg:
		return m, sampleProcesses(m.procs)
	case processesMsg:
		m.now = time.Now()
		m.procErr = msg.err
		if msg.err == nil {
			groups := apps.Build(msg.stats, m.apps.sort)
			m.trends.Record(groups, m.now)
			m.apps.setGroups(groups, m.now)
		}
		return m, processTick()
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case tea.KeyPressMsg:
		return m.handleKey(msg)
	}
	return m, nil
}

func (m Model) handleKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	m.now = time.Now()
	if m.view == AppsView && m.apps.typing {
		m.typeFilter(msg)
		return m, nil
	}
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "1", "2", "3", "4", "5":
		m.view = View(msg.String()[0] - '0')
		return m, nil
	}
	if m.view == AppsView {
		m.appsKey(msg.String())
	}
	return m, nil
}

func (m *Model) appsKey(key string) {
	s := &m.apps
	s.freeze(m.now)
	switch key {
	case "up":
		s.move(-1)
	case "down":
		s.move(1)
	case "pgup", "ctrl+u":
		s.move(-m.listHeight())
	case "pgdown", "ctrl+d":
		s.move(m.listHeight())
	case "g":
		s.moveTo(0)
	case "G":
		s.moveTo(len(s.rows) - 1)
	case "right":
		s.expand()
	case "left":
		s.collapse()
	case "space":
		s.toggle()
	case "enter":
		if r, ok := s.current(); ok && r.kind == moreRow {
			s.expand()
		}
	case "E":
		s.expandAll()
	case "C":
		s.collapseAll()
	case "c":
		s.setSort(apps.ByCPU)
	case "m":
		s.setSort(apps.ByMemory)
	case "/":
		s.typing = true
	case "esc":
		s.setFilter("")
	}
}

func (m *Model) typeFilter(msg tea.KeyPressMsg) {
	s := &m.apps
	switch msg.String() {
	case "esc":
		s.typing = false
		s.setFilter("")
	case "enter":
		s.typing = false
	case "backspace":
		if len(s.filter) > 0 {
			s.setFilter(s.filter[:len(s.filter)-1])
		}
	default:
		if msg.Text != "" {
			s.setFilter(s.filter + msg.Text)
		}
	}
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

// listHeight is the number of lines left for Apps rows under the header
func (m Model) listHeight() int {
	if m.height <= 0 {
		return 0
	}
	return max(m.height-topBarLines-3, 1)
}

func (m Model) View() tea.View {
	lines := m.topBar()
	if m.err != nil {
		lines = append(lines, m.theme.Critical.Render(fmt.Sprintf("%s error: %v", m.theme.Glyphs.Critical, m.err)))
	}

	switch m.view {
	case AppsView:
		body := m.appsView(m.listHeight() + 1)
		shown, _, _ := m.apps.viewport(m.listHeight())
		first, last := m.apps.position(shown)
		lines = append(lines, m.tabBar(first, last, len(m.apps.rows)))
		lines = append(lines, body...)
	case OverviewView:
		lines = append(lines, m.tabBar(0, 0, 0), "")
		lines = append(lines, fitColumns("│", m.width, m.usagePanel(), m.cpuPanel(), m.memPanel())...)
	default:
		lines = append(lines, m.tabBar(0, 0, 0), "", m.theme.Muted.Render(viewNames[m.view]+" is not built yet"))
	}

	for m.height > 0 && len(lines) < m.height-1 {
		lines = append(lines, "")
	}
	lines = append(lines, m.footer())

	v := tea.NewView(strings.Join(lines, "\n"))
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
