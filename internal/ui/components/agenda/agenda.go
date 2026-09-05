// Package agenda renders the event list for a single day.
package agenda

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/stuckinsnow/gcalendar-cli/internal/calendar"
	"github.com/stuckinsnow/gcalendar-cli/internal/ui/theme"
)

// New creates an agenda pane.
func New(t theme.Theme, timeLayout string) Model {
	return Model{
		theme:      t,
		timeLayout: timeLayout,
		day:        calendar.StartOfDay(time.Now()),
		viewport:   viewport.New(40, 10),
	}
}

// SetDay replaces the displayed day and its events, resetting the selection.
func (m *Model) SetDay(day time.Time, events []calendar.Event) {
	sameDay := calendar.SameDay(m.day, day)
	m.day = calendar.StartOfDay(day)
	m.events = events

	if !sameDay {
		m.cursor = 0
		m.viewport.GotoTop()
	}
	m.clampCursor()
	m.refresh()
}

// SetSize resizes the scrollable area.
func (m *Model) SetSize(width, height int) {
	m.viewport.Width = width
	m.viewport.Height = height
	m.refresh()
}

// Focus marks the pane as holding keyboard focus.
func (m *Model) Focus(focused bool) {
	m.focused = focused
	m.refresh()
}

// MoveCursor changes the selected event by delta, clamped to the list.
func (m *Model) MoveCursor(delta int) {
	m.cursor += delta
	m.clampCursor()
	m.ensureVisible()
	m.refresh()
}

// Selected returns the highlighted event, if any.
func (m Model) Selected() (calendar.Event, bool) {
	if m.cursor < 0 || m.cursor >= len(m.events) {
		return calendar.Event{}, false
	}
	return m.events[m.cursor], true
}

// Update forwards scrolling messages to the viewport.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

// View renders the date heading above the scrollable list.
func (m Model) View() string {
	return lipgloss.JoinVertical(lipgloss.Left, m.heading(), m.viewport.View())
}

// heading is the date line plus a count of events.
func (m Model) heading() string {
	date := m.theme.AgendaDate.Render(m.day.Format("Monday 2 January"))
	if len(m.events) == 0 {
		return date
	}
	count := m.theme.EventMeta.Render(fmt.Sprintf("  %d event%s", len(m.events), plural(len(m.events))))
	return date + count
}

// refresh re-renders the viewport contents.
func (m *Model) refresh() { m.viewport.SetContent(m.content()) }

// content builds the list body.
func (m Model) content() string {
	if len(m.events) == 0 {
		return m.theme.AgendaEmpty.Render("Nothing scheduled.")
	}

	now := time.Now()
	blocks := make([]string, 0, len(m.events))
	for i, e := range m.events {
		blocks = append(blocks, m.entry(e, i == m.cursor, now))
	}
	return strings.Join(blocks, "\n")
}

// entry renders a single event as a two-line block with a coloured bar.
func (m Model) entry(e calendar.Event, selected bool, now time.Time) string {
	bar := lipgloss.NewStyle().Foreground(m.theme.EventColor(e.Color)).Render("▌")

	timeStyle, titleStyle := m.theme.EventTime, m.theme.EventTitle
	switch {
	case e.IsNow(now):
		timeStyle = m.theme.EventNow
		titleStyle = titleStyle.Bold(true)
	case e.IsPast(now):
		timeStyle, titleStyle = m.theme.EventPast, m.theme.EventPast
	}
	if selected && m.focused {
		titleStyle = titleStyle.Bold(true).Foreground(m.theme.Palette.Secondary)
	}

	marker := "  "
	if selected {
		marker = m.theme.Key.Render("▸ ")
	}

	width := m.viewport.Width
	title := truncate(e.Title, width-len(e.TimeRange(m.timeLayout))-6)
	first := marker + bar + " " +
		timeStyle.Render(e.TimeRange(m.timeLayout)) + "  " +
		titleStyle.Render(title)

	meta := metaLine(e)
	if meta == "" {
		return first
	}
	return first + "\n" + "    " + m.theme.EventMeta.Render(truncate(meta, width-4))
}

// metaLine joins the secondary details shown under an event title.
func metaLine(e calendar.Event) string {
	var parts []string
	if e.Location != "" {
		parts = append(parts, e.Location)
	}
	if e.Calendar != "" {
		parts = append(parts, e.Calendar)
	}
	return strings.Join(parts, " · ")
}

// clampCursor keeps the selection inside the list bounds.
func (m *Model) clampCursor() {
	if len(m.events) == 0 {
		m.cursor = 0
		return
	}
	if m.cursor < 0 {
		m.cursor = 0
	}
	if m.cursor >= len(m.events) {
		m.cursor = len(m.events) - 1
	}
}

// ensureVisible scrolls the viewport so the selection stays on screen. Each
// entry occupies up to two lines.
func (m *Model) ensureVisible() {
	line := m.cursor * 2
	switch {
	case line < m.viewport.YOffset:
		m.viewport.SetYOffset(line)
	case line >= m.viewport.YOffset+m.viewport.Height:
		m.viewport.SetYOffset(line - m.viewport.Height + 2)
	}
}

// truncate shortens text to width, appending an ellipsis when cut.
func truncate(s string, width int) string {
	if width < 4 {
		width = 4
	}
	runes := []rune(s)
	if len(runes) <= width {
		return s
	}
	return string(runes[:width-1]) + "…"
}

func plural(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}
