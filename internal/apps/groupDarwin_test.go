//go:build darwin

package apps

import (
	"testing"
	"time"

	"github.com/ElshadHu/systui/internal/metrics"
)

const chrome = "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome"
const chromeRenderer = "/Applications/Google Chrome.app/Contents/Frameworks/Google Chrome Framework.framework/Versions/154/Helpers/Google Chrome Helper (Renderer).app/Contents/MacOS/Google Chrome Helper (Renderer)"

func proc(pid, ppid int32, name, exe string, mem uint64) metrics.ProcessStats {
	return metrics.ProcessStats{PID: pid, PPID: ppid, Name: name, Exe: exe, Memory: mem, UID: 501, User: "me", Started: time.Unix(int64(pid), 0)}
}

func find(groups []Group, id string) *Group {
	for i := range groups {
		if groups[i].ID == id {
			return &groups[i]
		}
	}
	return nil
}

func TestHelpersJoinTheOutermostBundle(t *testing.T) {
	stats := []metrics.ProcessStats{
		proc(1, 0, "launchd", "/sbin/launchd", 10<<20),
		proc(10, 1, "Google Chrome", chrome, 300<<20),
		proc(11, 10, "Google Chrome Helper (Renderer)", chromeRenderer, 800<<20),
	}
	stats[2].Cmdline = chromeRenderer + " --type=renderer --lang=en"
	groups := Build(stats, ByMemory)
	g := find(groups, "/Applications/Google Chrome.app")
	if g == nil || g.Name != "Google Chrome" || g.Kind != App || len(g.Members) != 2 {
		t.Fatalf("want one Chrome group with 2 members, got %+v", groups)
	}
	if g.Memory != 1100<<20 {
		t.Fatalf("group memory should sum members, got %d", g.Memory)
	}
	if g.Members[0].Label != "Renderer" || g.Members[0].Main {
		t.Fatalf("renderer should sort first by memory with its role label, got %+v", g.Members[0])
	}
	if g.Members[1].Label != "Google Chrome" || !g.Members[1].Main {
		t.Fatalf("bundle binary should be the main member, got %+v", g.Members[1])
	}
}

func TestChildWithoutBundleJoinsNearestAncestorBundle(t *testing.T) {
	code := "/Applications/Visual Studio Code.app/Contents/MacOS/Electron"
	plugin := "/Applications/Visual Studio Code.app/Contents/Frameworks/Code Helper (Plugin).app/Contents/MacOS/Code Helper (Plugin)"
	stats := []metrics.ProcessStats{
		proc(1, 0, "launchd", "/sbin/launchd", 10<<20),
		proc(20, 1, "Electron", code, 400<<20),
		proc(21, 20, "Code Helper (Plugin)", plugin, 300<<20),
		proc(22, 21, "node", "/usr/local/bin/node", 100<<20),
	}
	groups := Build(stats, ByMemory)
	g := find(groups, "/Applications/Visual Studio Code.app")
	if g == nil || len(g.Members) != 3 {
		t.Fatalf("node should join VS Code, got %+v", groups)
	}
	if g.Members[1].Label != "Plugin" {
		t.Fatalf("helper name should give the role, got %q", g.Members[1].Label)
	}
}

func TestShellStopsTheWalk(t *testing.T) {
	iterm := "/Applications/iTerm.app/Contents/MacOS/iTerm2"
	stats := []metrics.ProcessStats{
		proc(1, 0, "launchd", "/sbin/launchd", 10<<20),
		proc(30, 1, "iTerm2", iterm, 200<<20),
		proc(31, 30, "zsh", "/bin/zsh", 5<<20),
		proc(32, 31, "node", "/usr/local/bin/node", 600<<20),
	}
	groups := Build(stats, ByMemory)
	g := find(groups, "/usr/local/bin/node")
	if g == nil || g.Kind != Tool || g.Name != "node" {
		t.Fatalf("shell child should be its own group, got %+v", groups)
	}
	if term := find(groups, "/Applications/iTerm.app"); term == nil || len(term.Members) != 1 {
		t.Fatalf("iTerm should keep only itself, got %+v", groups)
	}
}

