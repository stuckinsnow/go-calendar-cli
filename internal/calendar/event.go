// Package calendar holds the domain model shared by every provider and by the
// terminal UI. It deliberately knows nothing about Google's API or Bubble Tea.
package calendar

import "time"

// Event is a single calendar entry, normalised away from any API shape.
type Event struct {
	ID            string
	Title         string
	Start         time.Time
	End           time.Time
	AllDay        bool
	Location      string
	Description   string
	Calendar      string // display name of the owning calendar
	Color         string // hex colour such as "#7D56F4"; may be empty
	URL           string // link to the event in Google Calendar, when known
	ConferenceURL string // Meet/Zoom style joining link, when known
	Declined      bool
}

// Duration reports the event length, clamped to at least a minute so that
// zero-length entries still occupy space when rendered.
func (e Event) Duration() time.Duration {
	if d := e.End.Sub(e.Start); d > time.Minute {
		return d
	}
	return time.Minute
}

// TimeRange renders a compact label such as "09:30 – 10:15", or "all day".
// layout is any Go time layout, e.g. "15:04" or "3:04pm".
func (e Event) TimeRange(layout string) string {
	if e.AllDay {
		return "all day"
	}
	return e.Start.Format(layout) + " – " + e.End.Format(layout)
}

// Days returns the midnight timestamps of every day the event touches.
func (e Event) Days() []time.Time {
	first := StartOfDay(e.Start)
	last := StartOfDay(e.End)

	switch {
	case e.AllDay:
		// All-day events carry an exclusive end date.
		last = StartOfDay(e.End.AddDate(0, 0, -1))
	case e.End.Equal(last):
		// Ending exactly at midnight belongs to the previous day.
		last = last.AddDate(0, 0, -1)
	}
	if last.Before(first) {
		last = first
	}

	var days []time.Time
	for d := first; !d.After(last); d = d.AddDate(0, 0, 1) {
		days = append(days, d)
	}
	return days
}

// IsPast reports whether the event has already finished relative to now.
func (e Event) IsPast(now time.Time) bool { return e.End.Before(now) }

// IsNow reports whether the event is currently in progress.
func (e Event) IsNow(now time.Time) bool {
	return !e.AllDay && !e.Start.After(now) && e.End.After(now)
}
