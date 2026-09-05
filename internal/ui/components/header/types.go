package header

import (
	"time"

	"github.com/stuckinsnow/gcalendar-cli/internal/ui/theme"
)

// Model is the top-of-screen title bar.
type Model struct {
	theme theme.Theme
	width int

	month  time.Time
	source string
}
