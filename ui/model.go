package ui

import (
	"fmt"
	"os"
	"slices"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/ElshadHu/systui/internal/apps"
	"github.com/ElshadHu/systui/internal/insight"
	"github.com/ElshadHu/systui/internal/metrics"
	"github.com/ElshadHu/systui/ui/applist"
	"github.com/ElshadHu/systui/ui/theme"
)

const (
	refreshInterval = time.Second
	topBarLines     = 2
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
	history  *apps.Histories
	cpu      metrics.CPUStats
	mem      metrics.MemStats
	pressure metrics.PressureLevel
	insights []insight.Insight
	system   []insight.Insight
	peaks    []insight.Insight
	err      error
	procErr  error
	barWidth int
	width    int
	height   int
	view     View
	apps     applist.State
	uid      uint32
	now      time.Time
	sampling bool
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
		history:  apps.NewHistories(),
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

// processTick fires once per refreshInterval from when the last walk started
func processTick() tea.Cmd {
	return tea.Tick(refreshInterval, func(time.Time) tea.Msg {
		return processTickMsg{}
	})
}

func processTickNow() tea.Cmd {
	return func() tea.Msg { return processTickMsg{} }
}

// sampleProcesses runs the slow process walk off the update loop
func sampleProcesses(s *metrics.ProcessSampler) tea.Cmd {
	return func() tea.Msg {
		stats, err := s.Sample()
		return processesMsg{stats: stats, err: err}
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(tick(), processTickNow())
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tickMsg:
		m.now = time.Time(msg)
		m.err = m.refresh()
		return m, tick()
	case processTickMsg:
		if m.sampling {
			return m, processTick()
		}
		m.sampling = true
		return m, tea.Batch(sampleProcesses(m.procs), processTick())
	case processesMsg:
		m.now = time.Now()
		m.sampling = false
		m.procErr = msg.err
		if msg.err == nil {
			m.installGroups(apps.Build(msg.stats))
		}
		return m, nil
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case tea.KeyPressMsg:
		return m.handleKey(msg)
	case tea.MouseClickMsg:
		return m.handleClick(msg)
	case tea.MouseWheelMsg:
		return m.handleWheel(msg)
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
	m.system = insight.System(cpuStats.Usage, pressure, memStats.SwapUsed)
	m.rankInsights()
	return nil
}

// installGroups records a fresh sample into the history, orders it and refreshes the rows
func (m *Model) installGroups(groups []apps.Group) {
	m.history.Record(groups, m.now)
	apps.Sort(groups, m.apps.Sort())
	m.peaks = insight.Peak(groups)
	m.apps.SetGroups(groups, m.now)
	m.rankInsights()
}

func (m *Model) rankInsights() {
	m.insights = insight.Top(slices.Concat(m.system, m.peaks), maxInsights)
}

// listsApps is true on the tabs that show the grouped process rows
func (m Model) listsApps() bool {
	return m.view == AppsView || m.view == LifetimeView
}

// tabBarLine is the screen line the tab bar is drawn on
func (m Model) tabBarLine() int {
	if m.err != nil {
		return topBarLines + 1
	}
	return topBarLines
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
	case AppsView, LifetimeView:
		body := m.appsView(m.listHeight() + 1)
		shown, _, _ := m.apps.Viewport(m.listHeight())
		first, last := m.apps.Position(shown)
		lines = append(lines, m.tabBar(first, last, len(m.apps.Rows)))
		lines = append(lines, body...)
	case OverviewView:
		lines = append(lines, m.tabBar(0, 0, 0), "")
		lines = append(lines, m.overviewView()...)
	default:
		lines = append(lines, m.tabBar(0, 0, 0), "", m.theme.Muted.Render(viewNames[m.view]+" is not built yet"))
	}

	for m.height > 0 && len(lines) < m.height-1 {
		lines = append(lines, "")
	}
	lines = append(lines, m.footer())

	v := tea.NewView(strings.Join(lines, "\n"))
	v.AltScreen = true
	v.MouseMode = tea.MouseModeCellMotion
	v.BackgroundColor = m.theme.Colors.Background
	v.ForegroundColor = m.theme.Colors.Primary
	return v
}
