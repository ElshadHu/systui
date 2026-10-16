package ui

import (
	"time"

	tea "charm.land/bubbletea/v2"
)

// handleClick switches tabs, sorts by a header column or moves the cursor
// onto the clicked row
func (m Model) handleClick(msg tea.MouseClickMsg) (tea.Model, tea.Cmd) {
	if msg.Button != tea.MouseLeft {
		return m, nil
	}
	m.now = time.Now()
	if msg.Y == m.tabBarLine() {
		if v, ok := tabAt(msg.X); ok {
			m.view = v
		}
		return m, nil
	}
	if !m.listsApps() || m.procErr != nil {
		return m, nil
	}
	s := &m.apps
	headerLine := m.tabBarLine() + 1
	if msg.Y == headerLine {
		if key, ok := headerSortAt(m.columns(), msg.X); ok {
			s.setSort(key)
		}
		return m, nil
	}
	if r, ok := m.rowAt(msg.Y); ok {
		s.freeze(m.now)
		s.click(r.id)
	}
	return m, nil
}

// handleWheel moves the Apps cursor one row
func (m Model) handleWheel(msg tea.MouseWheelMsg) (tea.Model, tea.Cmd) {
	if !m.listsApps() || m.procErr != nil {
		return m, nil
	}
	m.now = time.Now()
	s := &m.apps
	switch msg.Button {
	case tea.MouseWheelUp:
		s.freeze(m.now)
		s.move(-1)
	case tea.MouseWheelDown:
		s.freeze(m.now)
		s.move(1)
	}
	return m, nil
}

// rowAt finds the Apps row drawn on screen line y, skipping the above marker
func (m Model) rowAt(y int) (row, bool) {
	shown, above, _ := m.apps.viewport(m.listHeight())
	first := m.tabBarLine() + 2
	if above > 0 {
		first++
	}
	i := y - first
	if i < 0 || i >= len(shown) {
		return row{}, false
	}
	return shown[i], true
}
