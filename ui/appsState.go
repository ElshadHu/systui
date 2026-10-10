package ui

import (
	"strings"
	"time"

	"github.com/ElshadHu/systui/internal/apps"
)

const (
	maxShownMembers = 5
	sortPause       = 2 * time.Second
)

type rowKind int

const (
	groupRow rowKind = iota
	memberRow
	moreRow
)

// rowID names a row by what it shows, never by its position
type rowID struct {
	group string
	pid   int32
	more  bool
}

type row struct {
	id     rowID
	kind   rowKind
	group  *apps.Group
	member apps.Member
	hidden int
}

// appsState is everything the Apps view remembers between refreshes
type appsState struct {
	groups      []apps.Group
	rows        []row
	selected    rowID
	offset      int
	sort        apps.SortKey
	frozenUntil time.Time
	expanded    string
	showAll     string
	allExpanded bool
	filter      string
	typing      bool
}

// setGroups installs a fresh sample, keeping the row order while the sort is paused
func (s *appsState) setGroups(groups []apps.Group, now time.Time) {
	if now.Before(s.frozenUntil) {
		groups = apps.KeepOrder(s.groups, groups)
	}
	s.groups = groups
	s.rebuild()
}

func (s *appsState) setSort(key apps.SortKey) {
	s.sort = key
	s.frozenUntil = time.Time{}
	apps.Sort(s.groups, key)
	s.rebuild()
}

func (s *appsState) freeze(now time.Time) {
	s.frozenUntil = now.Add(sortPause)
}

func (s *appsState) frozen(now time.Time) bool {
	return now.Before(s.frozenUntil)
}

// rebuild flattens groups into rows and moves the cursor back onto what it was on
func (s *appsState) rebuild() {
	was := s.selectedIndex()
	s.rows = s.rows[:0]
	needle := strings.ToLower(s.filter)
	for i := range s.groups {
		g := &s.groups[i]
		members := g.Members
		nameMatch := needle == "" || strings.Contains(strings.ToLower(g.Name), needle)
		if !nameMatch {
			members = matchingMembers(members, needle)
			if len(members) == 0 {
				continue
			}
		}
		s.rows = append(s.rows, row{id: rowID{group: g.ID}, kind: groupRow, group: g})
		if nameMatch && !s.allExpanded && s.expanded != g.ID {
			continue
		}
		shown := min(len(members), maxShownMembers)
		if s.showAll == g.ID {
			shown = len(members)
		}
		for _, m := range members[:shown] {
			s.rows = append(s.rows, row{id: rowID{group: g.ID, pid: m.PID}, kind: memberRow, group: g, member: m})
		}
		if hidden := len(members) - shown; hidden > 0 {
			s.rows = append(s.rows, row{id: rowID{group: g.ID, more: true}, kind: moreRow, group: g, hidden: hidden})
		}
	}
	s.reselect(was)
}

func matchingMembers(members []apps.Member, needle string) []apps.Member {
	var out []apps.Member
	for _, m := range members {
		if strings.Contains(strings.ToLower(m.Label), needle) {
			out = append(out, m)
		}
	}
	return out
}

// reselect keeps the cursor on the same identity, climbs to its group when a
// process is gone, and falls back to the old position when the group is gone
func (s *appsState) reselect(was int) {
	if len(s.rows) == 0 {
		s.selected = rowID{}
		return
	}
	if i := s.indexOf(s.selected); i >= 0 {
		return
	}
	if i := s.indexOf(rowID{group: s.selected.group}); i >= 0 {
		s.selected = s.rows[i].id
		return
	}
	s.selected = s.rows[min(max(was, 0), len(s.rows)-1)].id
}

func (s *appsState) indexOf(id rowID) int {
	for i, r := range s.rows {
		if r.id == id {
			return i
		}
	}
	return -1
}

func (s *appsState) selectedIndex() int {
	return s.indexOf(s.selected)
}

func (s *appsState) current() (row, bool) {
	i := s.selectedIndex()
	if i < 0 {
		return row{}, false
	}
	return s.rows[i], true
}

func (s *appsState) moveTo(i int) {
	if len(s.rows) == 0 {
		return
	}
	s.selected = s.rows[min(max(i, 0), len(s.rows)-1)].id
}

func (s *appsState) move(delta int) {
	s.moveTo(s.selectedIndex() + delta)
}

func (s *appsState) expand() {
	r, ok := s.current()
	if !ok {
		return
	}
	if r.kind == moreRow {
		s.showAll = r.group.ID
	} else {
		s.expanded = r.group.ID
	}
	s.rebuild()
}

// collapse closes the group under the cursor and puts the cursor on its row
func (s *appsState) collapse() {
	r, ok := s.current()
	if !ok {
		return
	}
	if s.expanded == r.group.ID {
		s.expanded = ""
	}
	if s.showAll == r.group.ID {
		s.showAll = ""
	}
	s.allExpanded = false
	s.selected = rowID{group: r.group.ID}
	s.rebuild()
}

func (s *appsState) toggle() {
	r, ok := s.current()
	if !ok {
		return
	}
	if r.kind != groupRow || s.allExpanded || s.expanded == r.group.ID {
		s.collapse()
		return
	}
	s.expand()
}

func (s *appsState) expandAll() {
	s.allExpanded = true
	s.rebuild()
}

func (s *appsState) collapseAll() {
	s.allExpanded = false
	s.expanded = ""
	s.showAll = ""
	if r, ok := s.current(); ok {
		s.selected = rowID{group: r.group.ID}
	}
	s.rebuild()
}

func (s *appsState) setFilter(text string) {
	s.filter = text
	s.rebuild()
}

// viewport returns the rows that fit in height lines, after sliding the
// window one row at a time so the cursor stays inside it. Lines used by the
// above and below markers come out of height
func (s *appsState) viewport(height int) (rows []row, above, below int) {
	total := len(s.rows)
	if height <= 0 || total == 0 {
		return s.rows, 0, 0
	}
	sel := max(s.selectedIndex(), 0)
	for range 3 {
		visible := height
		if s.offset > 0 {
			visible--
		}
		if s.offset+visible < total {
			visible--
		}
		visible = max(visible, 1)
		if sel < s.offset {
			s.offset = sel
		} else if sel >= s.offset+visible {
			s.offset = sel - visible + 1
		}
		s.offset = min(max(s.offset, 0), max(total-visible, 0))
	}
	visible := height
	if s.offset > 0 {
		visible--
	}
	if s.offset+visible < total {
		visible--
	}
	end := min(s.offset+max(visible, 1), total)
	return s.rows[s.offset:end], s.offset, total - end
}

// position reports the first and last visible row numbers, counted from 1
func (s *appsState) position(shown []row) (first, last int) {
	if len(shown) == 0 {
		return 0, 0
	}
	return s.offset + 1, s.offset + len(shown)
}
