// Package statusbar renders the bottom status line: loading state, messages
// and errors.
package statusbar

import (
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"

	"github.com/stuckinsnow/gcalendar-cli/internal/ui/theme"
)

// New creates a status bar.
func New(t theme.Theme) Model { return Model{theme: t, width: 80} }

// SetWidth sets the total width.
func (m *Model) SetWidth(w int) { m.width = w }

// SetMessage replaces the transient message and clears any error.
func (m *Model) SetMessage(msg string) {
	m.message, m.err = msg, nil
}

// SetError shows an error, which takes priority over the message.
func (m *Model) SetError(err error) { m.err = err }

// SetLoading toggles the loading indicator; spinner is the current frame.
func (m *Model) SetLoading(loading bool, spinner string) {
	m.loading, m.spinner = loading, spinner
}

// MarkSynced records a successful fetch.
func (m *Model) MarkSynced(at time.Time) {
	m.syncedAt = at
	m.loading = false
}

// Height is the number of lines the bar occupies.
func (Model) Height() int { return 1 }

// View renders the bar with the message on the left and sync state right.
func (m Model) View() string {
	left := m.leftSegment()
	right := m.rightSegment()

	gap := m.width - lipgloss.Width(left) - lipgloss.Width(right)
	if gap < 1 {
		gap = 1
	}
	return left + strings.Repeat(" ", gap) + right
}

func (m Model) leftSegment() string {
	switch {
	case m.err != nil:
		return m.theme.ErrorMsg.Render("⚠ " + m.err.Error())
	case m.loading:
		return m.theme.Status.Render(m.spinner + " loading events…")
	case m.message != "":
		return m.theme.Status.Render(m.message)
	default:
		return ""
	}
}

func (m Model) rightSegment() string {
	if m.syncedAt.IsZero() {
		return ""
	}
	return m.theme.Status.Render("synced " + m.syncedAt.Format("15:04:05"))
}
