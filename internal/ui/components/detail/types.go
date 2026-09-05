package detail

import (
	"github.com/stuckinsnow/gcalendar-cli/internal/calendar"
	"github.com/stuckinsnow/gcalendar-cli/internal/ui/theme"
)

// Model holds the overlay state: which event is shown and the area to centre
// the card within.
type Model struct {
	theme      theme.Theme
	timeLayout string

	event   calendar.Event
	visible bool
	width   int
	height  int
}