func TestSystemPathsAndUnreadableRootProcesses(t *testing.T) {
	stats := []metrics.ProcessStats{
		proc(1, 0, "launchd", "/sbin/launchd", 10<<20),
		proc(40, 1, "WindowServer", "/System/Library/PrivateFrameworks/SkyLight.framework/Resources/WindowServer", 500<<20),
		{PID: 41, PPID: 1, Name: "secretd", UID: 0, User: "root"},
		{PID: 42, PPID: 1, Name: "_mdnsresponder", UID: 65, User: "_mdnsresponder"},
	}
	groups := Build(stats, ByMemory)
	g := find(groups, SystemGroupID)
	if g == nil || g.Kind != System || len(g.Members) != 4 {
		t.Fatalf("want 4 system members, got %+v", groups)
	}
}

func TestQuietGroupsMergeIntoSmallBucketLast(t *testing.T) {
	stats := []metrics.ProcessStats{
		proc(50, 1, "big", "/opt/big", 900<<20),
		proc(51, 1, "tiny", "/opt/tiny", 1<<20),
		proc(52, 1, "mini", "/opt/mini", 2<<20),
	}
	groups := Build(stats, ByMemory)
	if len(groups) != 2 || groups[0].ID != "/opt/big" {
		t.Fatalf("want big then the bucket, got %+v", groups)
	}
	small := groups[1]
	if small.Kind != Small || small.Name != "2 small processes" || small.Memory != 3<<20 {
		t.Fatalf("bad small bucket %+v", small)
	}
}

func TestMainIsTheMemberWhoseParentIsOutside(t *testing.T) {
	code := "/Applications/Visual Studio Code.app/Contents/MacOS/Electron"
	plugin := "/Applications/Visual Studio Code.app/Contents/Frameworks/Code Helper (Plugin).app/Contents/MacOS/Code Helper (Plugin)"
	stats := []metrics.ProcessStats{
		proc(20, 1, "Electron", code, 100<<20),
		proc(21, 20, "Code Helper (Plugin)", plugin, 300<<20),
		proc(22, 21, "gopls", "/Users/me/go/bin/gopls", 400<<20),
	}
	g := find(Build(stats, ByMemory), "/Applications/Visual Studio Code.app")
	for _, m := range g.Members {
		if m.Main != (m.PID == 20) {
			t.Fatalf("Electron should be the only main member, got %+v", g.Members)
		}
	}
}

func TestToolGroupIsNamedByTheCommandItRunsAs(t *testing.T) {
	exe := "/Users/me/.local/share/claude/versions/2.1.296"
	stats := []metrics.ProcessStats{
		proc(70, 1, "2.1.296", exe, 250<<20),
		proc(71, 70, "2.1.296", exe, 20<<20),
		proc(72, 1, "-zsh", "/opt/homebrew/bin/zsh", 120<<20),
	}
	stats[0].Argv0 = "claude"
	stats[1].Argv0 = "ugrep"
	stats[2].Argv0 = "-zsh"
	groups := Build(stats, ByMemory)
	g := find(groups, exe)
	if g == nil || g.Name != "claude" || g.Members[1].Label != "ugrep" {
		t.Fatalf("tool group and members should be named by argv[0], got %+v", g)
	}
	if shell := find(groups, "/opt/homebrew/bin/zsh"); shell == nil || shell.Name != "zsh" {
		t.Fatalf("login shell dash should be stripped, got %+v", shell)
	}
}

func TestAppleAppsAreAppsNotSystem(t *testing.T) {
	stats := []metrics.ProcessStats{
		proc(80, 1, "Terminal", "/System/Applications/Utilities/Terminal.app/Contents/MacOS/Terminal", 200<<20),
		proc(81, 1, "Finder", "/System/Library/CoreServices/Finder.app/Contents/MacOS/Finder", 200<<20),
	}
	groups := Build(stats, ByMemory)
	if g := find(groups, "/System/Applications/Utilities/Terminal.app"); g == nil || g.Kind != App {
		t.Fatalf("Terminal should be an app, got %+v", groups)
	}
	if g := find(groups, SystemGroupID); g == nil || len(g.Members) != 1 {
		t.Fatalf("Finder should stay in the system group, got %+v", groups)
	}
}

func TestSortByCPU(t *testing.T) {
	stats := []metrics.ProcessStats{
		proc(60, 1, "heavy", "/opt/heavy", 100<<20),
		proc(61, 1, "busy", "/opt/busy", 60<<20),
	}
	stats[1].CPU = 40
	groups := Build(stats, ByCPU)
	if groups[0].Name != "busy" {
		t.Fatalf("busy should lead by CPU, got %+v", groups)
	}
}
