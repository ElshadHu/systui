package metrics

import (
	"errors"
	"os"
	"runtime"
	"testing"
)

func TestSampleIncludesOwnProcess(t *testing.T) {
	s := NewProcessSampler()
	stats, err := s.Sample()
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range stats {
		if int(p.PID) == os.Getpid() {
			if p.Name == "" || p.User == "" || p.Started.IsZero() {
				t.Fatalf("own process row is incomplete: %+v", p)
			}
			return
		}
	}
	t.Fatalf("own PID %d missing from %d rows", os.Getpid(), len(stats))
}

func TestSampleDropsExitedProcesses(t *testing.T) {
	s := NewProcessSampler()
	s.tracked[-1] = nil
	if _, err := s.Sample(); err != nil {
		t.Fatal(err)
	}
	if _, ok := s.tracked[-1]; ok {
		t.Fatal("PID -1 should have been pruned")
	}
}

func TestSignalRefusesOwnProcess(t *testing.T) {
	if err := KillProcess(int32(os.Getpid())); err != ErrOwnProcess {
		t.Fatalf("want ErrOwnProcess, got %v", err)
	}
}

func TestReadProcessDetailsOwnProcess(t *testing.T) {
	d, err := ReadProcessDetails(int32(os.Getpid()))
	if err != nil {
		t.Fatal(err)
	}
	if d.Name == "" || d.Threads == 0 || d.RSS == 0 || d.ParentPID == 0 {
		t.Fatalf("details are incomplete: %+v", d)
	}
}

func TestParseLsofNames(t *testing.T) {
	out := "p123\nfcwd\nn/Users/me\nftxt\nn/usr/bin/node\nf3\nn/Users/me\nf4\nn/tmp/log.txt\n"
	got := parseLsofNames(out)
	want := []string{"/Users/me", "/usr/bin/node", "/tmp/log.txt"}
	if len(got) != len(want) {
		t.Fatalf("want %v, got %v", want, got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("want %v, got %v", want, got)
		}
	}
}

func TestReadOpenFilesOwnProcess(t *testing.T) {
	files, err := ReadOpenFiles(int32(os.Getpid()))
	if err != nil {
		t.Fatal(err)
	}
	if len(files) == 0 {
		t.Fatal("expected at least one open file")
	}
}

func TestReadProcessDetailsRecordsUnreadableFields(t *testing.T) {
	if runtime.GOOS != "darwin" || os.Geteuid() == 0 {
		t.Skip("needs a process owned by another user without root")
	}
	d, err := ReadProcessDetails(1)
	if err != nil {
		t.Fatal(err)
	}
	if len(d.Unreadable) == 0 {
		t.Fatalf("launchd fields should be unreadable without root: %+v", d)
	}
	for _, f := range d.Unreadable {
		if f.Field == "" || f.Err == nil {
			t.Fatalf("incomplete field error: %+v", f)
		}
	}
}

func TestLsofPathsDeniedForOtherUser(t *testing.T) {
	if runtime.GOOS != "darwin" || os.Geteuid() == 0 {
		t.Skip("needs a process owned by another user without root")
	}
	if _, err := lsofPaths(1); !errors.Is(err, ErrLsofDenied) {
		t.Fatalf("want ErrLsofDenied, got %v", err)
	}
}
