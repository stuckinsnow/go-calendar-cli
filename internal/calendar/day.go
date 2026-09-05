package calendar

import (
	"sort"
	"time"
)

// StartOfDay truncates a timestamp to midnight in its own location.
func StartOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

// StartOfMonth returns midnight on the first day of t's month.
func StartOfMonth(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
}

// SameDay reports whether two timestamps fall on the same calendar day.
func SameDay(a, b time.Time) bool { return StartOfDay(a).Equal(StartOfDay(b)) }

// Sort orders events chronologically, with all-day entries hoisted to the top
// of the day they belong to.
func Sort(events []Event) {
	sort.SliceStable(events, func(i, j int) bool {
		a, b := events[i], events[j]
		if a.AllDay != b.AllDay && SameDay(a.Start, b.Start) {
			return a.AllDay
		}
		if !a.Start.Equal(b.Start) {
			return a.Start.Before(b.Start)
		}
		return a.Title < b.Title
	})
}

// GroupByDay buckets events under the midnight timestamp of every day they
// touch, so multi-day events appear on each of them.
func GroupByDay(events []Event) map[time.Time][]Event {
	out := make(map[time.Time][]Event)
	for _, e := range events {
		for _, d := range e.Days() {
			out[d] = append(out[d], e)
		}
	}
	for day := range out {
		Sort(out[day])
	}
	return out
}

// Week is one row of a month grid: seven consecutive days.
type Week [7]time.Time

// GridRows is the number of week rows in a month grid. Six rows cover every
// possible month layout, which keeps the view a fixed height.
const GridRows = 6

// Grid lays out the month containing t as six weeks of seven days, padded with
// the trailing days of the previous month and the leading days of the next.
func Grid(t time.Time, weekStart time.Weekday) []Week {
	first := StartOfMonth(t)
	offset := (int(first.Weekday()) - int(weekStart) + 7) % 7
	cursor := first.AddDate(0, 0, -offset)

	weeks := make([]Week, GridRows)
	for r := range weeks {
		for c := 0; c < 7; c++ {
			weeks[r][c] = cursor
			cursor = cursor.AddDate(0, 0, 1)
		}
	}
	return weeks
}

// WeekdayNames returns short weekday labels ordered from weekStart.
func WeekdayNames(weekStart time.Weekday) [7]string {
	var names [7]string
	for i := 0; i < 7; i++ {
		d := time.Weekday((int(weekStart) + i) % 7)
		names[i] = d.String()[:3]
	}
	return names
}

// IsWeekend reports whether a day falls on Saturday or Sunday.
func IsWeekend(t time.Time) bool {
	return t.Weekday() == time.Saturday || t.Weekday() == time.Sunday
}
