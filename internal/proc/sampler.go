package proc

import (
	"os/user"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v4/process"
)

// Stats is one live process
type Stats struct {
	PID     int32
	PPID    int32
	UID     uint32
	Name    string
	Exe     string
	Argv0   string
	Cmdline string
	User    string
	Status  string
	CPU     float64
	Memory  uint64
	Started time.Time
	Lifetime
}

// Lifetime is what a process has accumulated since it started
type Lifetime struct {
	PeakMemory uint64
	CPUTime    time.Duration
}

// usage is what one sample reads for one process
type usage struct {
	Memory uint64
	Lifetime
}

// trackedProcess keeps the fields that never change for one process identity
type trackedProcess struct {
	proc    *process.Process
	created int64
	stats   Stats
}

// Sampler keeps a Process per PID across samples so that
// Percent can measure CPU since the previous call
type Sampler struct {
	mu      sync.Mutex
	tracked map[int32]*trackedProcess
	users   map[uint32]string
}

func NewSampler() *Sampler {
	return &Sampler{
		tracked: make(map[int32]*trackedProcess),
		users:   make(map[uint32]string),
	}
}

// Sample returns every live process
func (s *Sampler) Sample() ([]Stats, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	pids, err := process.Pids()
	if err != nil {
		return nil, err
	}

	seen := make(map[int32]bool, len(pids))
	stats := make([]Stats, 0, len(pids))
	for _, pid := range pids {
		t, ok := s.identify(pid)
		if !ok {
			continue
		}
		seen[pid] = true
		stats = append(stats, s.read(t))
	}

	for pid := range s.tracked {
		if !seen[pid] {
			delete(s.tracked, pid)
		}
	}
	return stats, nil
}

// identify returns the tracked entry for pid
func (s *Sampler) identify(pid int32) (*trackedProcess, bool) {
	fresh, err := process.NewProcess(pid)
	if err != nil {
		return nil, false
	}
	created, err := fresh.CreateTime()
	if err != nil {
		return nil, false
	}
	if t, ok := s.tracked[pid]; ok && t.created == created {
		return t, true
	}
	name, err := fresh.Name()
	if err != nil {
		return nil, false
	}
	t := &trackedProcess{
		proc:    fresh,
		created: created,
		stats: Stats{
			PID:     pid,
			Name:    name,
			Started: time.UnixMilli(created),
		},
	}
	if ppid, err := fresh.Ppid(); err == nil {
		t.stats.PPID = ppid
	}
	if exe, err := fresh.Exe(); err == nil {
		t.stats.Exe = exe
	}
	if args, err := fresh.CmdlineSlice(); err == nil && len(args) > 0 {
		t.stats.Argv0 = args[0]
		t.stats.Cmdline = strings.Join(args, " ")
	}
	t.stats.UID, t.stats.User = s.owner(fresh)
	s.tracked[pid] = t
	return t, true
}

// read refreshes the fields that change between samples
func (s *Sampler) read(t *trackedProcess) Stats {
	row := t.stats
	if cpu, err := t.proc.Percent(0); err == nil {
		row.CPU = cpu
	}
	u := readUsage(t.proc)
	row.Memory, row.Lifetime = u.Memory, u.Lifetime
	if status, err := readStatus(t.proc); err == nil {
		row.Status = status
	}
	return row
}

// portableUsage is what gopsutil can read everywhere
func portableUsage(p *process.Process) usage {
	var u usage
	if info, err := p.MemoryInfo(); err == nil {
		u.Memory = info.RSS
	}
	if times, err := p.Times(); err == nil {
		u.CPUTime = time.Duration((times.User + times.System) * float64(time.Second))
	}
	return u
}

// owner resolves the real UID once and reuses it for later samples
func (s *Sampler) owner(p *process.Process) (uint32, string) {
	uids, err := p.Uids()
	if err != nil || len(uids) == 0 {
		return 0, "-"
	}
	uid := uids[0]
	if name, ok := s.users[uid]; ok {
		return uid, name
	}
	name := strconv.FormatUint(uint64(uid), 10)
	if u, err := user.LookupId(name); err == nil {
		name = u.Username
	}
	s.users[uid] = name
	return uid, name
}
