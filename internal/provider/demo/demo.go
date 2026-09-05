// Package demo provides an offline calendar provider so the UI can be explored
// without Google credentials.
package demo

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/stuckinsnow/gcalendar-cli/internal/calendar"
)

// Provider synthesises a plausible schedule around today's date.
type Provider struct {
	// Seed makes the generated schedule deterministic. Zero seeds from today.
	Seed int64
}

// New returns a demo provider seeded from the current day, so the schedule is
// stable for the whole session but varies between days.
func New() *Provider { return &Provider{} }

// Name identifies the provider in the status bar.
func (*Provider) Name() string { return "demo calendar" }

// Events builds the demo schedule for the requested window.
func (p *Provider) Events(_ context.Context, from, to time.Time) ([]calendar.Event, error) {
	today := calendar.StartOfDay(time.Now())

	seed := p.Seed
	if seed == 0 {
		seed = today.Unix()
	}
	rng := rand.New(rand.NewSource(seed))

	events := recurringEvents(from, to, rng)
	events = append(events, oneOffEvents(today, from, to)...)
	calendar.Sort(events)
	return events, nil
}

// recurringEvents expands the weekly templates across the window, skipping the
// occasional occurrence so the grid does not look mechanical. Today is never
// skipped, so the demo always opens on a populated day.
func recurringEvents(from, to time.Time, rng *rand.Rand) []calendar.Event {
	today := calendar.StartOfDay(time.Now())

	var out []calendar.Event
	for day := calendar.StartOfDay(from); day.Before(to); day = day.AddDate(0, 0, 1) {
		for i, t := range weeklyTemplates {
			if !t.occursOn(day.Weekday()) {
				continue
			}
			if rng.Intn(12) == 0 && !day.Equal(today) {
				continue
			}
			start := day.Add(t.at())
			out = append(out, calendar.Event{
				ID:       fmt.Sprintf("demo-weekly-%s-%d", day.Format("20060102"), i),
				Title:    t.Title,
				Start:    start,
				End:      start.Add(time.Duration(t.Minutes) * time.Minute),
				Location: t.Location,
				Calendar: t.Calendar,
				Color:    t.Color,
			})
		}
	}
	return out
}

// oneOffEvents places the fixed fixtures relative to today.
func oneOffEvents(today, from, to time.Time) []calendar.Event {
	var out []calendar.Event
	for i, f := range oneOffs {
		day := today.AddDate(0, 0, f.DayOffset)
		if day.Before(calendar.StartOfDay(from)) || !day.Before(to) {
			continue
		}

		e := calendar.Event{
			ID:          fmt.Sprintf("demo-oneoff-%d", i),
			Title:       f.Title,
			Location:    f.Location,
			Calendar:    f.Calendar,
			Color:       f.Color,
			Description: f.Description,
		}
		if f.AllDay {
			e.AllDay = true
			e.Start = day
			e.End = day.AddDate(0, 0, 1)
		} else {
			e.Start = day.Add(time.Duration(f.Hour)*time.Hour + time.Duration(f.Minute)*time.Minute)
			e.End = e.Start.Add(time.Duration(f.Minutes) * time.Minute)
		}
		out = append(out, e)
	}
	return out
}
