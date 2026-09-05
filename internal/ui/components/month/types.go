package month

import (
	"time"

	"github.com/stuckinsnow/gcalendar-cli/internal/calendar"
	"github.com/stuckinsnow/gcalendar-cli/internal/ui/theme"
)

// Source supplies the events for a given day. *calendar.Store satisfies it.
type Source interface {
	Day(day time.Time) []calendar.Event
}

// Model draws the grid. The app owns the cursor; this component only renders.
type Model struct {
	theme     theme.Theme
	source    Source
	weekStart time.Weekday

	cursor  time.Time // selected day, always midnight
	today   time.Time
	width   int
	focused bool
}
