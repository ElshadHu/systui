package applist

import (
	"strings"
	"time"

	"github.com/ElshadHu/systui/internal/apps"
)

const (
	maxShownMembers = 5
	sortPause       = 2 * time.Second
)

type RowKind int

const (
	GroupRow RowKind = iota
	MemberRow
	MoreRow
)

// RowID names a row by what it shows, never by its position
type RowID struct {
	Group string
	PID   int32
	More  bool
}

type Row struct {
	ID     RowID
	Kind   RowKind
	Group  *apps.Group
	Member apps.Member
	Hidden int
}

// State is everything the Apps list remembers between refreshes
type State struct {
	groups      []apps.Group
	Rows        []Row
	Selected    RowID
	offset      int
	sort        apps.SortKey
	frozenUntil time.Time
	expanded    string
	showAll     string
	allExpanded bool
	filter      string
	Typing      bool
}

// SetGroups installs a fresh sample, keeping the row order while the sort is paused
func (s *State) SetGroups(groups []apps.Group, now time.Time) {
	if now.Before(s.frozenUntil) {
		groups = apps.KeepOrder(s.groups, groups)
	}
	s.groups = groups
	s.rebuild()
}

func (s *State) Sort() apps.SortKey {
	return s.sort
}

func (s *State) SetSort(key apps.SortKey) {
	s.sort = key
	s.frozenUntil = time.Time{}
	apps.Sort(s.groups, key)
	s.rebuild()
}

func (s *State) Freeze(now time.Time) {
	s.frozenUntil = now.Add(sortPause)
}

func (s *State) Frozen(now time.Time) bool {
	return now.Before(s.frozenUntil)
}

func (s *State) Filter() string {
	return s.filter
}

func (s *State) SetFilter(text string) {
	s.filter = text
	s.rebuild()
}

// Open is true when the group's members are shown
func (s *State) Open(groupID string) bool {
	return s.allExpanded || s.expanded == groupID || s.filter != ""
}

// rebuild flattens groups into rows and moves the cursor back onto what it was on
func (s *State) rebuild() {
	was := s.selectedIndex()
	s.Rows = s.Rows[:0]
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
		s.Rows = append(s.Rows, Row{ID: RowID{Group: g.ID}, Kind: GroupRow, Group: g})
		if nameMatch && !s.allExpanded && s.expanded != g.ID {
			continue
		}
		shown := min(len(members), maxShownMembers)
		if s.showAll == g.ID {
			shown = len(members)
		}
		for _, m := range members[:shown] {
			s.Rows = append(s.Rows, Row{ID: RowID{Group: g.ID, PID: m.PID}, Kind: MemberRow, Group: g, Member: m})
		}
		if hidden := len(members) - shown; hidden > 0 {
			s.Rows = append(s.Rows, Row{ID: RowID{Group: g.ID, More: true}, Kind: MoreRow, Group: g, Hidden: hidden})
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
func (s *State) reselect(was int) {
	if len(s.Rows) == 0 {
		s.Selected = RowID{}
		return
	}
	if i := s.indexOf(s.Selected); i >= 0 {
		return
	}
	if i := s.indexOf(RowID{Group: s.Selected.Group}); i >= 0 {
		s.Selected = s.Rows[i].ID
		return
	}
	s.Selected = s.Rows[min(max(was, 0), len(s.Rows)-1)].ID
}

func (s *State) indexOf(id RowID) int {
	for i, r := range s.Rows {
		if r.ID == id {
			return i
		}
	}
	return -1
}

func (s *State) selectedIndex() int {
	return s.indexOf(s.Selected)
}

func (s *State) Current() (Row, bool) {
	i := s.selectedIndex()
	if i < 0 {
		return Row{}, false
	}
	return s.Rows[i], true
}

func (s *State) MoveTo(i int) {
	if len(s.Rows) == 0 {
		return
	}
	s.Selected = s.Rows[min(max(i, 0), len(s.Rows)-1)].ID
}

func (s *State) Move(delta int) {
	s.MoveTo(s.selectedIndex() + delta)
}

func (s *State) Expand() {
	r, ok := s.Current()
	if !ok {
		return
	}
	if r.Kind == MoreRow {
		s.showAll = r.Group.ID
	} else {
		s.expanded = r.Group.ID
	}
	s.rebuild()
}

func (s *State) Collapse() {
	r, ok := s.Current()
	if !ok {
		return
	}
	if s.expanded == r.Group.ID {
		s.expanded = ""
	}
	if s.showAll == r.Group.ID {
		s.showAll = ""
	}
	s.allExpanded = false
	s.Selected = RowID{Group: r.Group.ID}
	s.rebuild()
}

func (s *State) Toggle() {
	r, ok := s.Current()
	if !ok {
		return
	}
	if r.Kind != GroupRow || s.allExpanded || s.expanded == r.Group.ID {
		s.Collapse()
		return
	}
	s.Expand()
}

// Click puts the cursor on a row
func (s *State) Click(id RowID) {
	was := s.Selected
	s.Selected = id
	r, ok := s.Current()
	if !ok {
		return
	}
	switch {
	case r.Kind == MoreRow:
		s.Expand()
	case r.Kind == GroupRow && was == id:
		s.Toggle()
	}
}

func (s *State) ExpandAll() {
	s.allExpanded = true
	s.rebuild()
}

func (s *State) CollapseAll() {
	s.allExpanded = false
	s.expanded = ""
	s.showAll = ""
	if r, ok := s.Current(); ok {
		s.Selected = RowID{Group: r.Group.ID}
	}
	s.rebuild()
}

// Viewport returns the rows that fit in height lines, after sliding the
// window one row at a time so the cursor stays inside it. Lines used by the
// above and below markers come out of height
func (s *State) Viewport(height int) (rows []Row, above, below int) {
	total := len(s.Rows)
	if height <= 0 || total == 0 {
		return s.Rows, 0, 0
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
	return s.Rows[s.offset:end], s.offset, total - end
}

// Position reports the first and last visible row numbers, counted from 1
func (s *State) Position(shown []Row) (first, last int) {
	if len(shown) == 0 {
		return 0, 0
	}
	return s.offset + 1, s.offset + len(shown)
}
