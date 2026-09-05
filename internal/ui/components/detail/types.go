package detail

import (
	"github.com/stuckinsnow/gcalendar-cli/internal/calendar"
	"github.com/stuckinsnow/gcalendar-cli/internal/ui/theme"
)

// Model holds the overlay state: which event is shown, the width to use as a
// side pane, and the screen area to centre the overlay fallback within.
type Model struct {
	theme      theme.Theme
	timeLayout string

	event     calendar.Event
	visible   bool
	paneWidth int
	width     int
	height    int
}
