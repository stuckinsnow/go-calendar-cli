// Package detail renders a single event as a centred overlay card.
package detail

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"

	"github.com/stuckinsnow/gcalendar-cli/internal/calendar"
	"github.com/stuckinsnow/gcalendar-cli/internal/ui/theme"
)

// New creates a hidden detail overlay.
func New(t theme.Theme, timeLayout string) Model {
	return Model{theme: t, timeLayout: timeLayout}
}

// Show displays an event.
func (m *Model) Show(e calendar.Event) {
	m.event = e
	m.visible = true
}

// Hide dismisses the overlay.
func (m *Model) Hide() { m.visible = false }

// Visible reports whether the overlay is on screen.
func (m Model) Visible() bool { return m.visible }

// SetSize records the area the overlay is centred within.
func (m *Model) SetSize(width, height int) {
	m.width, m.height = width, height
}

// Overlay renders the card centred over the given background.
func (m Model) Overlay(background string) string {
	if !m.visible {
		return background
	}
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, m.card())
}

// card renders the event details box.
func (m Model) card() string {
	e := m.event
	boxWidth := clamp(m.width-20, 34, 68)

	accent := lipgloss.NewStyle().Foreground(m.theme.EventColor(e.Color)).Bold(true)
	lines := []string{
		accent.Render(wrap(e.Title, boxWidth)),
		"",
		m.field("When", m.when()),
	}
	if e.Location != "" {
		lines = append(lines, m.field("Where", wrap(e.Location, boxWidth-8)))
	}
	if e.Calendar != "" {
		lines = append(lines, m.field("Calendar", e.Calendar))
	}
	if e.Declined {
		lines = append(lines, m.field("Status", "declined"))
	}
	if e.Description != "" {
		lines = append(lines, "", m.theme.DetailBodyText.Render(wrap(e.Description, boxWidth)))
	}
	lines = append(lines, "", m.theme.EventMeta.Render("esc / enter to close"))

	return m.theme.DetailBox.Width(boxWidth).Render(strings.Join(lines, "\n"))
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

	sameDay := calendar.SameDay(e.Start, e.End)
	if sameDay {
		return fmt.Sprintf("%s · %s – %s (%s)",
			e.Start.Format("Monday 2 January"),
			e.Start.Format(m.timeLayout), e.End.Format(m.timeLayout),
			humanDuration(e.Duration()))
	}
	return fmt.Sprintf("%s %s – %s %s",
		e.Start.Format("Mon 2 Jan"), e.Start.Format(m.timeLayout),
		e.End.Format("Mon 2 Jan"), e.End.Format(m.timeLayout))
}

// field renders a label/value pair.
func (m Model) field(label, value string) string {
	return m.theme.DetailLabel.Render(label+": ") + m.theme.DetailBodyText.Render(value)
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
