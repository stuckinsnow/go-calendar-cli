package ui

import (
	"fmt"
	"strings"

	"github.com/stuckinsnow/gcalendar-cli/internal/calendar"
	"github.com/stuckinsnow/gcalendar-cli/internal/clipboard"
)

// copySelection puts the selected event on the clipboard: its Google Calendar
// link when the provider supplied one, otherwise a plain-text summary.
func (m *Model) copySelection() {
	event, ok := m.selectedEvent()
	if !ok {
		m.status.SetMessage("nothing selected to copy")
		return
	}

	payload, kind := copyPayload(event, m.cfg.TimeLayout())
	m.copyToClipboard(payload, kind)
}

// copyLink copies the nth link found inside the selected event, 1-based, as
// listed in the detail pane.
func (m *Model) copyLink(n int) {
	event, ok := m.selectedEvent()
	if !ok {
		return
	}

	links := event.Links()
	if n < 1 || n > len(links) {
		m.status.SetMessage(fmt.Sprintf("no link %d in this event", n))
		return
	}
	m.copyToClipboard(links[n-1], fmt.Sprintf("link %d", n))
}

// copyToClipboard performs the copy and reports the outcome in the status bar.
func (m *Model) copyToClipboard(payload, kind string) {
	via, err := clipboard.Copy(payload)
	if err != nil {
		m.status.SetError(fmt.Errorf("copy failed: %w", err))
		return
	}
	m.status.SetMessage(fmt.Sprintf("copied %s via %s", kind, via))
}

// selectedEvent is the event under the agenda cursor, or the first of the day
// when the grid has focus.
func (m Model) selectedEvent() (calendar.Event, bool) {
	if event, ok := m.agenda.Selected(); ok {
		return event, true
	}
	if events := m.store.Day(m.cursor); len(events) > 0 {
		return events[0], true
	}
	return calendar.Event{}, false
}

// copyPayload builds the clipboard text and a label for the status bar.
func copyPayload(e calendar.Event, timeLayout string) (payload, kind string) {
	if e.URL != "" {
		return e.URL, "link"
	}

	parts := []string{
		e.Title,
		e.Start.Format("Mon 2 Jan 2006") + " · " + e.TimeRange(timeLayout),
	}
	if e.Location != "" {
		parts = append(parts, e.Location)
	}
	if e.Calendar != "" {
		parts = append(parts, e.Calendar)
	}
	return strings.Join(parts, "\n"), "event details"
}
