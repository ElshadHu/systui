//go:build !darwin && !windows

package apps

import "path/filepath"

const (
	SystemGroupName = "Linux system"
	walkToTop       = true
)

var systemPrefixes = []string{
	"/usr/lib/systemd/",
	"/lib/systemd/",
	"/usr/libexec/",
	"/usr/sbin/",
	"/sbin/",
}

// appOf treats every executable as a possible app root; the parent walk
// then climbs to the top-most one below the session
func appOf(exe string) (id, name string, ok bool) {
	return exe, filepath.Base(exe), true
}

func isSystemApp(exe string) bool {
	return false
}
