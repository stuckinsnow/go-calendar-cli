// Package month renders a month grid with per-day event indicators.
package month

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"

	"github.com/stuckinsnow/gcalendar-cli/internal/calendar"
	"github.com/stuckinsnow/gcalendar-cli/internal/ui/theme"
)

// Grid layout constants.
const (
	minCellWidth = 5
	maxCellWidth = 12
	columns      = 7
	gutter       = 1 // blank column on the right of every cell
)

// New creates a month grid.
func New(t theme.Theme, weekStart time.Weekday, src Source) Model {
	now := calendar.StartOfDay(time.Now())
	return Model{
		theme:     t,
		source:    src,
		weekStart: weekStart,
		cursor:    now,
		today:     now,
		width:     minCellWidth * columns,
	}
}

// SetCursor moves the selected day.
func (m *Model) SetCursor(day time.Time) { m.cursor = calendar.StartOfDay(day) }

// Cursor returns the selected day.
func (m Model) Cursor() time.Time { return m.cursor }

// SetToday updates the "today" marker, for sessions that cross midnight.
func (m *Model) SetToday(day time.Time) { m.today = calendar.StartOfDay(day) }

// SetWidth sets the drawable width, excluding panel border and padding.
func (m *Model) SetWidth(w int) { m.width = w }

// Focus marks the grid as holding keyboard focus.
func (m *Model) Focus(focused bool) { m.focused = focused }

// Width reports the exact rendered width, which is a multiple of the cell size.
func (m Model) Width() int { return m.cellWidth() * columns }

// Height reports the rendered height: the weekday row plus two lines per week.
func (m Model) Height() int { return 1 + calendar.GridRows*2 }

// cellWidth divides the available width between the seven columns.
func (m Model) cellWidth() int {
	return clamp(m.width/columns, minCellWidth, maxCellWidth)
}

// View renders the weekday header followed by six week rows.
func (m Model) View() string {
	cellW := m.cellWidth()

	rows := make([]string, 0, calendar.GridRows+1)
	rows = append(rows, m.headerRow(cellW))

	for _, week := range calendar.Grid(m.cursor, m.weekStart) {
		cells := make([]string, 0, columns)
		for _, day := range week {
			cells = append(cells, m.cell(day, cellW))
		}
		rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Top, cells...))
	}
	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}

// headerRow renders centred weekday abbreviations above the columns.
func (m Model) headerRow(cellW int) string {
	cells := make([]string, 0, columns)
	for _, name := range calendar.WeekdayNames(m.weekStart) {
		cells = append(cells, m.theme.Weekday.Width(cellW-gutter).Render(name)+" ")
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, cells...)
}

// cell renders one day: the day number above a row of event indicators.
//
// Both lines are exactly cellW columns wide, the last being a gutter, so
// neighbouring days never touch and a highlighted cell cannot shift its row.
func (m Model) cell(day time.Time, cellW int) string {
	inMonth := day.Month() == m.cursor.Month()

	return lipgloss.JoinVertical(lipgloss.Left,
		m.numberLine(day, cellW, inMonth),
		m.indicatorLine(cellW, m.source.Day(day)),
	)
}

// numberLine styles the day number by selection, today and month membership.
// The today marker sits right of the number so the number column never moves.
func (m Model) numberLine(day time.Time, cellW int, inMonth bool) string {
	field := cellW - gutter
	label := fmt.Sprintf("%*d", field-1, day.Day())

	marker := " "
	if day.Equal(m.today) {
		marker = "·"
	}

	style := m.theme.Day
	switch {
	case day.Equal(m.cursor) && m.focused:
		style = lipgloss.NewStyle().
			Foreground(m.theme.Palette.Inverted).
			Background(m.theme.Palette.Primary).
			Bold(true)
	case day.Equal(m.cursor):
		style = m.theme.DaySelected.Bold(true)
	case day.Equal(m.today):
		style = m.theme.DayToday
	case !inMonth:
		style = m.theme.DayOutside
	case calendar.IsWeekend(day):
		style = m.theme.DayWeekend
	}

	return style.Width(field).Render(label+marker) + strings.Repeat(" ", gutter)
}

// indicatorLine draws one coloured dot per event, right-aligned beneath the day
// number, with an overflow count when the day is busier than the cell is wide.
func (m Model) indicatorLine(cellW int, events []calendar.Event) string {
	if len(events) == 0 {
		return strings.Repeat(" ", cellW)
	}

	field := cellW - gutter
	capacity := max(field-1, 1)

	shown, overflow := events, 0
	if len(events) > capacity {
		shown = events[:capacity-1]
		overflow = len(events) - len(shown)
	}

	var b strings.Builder
	for _, e := range shown {
		dot := "•"
		if e.AllDay {
			dot = "▪"
		}
		b.WriteString(lipgloss.NewStyle().Foreground(m.theme.EventColor(e.Color)).Render(dot))
	}
	if overflow > 0 {
		b.WriteString(m.theme.EventMeta.Render(fmt.Sprintf("+%d", overflow)))
	}

	dots := lipgloss.NewStyle().Width(field).Align(lipgloss.Right).Render(b.String())
	return dots + strings.Repeat(" ", gutter)
}

func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
