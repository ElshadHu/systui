//go:build windows

package apps

import (
	"path/filepath"
	"strings"
)

const (
	SystemGroupName = "Windows system"
	walkToTop       = false
)

var systemPrefixes = []string{
	`C:\Windows\`,
}

func appOf(exe string) (id, name string, ok bool) {
	return exe, strings.TrimSuffix(filepath.Base(exe), filepath.Ext(exe)), true
}

func isSystemApp(exe string) bool {
	return false
}
