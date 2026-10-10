package theme

import (
	"fmt"
	"image/color"

	"charm.land/lipgloss/v2"
)

// Palette holds the six neutral layers and six semantic colors of one theme.
// A nil color leaves the terminal's own color in place
type Palette struct {
	Background color.Color
	Surface    color.Color
	Border     color.Color
	Muted      color.Color
	Primary    color.Color
	Bright     color.Color
	OK         color.Color
	Warning    color.Color
	Critical   color.Color
	Accent     color.Color
	Dev        color.Color
	System     color.Color
}

// Glyphs pairs every semantic color with a symbol so that color is never the only signal
type Glyphs struct {
	OK         string
	Warning    string
	Critical   string
	Accent     string
	Dev        string
	System     string
	Dot        string
	BarLeft    string
	BarRight   string
	BarEighths []string
	BarTrack   string
	Expanded   string
	Collapsed  string
	More       string
	Above      string
	Below      string
	Enter      string
	Paused     string
	Stopped    string
	Zombie     string
	Foreign    string
	Ellipsis   string
}

type Theme struct {
	Colors   Palette
	Glyphs   Glyphs
	Border   lipgloss.Style
	Muted    lipgloss.Style
	Primary  lipgloss.Style
	Bright   lipgloss.Style
	OK       lipgloss.Style
	Warning  lipgloss.Style
	Critical lipgloss.Style
	Accent   lipgloss.Style
	Dev      lipgloss.Style
	System   lipgloss.Style
}

var glyphs = Glyphs{
	OK:         "●",
	Warning:    "⚠",
	Critical:   "✖",
	Accent:     "▌",
	Dev:        "◆",
	System:     "◇",
	Dot:        "·",
	BarLeft:    "▕",
	BarRight:   "▏",
	BarEighths: []string{"▏", "▎", "▍", "▌", "▋", "▊", "▉", "█"},
	BarTrack:   "░",
	Expanded:   "▼",
	Collapsed:  "▶",
	More:       "…",
	Above:      "▲",
	Below:      "▼",
	Enter:      "⏎",
	Paused:     "⏸",
	Stopped:    "‖",
	Zombie:     "†",
	Foreign:    "◌",
	Ellipsis:   "…",
}

var dark = Palette{
	Background: lipgloss.Color("#1a1d29"),
	Surface:    lipgloss.Color("#232736"),
	Border:     lipgloss.Color("#3a3f52"),
	Muted:      lipgloss.Color("#7c8299"),
	Primary:    lipgloss.Color("#d4d7e1"),
	Bright:     lipgloss.Color("#ffffff"),
	OK:         lipgloss.Color("#7ec699"),
	Warning:    lipgloss.Color("#e5c07b"),
	Critical:   lipgloss.Color("#e06c75"),
	Accent:     lipgloss.Color("#61afef"),
	Dev:        lipgloss.Color("#c678dd"),
	System:     lipgloss.Color("#56b6c2"),
}

var light = Palette{
	Background: lipgloss.Color("#fafafa"),
	Surface:    lipgloss.Color("#f0f1f4"),
	Border:     lipgloss.Color("#d0d4dc"),
	Muted:      lipgloss.Color("#6e7781"),
	Primary:    lipgloss.Color("#24292f"),
	Bright:     lipgloss.Color("#000000"),
	OK:         lipgloss.Color("#2e7d32"),
	Warning:    lipgloss.Color("#b08800"),
	Critical:   lipgloss.Color("#c62828"),
	Accent:     lipgloss.Color("#1565c0"),
	Dev:        lipgloss.Color("#7b1fa2"),
	System:     lipgloss.Color("#00838f"),
}

// ansi uses the terminal's own 16 colors, so systui inherits the user's theme
var ansi = Palette{
	Border:   lipgloss.Color("8"),
	Muted:    lipgloss.Color("8"),
	Bright:   lipgloss.Color("15"),
	OK:       lipgloss.Color("2"),
	Warning:  lipgloss.Color("3"),
	Critical: lipgloss.Color("1"),
	Accent:   lipgloss.Color("4"),
	Dev:      lipgloss.Color("5"),
	System:   lipgloss.Color("6"),
}

// Named returns the theme selected on the command line
func Named(name string) (Theme, error) {
	switch name {
	case "dark":
		return New(dark), nil
	case "light":
		return New(light), nil
	case "ansi":
		return New(ansi), nil
	}
	return Theme{}, fmt.Errorf("unknown theme %q, use dark, light or ansi", name)
}

func New(p Palette) Theme {
	return Theme{
		Colors:   p,
		Glyphs:   glyphs,
		Border:   foreground(p.Border),
		Muted:    foreground(p.Muted),
		Primary:  foreground(p.Primary),
		Bright:   foreground(p.Bright).Bold(true),
		OK:       foreground(p.OK),
		Warning:  foreground(p.Warning),
		Critical: foreground(p.Critical),
		Accent:   foreground(p.Accent),
		Dev:      foreground(p.Dev),
		System:   foreground(p.System),
	}
}

func foreground(c color.Color) lipgloss.Style {
	if c == nil {
		return lipgloss.NewStyle()
	}
	return lipgloss.NewStyle().Foreground(c)
}
