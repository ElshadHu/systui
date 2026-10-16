//go:build linux

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

// appOf treats every executable as a possible app root
func appOf(exe string) (id, name string, ok bool) {
	return exe, filepath.Base(exe), true
}

func isSystemApp(exe string) bool {
	return false
}
