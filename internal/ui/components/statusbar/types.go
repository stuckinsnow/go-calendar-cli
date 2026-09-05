package statusbar

import (
	"time"

	"github.com/stuckinsnow/gcalendar-cli/internal/ui/theme"
)

// Model is the status line: loading state, messages, errors and sync time.
type Model struct {
	theme theme.Theme
	width int

	message  string
	err      error
	loading  bool
	spinner  string
	syncedAt time.Time
}
