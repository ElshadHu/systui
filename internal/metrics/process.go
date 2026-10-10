package metrics

import (
	"os/user"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v4/process"
)

// ProcessStats is one live process
type ProcessStats struct {
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
}

// trackedProcess keeps the fields that never change for one process identity
type trackedProcess struct {
	proc    *process.Process
	created int64
	stats   ProcessStats
}

// ProcessSampler keeps a Process per PID across samples so that
// Percent can measure CPU since the previous call
type ProcessSampler struct {
	mu      sync.Mutex
	tracked map[int32]*trackedProcess
	users   map[uint32]string
}

func NewProcessSampler() *ProcessSampler {
	return &ProcessSampler{
		tracked: make(map[int32]*trackedProcess),
		users:   make(map[uint32]string),
	}
}

// Sample returns every live process
func (s *ProcessSampler) Sample() ([]ProcessStats, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	pids, err := process.Pids()
	if err != nil {
		return nil, err
	}

	seen := make(map[int32]bool, len(pids))
	stats := make([]ProcessStats, 0, len(pids))
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
func (s *ProcessSampler) identify(pid int32) (*trackedProcess, bool) {
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
		stats: ProcessStats{
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
func (s *ProcessSampler) read(t *trackedProcess) ProcessStats {
	row := t.stats
	if cpu, err := t.proc.Percent(0); err == nil {
		row.CPU = cpu
	}
	row.Memory = processMemory(t.proc)
	if status, err := readStatus(t.proc); err == nil {
		row.Status = status
	}
	return row
}

// processMemory prefers the platform's footprint and falls back to resident size
func processMemory(p *process.Process) uint64 {
	if n, ok := readFootprint(p.Pid); ok {
		return n
	}
	info, err := p.MemoryInfo()
	if err != nil {
		return 0
	}
	return info.RSS
}

// owner resolves the real UID once and reuses it for later samples
func (s *ProcessSampler) owner(p *process.Process) (uint32, string) {
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
