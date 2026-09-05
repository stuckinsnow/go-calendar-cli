package ui

import (
	"context"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/stuckinsnow/gcalendar-cli/internal/calendar"
)

// fetchTimeout bounds a single provider fetch.
const fetchTimeout = 25 * time.Second

// eventsLoadedMsg carries a completed fetch.
type eventsLoadedMsg struct {
	from   time.Time
	to     time.Time
	events []calendar.Event
}

// fetchFailedMsg reports a failed fetch.
type fetchFailedMsg struct{ err error }

// clockTickMsg nudges the UI so "now" indicators stay accurate.
type clockTickMsg time.Time

// loadEvents fetches the window covering the month grid around anchor.
func loadEvents(provider calendar.Provider, from, to time.Time) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), fetchTimeout)
		defer cancel()

		events, err := provider.Events(ctx, from, to)
		if err != nil {
			return fetchFailedMsg{err: err}
		}
		return eventsLoadedMsg{from: from, to: to, events: events}
	}
}

// clockTick schedules the next clock refresh.
func clockTick() tea.Cmd {
	return tea.Tick(time.Minute, func(t time.Time) tea.Msg { return clockTickMsg(t) })
}

// gridWindow returns the half-open range covering every day visible in the
// month grid for anchor, so off-month cells show their events too.
func gridWindow(anchor time.Time, weekStart time.Weekday) (from, to time.Time) {
	weeks := calendar.Grid(anchor, weekStart)
	first := weeks[0][0]
	last := weeks[len(weeks)-1][6]
	return first, last.AddDate(0, 0, 1)
}
