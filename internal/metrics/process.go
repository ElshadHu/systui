package metrics

import (
	"os/user"
	"strconv"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v4/mem"
	"github.com/shirou/gopsutil/v4/process"
)

// ProcessStats is one row of the process table
type ProcessStats struct {
	PID     int32
	Name    string
	CPU     float64
	Mem     float64
	User    string
	Started time.Time
}

// ProcessSampler keeps a Process per PID across samples so that
// Percent can measure CPU since the previous call
type ProcessSampler struct {
	mu      sync.Mutex
	tracked map[int32]*process.Process
	users   map[uint32]string
}

func NewProcessSampler() *ProcessSampler {
	return &ProcessSampler{
		tracked: make(map[int32]*process.Process),
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
	vm, err := mem.VirtualMemory()
	if err != nil {
		return nil, err
	}

	seen := make(map[int32]bool, len(pids))
	stats := make([]ProcessStats, 0, len(pids))
	for _, pid := range pids {
		p, ok := s.tracked[pid]
		if !ok {
			p, err = process.NewProcess(pid)
			if err != nil {
				continue
			}
			s.tracked[pid] = p
		}
		row, err := s.read(p, vm.Total)
		if err != nil {
			continue
		}
		seen[pid] = true
		stats = append(stats, row)
	}

	for pid := range s.tracked {
		if !seen[pid] {
			delete(s.tracked, pid)
		}
	}
	return stats, nil
}

func (s *ProcessSampler) read(p *process.Process, memTotal uint64) (ProcessStats, error) {
	name, err := p.Name()
	if err != nil {
		return ProcessStats{}, err
	}
	cpu, err := p.Percent(0)
	if err != nil {
		return ProcessStats{}, err
	}
	memInfo, err := p.MemoryInfo()
	if err != nil {
		return ProcessStats{}, err
	}
	created, err := p.CreateTime()
	if err != nil {
		return ProcessStats{}, err
	}
	return ProcessStats{
		PID:     p.Pid,
		Name:    name,
		CPU:     cpu,
		Mem:     memPercent(memInfo.RSS, memTotal),
		User:    s.username(p),
		Started: time.UnixMilli(created),
	}, nil
}

// username resolves the real UID once and reuses it for later samples
func (s *ProcessSampler) username(p *process.Process) string {
	uids, err := p.Uids()
	if err != nil || len(uids) == 0 {
		return "-"
	}
	uid := uids[0]
	if name, ok := s.users[uid]; ok {
		return name
	}
	name := strconv.FormatUint(uint64(uid), 10)
	if u, err := user.LookupId(name); err == nil {
		name = u.Username
	}
	s.users[uid] = name
	return name
}

func memPercent(rss, total uint64) float64 {
	if total == 0 {
		return 0
	}
	return float64(rss) / float64(total) * 100
}
