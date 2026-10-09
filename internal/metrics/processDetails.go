package metrics

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/shirou/gopsutil/v4/net"
	"github.com/shirou/gopsutil/v4/process"
)

// Connection is one socket owned by a process
type Connection struct {
	Protocol string
	Local    string
	Remote   string
	Status   string
}

// FieldError names a detail that could not be read and why
type FieldError struct {
	Field string
	Err   error
}

// ProcessDetails holds the slow to read fields shown in the drawer
type ProcessDetails struct {
	PID        int32
	Name       string
	Cmdline    string
	Status     string
	Started    time.Time
	ParentPID  int32
	ParentName string
	Threads    int32
	OpenFiles  int32
	RSS        uint64
	VMS        uint64
	MemPercent float64
	CPUTime    time.Duration
	Nice       int32
	Ports      []Connection
	Unreadable []FieldError
}

var (
	ErrOwnProcess = errors.New("refusing to signal systui itself")
	ErrLsofDenied = errors.New("not readable, lsof needs root for other users' processes")
	ErrNoData     = errors.New("no data, needs root for other users' processes")
)

// ReadProcessDetails reads everything the drawer shows for one PID.
// A field that cannot be read is listed in Unreadable
func ReadProcessDetails(pid int32) (ProcessDetails, error) {
	p, err := process.NewProcess(pid)
	if err != nil {
		return ProcessDetails{}, err
	}
	name, err := p.Name()
	if err != nil {
		return ProcessDetails{}, err
	}
	d := ProcessDetails{PID: pid, Name: name}

	if cmdline, err := p.Cmdline(); err == nil {
		d.Cmdline = cmdline
	} else {
		d.fail("command", err)
	}
	if status, err := p.Status(); err != nil {
		d.fail("status", err)
	} else if len(status) > 0 {
		d.Status = status[0]
	}
	if created, err := p.CreateTime(); err == nil {
		d.Started = time.UnixMilli(created)
	} else {
		d.fail("started", err)
	}
	if ppid, err := p.Ppid(); err == nil {
		d.ParentPID = ppid
		d.ParentName = parentName(ppid)
	} else {
		d.fail("parent", err)
	}
	if threads, err := p.NumThreads(); err != nil {
		d.fail("threads", err)
	} else if threads == 0 {
		d.fail("threads", ErrNoData)
	} else {
		d.Threads = threads
	}
	if fds, err := p.NumFDs(); err == nil {
		d.OpenFiles = fds
	} else {
		d.fail("open files", err)
	}
	if mem, err := p.MemoryInfo(); err != nil {
		d.fail("memory", err)
	} else if mem.RSS == 0 {
		d.fail("memory", ErrNoData)
	} else {
		d.RSS = mem.RSS
		d.VMS = mem.VMS
	}
	if pct, err := p.MemoryPercent(); err == nil {
		d.MemPercent = float64(pct)
	} else {
		d.fail("memory percent", err)
	}
	if times, err := p.Times(); err == nil {
		d.CPUTime = time.Duration((times.User + times.System) * float64(time.Second))
	} else {
		d.fail("cpu time", err)
	}
	if nice, err := p.Nice(); err == nil {
		d.Nice = nice
	} else {
		d.fail("nice", err)
	}
	if conns, err := p.Connections(); err == nil {
		d.Ports = connections(conns)
	} else {
		d.fail("ports", err)
	}
	return d, nil
}

func (d *ProcessDetails) fail(field string, err error) {
	d.Unreadable = append(d.Unreadable, FieldError{Field: field, Err: err})
}

// parentName returns an empty string when the parent is gone or not readable
func parentName(ppid int32) string {
	parent, err := process.NewProcess(ppid)
	if err != nil {
		return ""
	}
	name, err := parent.Name()
	if err != nil {
		return ""
	}
	return name
}

// ReadOpenFiles lists the paths a process has open, falling back to lsof
func ReadOpenFiles(pid int32) ([]string, error) {
	p, err := process.NewProcess(pid)
	if err != nil {
		return nil, err
	}
	files, err := p.OpenFiles()
	if err != nil {
		return lsofPaths(pid)
	}
	paths := make([]string, 0, len(files))
	for _, f := range files {
		paths = append(paths, f.Path)
	}
	return paths, nil
}

func lsofPaths(pid int32) ([]string, error) {
	out, err := exec.Command("lsof", "-p", strconv.Itoa(int(pid)), "-Fn").Output()
	if err != nil && len(out) == 0 {
		return nil, ErrLsofDenied
	}
	return parseLsofNames(string(out)), nil
}

// parseLsofNames keeps each path once from lsof -Fn output
func parseLsofNames(out string) []string {
	seen := make(map[string]bool)
	var paths []string
	for _, line := range strings.Split(out, "\n") {
		if !strings.HasPrefix(line, "n") || len(line) < 2 {
			continue
		}
		path := line[1:]
		if seen[path] {
			continue
		}
		seen[path] = true
		paths = append(paths, path)
	}
	return paths
}

func connections(conns []net.ConnectionStat) []Connection {
	out := make([]Connection, 0, len(conns))
	for _, c := range conns {
		if c.Laddr.Port == 0 {
			continue
		}
		proto := "tcp"
		if c.Type == 2 {
			proto = "udp"
		}
		conn := Connection{
			Protocol: proto,
			Local:    address(c.Laddr),
			Status:   c.Status,
		}
		if c.Raddr.Port != 0 {
			conn.Remote = address(c.Raddr)
		}
		out = append(out, conn)
	}
	return out
}

func address(a net.Addr) string {
	if a.IP == "" || a.IP == "*" || a.IP == "0.0.0.0" || a.IP == "::" {
		return fmt.Sprintf(":%d", a.Port)
	}
	return fmt.Sprintf("%s:%d", a.IP, a.Port)
}

func KillProcess(pid int32) error {
	return signal(pid, (*process.Process).Kill)
}

func TerminateProcess(pid int32) error {
	return signal(pid, (*process.Process).Terminate)
}

func SuspendProcess(pid int32) error {
	return signal(pid, (*process.Process).Suspend)
}

func ResumeProcess(pid int32) error {
	return signal(pid, (*process.Process).Resume)
}

func signal(pid int32, send func(*process.Process) error) error {
	if int(pid) == os.Getpid() {
		return ErrOwnProcess
	}
	p, err := process.NewProcess(pid)
	if err != nil {
		return err
	}
	return send(p)
}
