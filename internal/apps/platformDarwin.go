//go:build darwin

package apps

import (
	"path/filepath"
	"strings"
)

const (
	SystemGroupName = "macOS system"
	bundleSuffix    = ".app"
	walkToTop       = false
)

var systemPrefixes = []string{
	"/System/",
	"/usr/libexec/",
	"/usr/sbin/",
	"/usr/bin/",
	"/sbin/",
	"/bin/",
	"/Library/Apple/",
	"/private/var/db/",
}

// isSystemApp keeps Apple's user-facing apps out of the system group
func isSystemApp(exe string) bool {
	return strings.HasPrefix(exe, "/System/Applications/")
}

// appOf finds the outermost .app bundle in an executable path
func appOf(exe string) (id, name string, ok bool) {
	i := strings.Index(exe, bundleSuffix+"/")
	if i < 0 {
		return "", "", false
	}
	id = exe[:i+len(bundleSuffix)]
	return id, strings.TrimSuffix(filepath.Base(id), bundleSuffix), true
}
