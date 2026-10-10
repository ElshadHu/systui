package apps

import (
	"cmp"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/ElshadHu/systui/internal/metrics"
)

type Kind int

const (
	App Kind = iota
	Tool
	System
	Small
)

type SortKey int

const (
	ByMemory SortKey = iota
	ByCPU
)

const (
	SystemGroupID = "system"
	SmallGroupID  = "small"

	smallMemoryLimit = 50 << 20
	smallCPULimit    = 1.0
	maxWalkDepth     = 32
)

// Member is one process inside a group
type Member struct {
	PID     int32
	PPID    int32
	Label   string
	Status  string
	UID     uint32
	CPU     float64
	Memory  uint64
	Exe     string
	Started int64
	Main    bool
}

// Group is one row of the Apps view: an app, a tool, the system or the small bucket
type Group struct {
	ID      string
	Name    string
	Kind    Kind
	CPU     float64
	Memory  uint64
	Members []Member
}

var shells = map[string]bool{
	"sh": true, "bash": true, "zsh": true, "fish": true, "dash": true,
	"ksh": true, "tcsh": true, "csh": true, "nu": true, "elvish": true,
	"login": true,
}

// Build groups processes and sorts groups and members by key, with the
// small bucket always last
func Build(stats []metrics.ProcessStats, key SortKey) []Group {
	byPID := make(map[int32]*metrics.ProcessStats, len(stats))
	for i := range stats {
		byPID[stats[i].PID] = &stats[i]
	}

	groups := make(map[string]*Group)
	cmdlines := make(map[int32]string, len(stats))
	for i := range stats {
		p := &stats[i]
		cmdlines[p.PID] = p.Cmdline
		id, name, kind := rootOf(p, byPID)
		g, ok := groups[id]
		if !ok {
			g = &Group{ID: id, Name: name, Kind: kind}
			groups[id] = g
		}
		g.Members = append(g.Members, Member{
			PID:     p.PID,
			PPID:    p.PPID,
			Label:   labelOf(p),
			Status:  p.Status,
			UID:     p.UID,
			CPU:     p.CPU,
			Memory:  p.Memory,
			Exe:     p.Exe,
			Started: p.Started.UnixMilli(),
		})
	}

	list := make([]Group, 0, len(groups))
	small := Group{ID: SmallGroupID, Kind: Small}
	for _, g := range groups {
		total(g)
		main := markMain(g)
		if g.Kind == Tool {
			g.Name = main.Label
		}
		labelMembers(g, cmdlines)
		if g.Kind != System && g.Memory < smallMemoryLimit && g.CPU < smallCPULimit {
			small.Members = append(small.Members, g.Members...)
			continue
		}
		list = append(list, *g)
	}
	if len(small.Members) > 0 {
		total(&small)
		small.Name = strconv.Itoa(len(small.Members)) + " small processes"
		list = append(list, small)
	}
	Sort(list, key)
	return list
}

// Sort orders groups and their members by key, larger first, small bucket last
func Sort(list []Group, key SortKey) {
	slices.SortStableFunc(list, func(a, b Group) int {
		if a.Kind == Small || b.Kind == Small {
			return cmp.Compare(boolInt(a.Kind == Small), boolInt(b.Kind == Small))
		}
		return cmp.Or(compare(a.CPU, a.Memory, b.CPU, b.Memory, key), strings.Compare(a.Name, b.Name))
	})
	for i := range list {
		sortMembers(list[i].Members, key)
	}
}

func boolInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// rootOf decides which group a process belongs to
func rootOf(p *metrics.ProcessStats, byPID map[int32]*metrics.ProcessStats) (id, name string, kind Kind) {
	if p.Exe == "" {
		if p.UID == 0 || strings.HasPrefix(p.User, "_") {
			return SystemGroupID, SystemGroupName, System
		}
		return "name:" + p.Name, p.Name, Tool
	}
	if isSystemPath(p.Exe) && !isSystemApp(p.Exe) {
		return SystemGroupID, SystemGroupName, System
	}

	var best *metrics.ProcessStats
	cur := p
	for depth := 0; depth < maxWalkDepth; depth++ {
		if _, _, ok := appOf(cur.Exe); ok {
			best = cur
			if !walkToTop {
				break
			}
		}
		parent, ok := byPID[cur.PPID]
		if !ok || parent.PID == cur.PID || parent.PID <= 1 {
			break
		}
		if parent.Exe == "" || shells[parent.Name] || (isSystemPath(parent.Exe) && !isSystemApp(parent.Exe)) {
			break
		}
		cur = parent
	}
	if best != nil {
		id, name, _ := appOf(best.Exe)
		return id, name, App
	}
	return p.Exe, filepath.Base(p.Exe), Tool
}

// labelOf prefers the command a process was started as, since a multicall
func labelOf(p *metrics.ProcessStats) string {
	base := filepath.Base(strings.TrimPrefix(p.Argv0, "-"))
	if p.Argv0 == "" || base == "." || base == "/" {
		return p.Name
	}
	return base
}

func isSystemPath(exe string) bool {
	for _, prefix := range systemPrefixes {
		if strings.HasPrefix(exe, prefix) {
			return true
		}
	}
	return false
}

func total(g *Group) {
	g.CPU, g.Memory = 0, 0
	for _, m := range g.Members {
		g.CPU += m.CPU
		g.Memory += m.Memory
	}
}

// markMain flags the process that owns the group
func markMain(g *Group) Member {
	inside := make(map[int32]bool, len(g.Members))
	for _, m := range g.Members {
		inside[m.PID] = true
	}
	main := -1
	for i, m := range g.Members {
		if inside[m.PPID] {
			continue
		}
		if main < 0 || m.Started < g.Members[main].Started {
			main = i
		}
	}
	if main < 0 {
		main = 0
	}
	g.Members[main].Main = true
	return g.Members[main]
}

func labelMembers(g *Group, cmdlines map[int32]string) {
	if g.Kind != App {
		return
	}
	for i := range g.Members {
		m := &g.Members[i]
		if !m.Main {
			m.Label = roleLabel(m.Label, cmdlines[m.PID])
		}
	}
}

func sortMembers(list []Member, key SortKey) {
	slices.SortStableFunc(list, func(a, b Member) int {
		return cmp.Or(compare(a.CPU, a.Memory, b.CPU, b.Memory, key), cmp.Compare(a.PID, b.PID))
	})
}

// compare orders the larger value first for the chosen key
func compare(aCPU float64, aMem uint64, bCPU float64, bMem uint64, key SortKey) int {
	if key == ByCPU {
		return cmp.Or(cmp.Compare(bCPU, aCPU), cmp.Compare(bMem, aMem))
	}
	return cmp.Or(cmp.Compare(bMem, aMem), cmp.Compare(bCPU, aCPU))
}

// KeepOrder reorders next to match the order prev had
func KeepOrder(prev, next []Group) []Group {
	position := make(map[string]int, len(prev))
	for i, g := range prev {
		position[g.ID] = i
	}
	memberPosition := make(map[string]map[int32]int, len(prev))
	for _, g := range prev {
		slots := make(map[int32]int, len(g.Members))
		for i, m := range g.Members {
			slots[m.PID] = i
		}
		memberPosition[g.ID] = slots
	}
	rank := func(id string) int {
		if id == SmallGroupID {
			return len(prev) + 1
		}
		if i, ok := position[id]; ok {
			return i
		}
		return len(prev)
	}
	slices.SortStableFunc(next, func(a, b Group) int {
		return cmp.Compare(rank(a.ID), rank(b.ID))
	})
	for i := range next {
		slots := memberPosition[next[i].ID]
		slices.SortStableFunc(next[i].Members, func(a, b Member) int {
			ra, oka := slots[a.PID]
			rb, okb := slots[b.PID]
			if !oka {
				ra = len(slots)
			}
			if !okb {
				rb = len(slots)
			}
			return cmp.Compare(ra, rb)
		})
	}
	return next
}
