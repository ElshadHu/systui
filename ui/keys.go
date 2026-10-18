package ui

import (
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/ElshadHu/systui/internal/apps"
	"github.com/ElshadHu/systui/ui/applist"
)

func (m Model) handleKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	m.now = time.Now()
	if m.listsApps() && m.apps.Typing {
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
	if m.listsApps() {
		m.appsKey(msg.String())
	}
	return m, nil
}

func (m *Model) appsKey(key string) {
	s := &m.apps
	s.Freeze(m.now)
	switch key {
	case "up":
		s.Move(-1)
	case "down":
		s.Move(1)
	case "pgup", "ctrl+u":
		s.Move(-m.listHeight())
	case "pgdown", "ctrl+d":
		s.Move(m.listHeight())
	case "g":
		s.MoveTo(0)
	case "G":
		s.MoveTo(len(s.Rows) - 1)
	case "right":
		s.Expand()
	case "left":
		s.Collapse()
	case "space":
		s.Toggle()
	case "enter":
		if r, ok := s.Current(); ok && r.Kind == applist.MoreRow {
			s.Expand()
		}
	case "E":
		s.ExpandAll()
	case "C":
		s.CollapseAll()
	case "c":
		s.SetSort(apps.ByCPU)
	case "m":
		s.SetSort(apps.ByMemory)
	case "p":
		s.SetSort(apps.ByPeak)
	case "t":
		s.SetSort(apps.ByCPUTime)
	case "/":
		s.Typing = true
	case "esc":
		s.SetFilter("")
	}
}

func (m *Model) typeFilter(msg tea.KeyPressMsg) {
	s := &m.apps
	switch msg.String() {
	case "esc":
		s.Typing = false
		s.SetFilter("")
	case "enter":
		s.Typing = false
	case "backspace":
		if filter := s.Filter(); len(filter) > 0 {
			s.SetFilter(filter[:len(filter)-1])
		}
	default:
		if msg.Text != "" {
			s.SetFilter(s.Filter() + msg.Text)
		}
	}
}
