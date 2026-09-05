package agenda

import (
	"time"

	"github.com/charmbracelet/bubbles/viewport"

	"github.com/stuckinsnow/gcalendar-cli/internal/calendar"
	"github.com/stuckinsnow/gcalendar-cli/internal/ui/theme"
)

// Model is the day agenda: a scrollable, selectable list of events.
type Model struct {
	theme      theme.Theme
	timeLayout string

	day      time.Time
	events   []calendar.Event
	cursor   int
	focused  bool
	viewport viewport.Model
}
