package legend

import (
	"github.com/stuckinsnow/gcalendar-cli/internal/ui/theme"
)

// Model renders the calendar legend.
type Model struct {
	theme theme.Theme

	entries []entry
	width   int
	height  int
}

// entry is one calendar and its colour.
type entry struct {
	name  string
	color string
}
