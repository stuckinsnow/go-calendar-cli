package calendar

import (
	"sync"
	"time"
)

// Store is an in-memory, day-indexed cache of fetched events. It is safe for
// concurrent use so background fetches can populate it while the UI reads.
type Store struct {
	mu     sync.RWMutex
	byDay  map[time.Time][]Event
	spans  map[string]time.Time // month key -> time the month was last loaded
	colors map[string]string    // calendar name -> hex colour
}

// NewStore returns an empty store.
func NewStore() *Store {
	return &Store{
		byDay:  make(map[time.Time][]Event),
		spans:  make(map[string]time.Time),
		colors: make(map[string]string),
	}
}

// monthKey identifies a month for load tracking.
func monthKey(t time.Time) string { return t.Format("2006-01") }

// Replace swaps in the events for [from, to), dropping anything previously
// cached in that window so refreshes cannot leave stale entries behind.
func (s *Store) Replace(from, to time.Time, events []Event) {
	grouped := GroupByDay(events)

	s.mu.Lock()
	defer s.mu.Unlock()

	for d := StartOfDay(from); d.Before(to); d = d.AddDate(0, 0, 1) {
		delete(s.byDay, d)
	}
	for day, evs := range grouped {
		if day.Before(StartOfDay(from)) || !day.Before(to) {
			continue
		}
		s.byDay[day] = evs
	}
	for m := StartOfMonth(from); m.Before(to); m = m.AddDate(0, 1, 0) {
		s.spans[monthKey(m)] = time.Now()
	}
	for _, e := range events {
		if e.Calendar != "" && e.Color != "" {
			s.colors[e.Calendar] = e.Color
		}
	}
}

// Day returns the cached events for a single day.
func (s *Store) Day(day time.Time) []Event {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.byDay[StartOfDay(day)]
}

// Range returns every cached event overlapping [from, to), de-duplicated and
// sorted chronologically.
func (s *Store) Range(from, to time.Time) []Event {
	s.mu.RLock()
	defer s.mu.RUnlock()

	seen := make(map[string]struct{})
	var out []Event
	for d := StartOfDay(from); d.Before(to); d = d.AddDate(0, 0, 1) {
		for _, e := range s.byDay[d] {
			if _, dup := seen[e.ID]; dup {
				continue
			}
			seen[e.ID] = struct{}{}
			out = append(out, e)
		}
	}
	Sort(out)
	return out
}

// HasMonth reports whether the month containing t has been loaded.
func (s *Store) HasMonth(t time.Time) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, ok := s.spans[monthKey(t)]
	return ok
}

// Invalidate forgets every cached day and load marker.
func (s *Store) Invalidate() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.byDay = make(map[time.Time][]Event)
	s.spans = make(map[string]time.Time)
}

// Calendars returns the known calendar names mapped to their colour.
func (s *Store) Calendars() map[string]string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make(map[string]string, len(s.colors))
	for k, v := range s.colors {
		out[k] = v
	}
	return out
}
