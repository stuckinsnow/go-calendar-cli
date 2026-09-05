// Package detail renders a single event, either as a side pane or, when the
// terminal is too narrow for a third column, as a centred overlay card.
package detail

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"

	"github.com/stuckinsnow/gcalendar-cli/internal/calendar"
	"github.com/stuckinsnow/gcalendar-cli/internal/ui/theme"
)

// New creates a hidden detail view.
func New(t theme.Theme, timeLayout string) Model {
	return Model{theme: t, timeLayout: timeLayout}
}

// Show displays an event.
func (m *Model) Show(e calendar.Event) {
	m.event = e
	m.visible = true
}

// Hide dismisses the view.
func (m *Model) Hide() { m.visible = false }

// Visible reports whether the view is on screen.
func (m Model) Visible() bool { return m.visible }

// Event returns the event being shown.
func (m Model) Event() calendar.Event { return m.event }

// SetSize records the screen area, used to centre the overlay.
func (m *Model) SetSize(width, height int) {
	m.width, m.height = width, height
}

// SetPaneWidth sets the content width when rendering as a side pane.
func (m *Model) SetPaneWidth(width int) { m.paneWidth = width }

// Pane renders the details for the side column, without a border: the caller's
// panel style supplies that.
func (m Model) Pane() string {
	if !m.visible {
		return m.theme.AgendaEmpty.Render("Select an event to see its details.")
	}
	return m.body(m.paneWidth, "esc to close")
}

// Overlay renders the card centred over the given background.
func (m Model) Overlay(background string) string {
	if !m.visible {
		return background
	}
	width := clamp(m.width-20, 34, 68)
	card := m.theme.DetailBox.Width(width).Render(m.body(width, "esc / enter to close"))

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, card)
}

// body renders the event fields wrapped to width, ending with a dismiss hint.
func (m Model) body(width int, hint string) string {
	e := m.event

	accent := lipgloss.NewStyle().Foreground(m.theme.EventColor(e.Color)).Bold(true)
	lines := []string{
		accent.Render(wrap(e.Title, width)),
		"",
		m.field("When", m.when(), width),
	}
	if e.Location != "" {
		lines = append(lines, m.field("Where", e.Location, width))
	}
	if e.Calendar != "" {
		lines = append(lines, m.field("Calendar", e.Calendar, width))
	}
	if e.Declined {
		lines = append(lines, m.field("Status", "declined", width))
	}
	if e.Description != "" {
		lines = append(lines, "", m.theme.DetailBodyText.Render(wrap(e.Description, width)))
	}
	if links := e.Links(); len(links) > 0 {
		lines = append(lines, "", m.theme.DetailLabel.Render("Links"))
		for i, link := range links {
			number := m.theme.Key.Render(fmt.Sprintf("%d", i+1))
			lines = append(lines, number+" "+m.theme.DetailBodyText.Render(wrap(link, width-2)))
		}
		lines = append(lines, m.theme.EventMeta.Render("press 1–"+
			fmt.Sprintf("%d", len(links))+" to copy a link"))
	}
	if hint != "" {
		lines = append(lines, "", m.theme.EventMeta.Render(hint))
	}
	return strings.Join(lines, "\n")
}

// when describes the event's timing in prose.
func (m Model) when() string {
	e := m.event
	if e.AllDay {
		days := len(e.Days())
		if days <= 1 {
			return e.Start.Format("Monday 2 January 2006") + " · all day"
		}
		return fmt.Sprintf("%s – %s · %d days",
			e.Start.Format("Mon 2 Jan"),
			e.End.AddDate(0, 0, -1).Format("Mon 2 Jan"), days)
	}

	if calendar.SameDay(e.Start, e.End) {
		return fmt.Sprintf("%s · %s – %s (%s)",
			e.Start.Format("Monday 2 January"),
			e.Start.Format(m.timeLayout), e.End.Format(m.timeLayout),
			humanDuration(e.Duration()))
	}
	return fmt.Sprintf("%s %s – %s %s",
		e.Start.Format("Mon 2 Jan"), e.Start.Format(m.timeLayout),
		e.End.Format("Mon 2 Jan"), e.End.Format(m.timeLayout))
}

// field renders a label/value pair, wrapping the value under the label when it
// does not fit on one line.
func (m Model) field(label, value string, width int) string {
	head := m.theme.DetailLabel.Render(label + ": ")
	if lipgloss.Width(head)+lipgloss.Width(value) <= width && !strings.Contains(value, "\n") {
		return head + m.theme.DetailBodyText.Render(value)
	}
	return head + "\n" + m.theme.DetailBodyText.Render(wrap(value, width))
}

// humanDuration renders a duration as "1h 30m".
func humanDuration(d time.Duration) string {
	hours := int(d.Hours())
	mins := int(d.Minutes()) % 60
	switch {
	case hours == 0:
		return fmt.Sprintf("%dm", mins)
	case mins == 0:
		return fmt.Sprintf("%dh", hours)
	default:
		return fmt.Sprintf("%dh %dm", hours, mins)
	}
}

func wrap(s string, width int) string {
	if width < 10 {
		width = 10
	}
	return lipgloss.NewStyle().Width(width).Render(s)
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
