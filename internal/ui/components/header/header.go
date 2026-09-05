// Package header renders the title bar above the calendar panes.
package header

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"

	"github.com/stuckinsnow/gcalendar-cli/internal/ui/theme"
)

// New creates a header for the given source label (provider name).
func New(t theme.Theme, source string) Model {
	return Model{theme: t, source: source, month: time.Now(), width: 80}
}

// SetMonth updates the displayed month.
func (m *Model) SetMonth(month time.Time) { m.month = month }

// SetWidth sets the total width.
func (m *Model) SetWidth(w int) { m.width = w }

// Height is the number of lines the header occupies.
func (Model) Height() int { return 2 }

// View renders "September 2026" with the source on the right, over a rule.
func (m Model) View() string {
	title := m.theme.Title.Render(m.month.Format("January 2006"))
	week := m.theme.Subtitle.Render(fmt.Sprintf("  week %d", isoWeek(m.month)))
	left := title + week
	right := m.theme.Subtitle.Render(m.source)

	gap := m.width - lipgloss.Width(left) - lipgloss.Width(right)
	if gap < 1 {
		gap = 1
	}

	rule := lipgloss.NewStyle().
		Foreground(m.theme.Palette.Border).
		Render(strings.Repeat("─", max(m.width, 1)))

	return lipgloss.JoinVertical(lipgloss.Left, left+strings.Repeat(" ", gap)+right, rule)
}

func isoWeek(t time.Time) int {
	_, week := t.ISOWeek()
	return week
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
