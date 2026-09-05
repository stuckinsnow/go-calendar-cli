package ui

import (
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/spinner"

	"github.com/stuckinsnow/gcalendar-cli/internal/calendar"
	"github.com/stuckinsnow/gcalendar-cli/internal/config"
	"github.com/stuckinsnow/gcalendar-cli/internal/ui/components/agenda"
	"github.com/stuckinsnow/gcalendar-cli/internal/ui/components/detail"
	"github.com/stuckinsnow/gcalendar-cli/internal/ui/components/header"
	"github.com/stuckinsnow/gcalendar-cli/internal/ui/components/legend"
	"github.com/stuckinsnow/gcalendar-cli/internal/ui/components/month"
	"github.com/stuckinsnow/gcalendar-cli/internal/ui/components/statusbar"
	"github.com/stuckinsnow/gcalendar-cli/internal/ui/theme"
)

// focusArea identifies which pane owns the keyboard.
type focusArea int

const (
	focusMonth focusArea = iota
	focusAgenda
)

// pane is the geometry handed to a lipgloss panel style. Width and Height are
// what Style.Width/Height expect: the content box including padding but
// excluding the border.
type pane struct {
	Width  int
	Height int
}

// InnerWidth is the width usable by a component inside the panel.
func (p pane) InnerWidth() int { return p.Width - 2*paddingWidth }

// InnerHeight is the height usable by a component inside the panel.
func (p pane) InnerHeight() int { return p.Height }

// OuterWidth includes the border, i.e. the columns the panel occupies.
func (p pane) OuterWidth() int { return p.Width + 2*borderWidth }

// OuterHeight includes the border, i.e. the rows the panel occupies.
func (p pane) OuterHeight() int { return p.Height + 2*borderWidth }

// layout holds the computed geometry for one render pass.
type layout struct {
	// stacked places the agenda below the grid instead of beside it.
	stacked bool
	// tooSmall reports that the terminal cannot host the UI.
	tooSmall bool
	// showLegend fills leftover space under the grid with the calendar legend.
	showLegend bool
	// showDetail gives the event details their own column on the right.
	showDetail bool

	month  pane
	agenda pane
	legend pane
	detail pane
}

// Model is the root Bubble Tea model. It owns the selected day and the event
// cache; the child components are pure views over that state.
type Model struct {
	// Dependencies
	cfg      config.Config
	provider calendar.Provider
	store    *calendar.Store
	theme    theme.Theme

	// Chrome and components
	keys    KeyMap
	help    help.Model
	spinner spinner.Model
	header  header.Model
	month   month.Model
	legend  legend.Model
	agenda  agenda.Model
	detail  detail.Model
	status  statusbar.Model

	// State
	cursor   time.Time // selected day, always midnight
	focus    focusArea
	inFlight int // outstanding provider fetches
	layout   layout
	width    int
	height   int
	ready    bool
}
