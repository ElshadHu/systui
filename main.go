package main

import (
	"flag"
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
	"github.com/ElshadHu/systui/ui"
	"github.com/ElshadHu/systui/ui/theme"
)

func main() {
	themeName := flag.String("theme", "dark", "color theme: dark, light or ansi")
	flag.Parse()

	t, err := theme.Named(*themeName)
	if err != nil {
		fmt.Fprintln(os.Stderr, "systui:", err)
		os.Exit(1)
	}
	m, err := ui.New(t)
	if err != nil {
		fmt.Fprintln(os.Stderr, "systui:", err)
		os.Exit(1)
	}
	if _, err := tea.NewProgram(m).Run(); err != nil {
		fmt.Fprintln(os.Stderr, "systui:", err)
		os.Exit(1)
	}
}
