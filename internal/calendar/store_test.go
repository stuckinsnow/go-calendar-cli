package calendar

import (
	"testing"
	"time"
)

func TestStoreReplaceAndDay(t *testing.T) {
	s := NewStore()
	from := date(2026, time.September, 1)
	to := from.AddDate(0, 1, 0)

	s.Replace(from, to, []Event{
		{ID: "1", Title: "One", Start: from.Add(9 * time.Hour), End: from.Add(10 * time.Hour), Calendar: "Work", Color: "#7D56F4"},
	})

	if got := len(s.Day(from)); got != 1 {
		t.Fatalf("day has %d events, want 1", got)
	}
	if !s.HasMonth(from) {
		t.Error("month should be marked loaded")
	}
	if s.Calendars()["Work"] != "#7D56F4" {
		t.Errorf("calendar colours = %v", s.Calendars())
	}
}

func TestStoreReplaceDropsStaleEvents(t *testing.T) {
	s := NewStore()
	from := date(2026, time.September, 1)
	to := from.AddDate(0, 1, 0)

	s.Replace(from, to, []Event{{ID: "stale", Start: from.Add(9 * time.Hour), End: from.Add(10 * time.Hour)}})
	s.Replace(from, to, []Event{{ID: "fresh", Start: from.Add(11 * time.Hour), End: from.Add(12 * time.Hour)}})

	events := s.Day(from)
	if len(events) != 1 || events[0].ID != "fresh" {
		t.Fatalf("day events = %v, want only the fresh one", events)
	}
}

func TestStoreRangeDeduplicatesMultiDayEvents(t *testing.T) {
	s := NewStore()
	from := date(2026, time.September, 1)
	to := from.AddDate(0, 1, 0)

	s.Replace(from, to, []Event{{
		ID:     "offsite",
		AllDay: true,
		Start:  date(2026, time.September, 14),
		End:    date(2026, time.September, 17),
	}})

	if got := len(s.Range(from, to)); got != 1 {
		t.Errorf("range returned %d events, want 1 de-duplicated", got)
	}
}

func TestStoreInvalidate(t *testing.T) {
	s := NewStore()
	from := date(2026, time.September, 1)
	s.Replace(from, from.AddDate(0, 1, 0), []Event{{ID: "1", Start: from, End: from.Add(time.Hour)}})

	s.Invalidate()
	if s.HasMonth(from) || len(s.Day(from)) != 0 {
		t.Error("store should be empty after Invalidate")
	}
}
