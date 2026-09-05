package calendar

import (
	"testing"
	"time"
)

func date(y int, m time.Month, d int) time.Time {
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

func TestGridCoversMonthWithPadding(t *testing.T) {
	// September 2026 starts on a Tuesday.
	weeks := Grid(date(2026, time.September, 5), time.Monday)

	if len(weeks) != GridRows {
		t.Fatalf("want %d week rows, got %d", GridRows, len(weeks))
	}
	if got := weeks[0][0]; !got.Equal(date(2026, time.August, 31)) {
		t.Errorf("first cell = %s, want 2026-08-31 (Monday)", got.Format("2006-01-02"))
	}
	if got := weeks[0][0].Weekday(); got != time.Monday {
		t.Errorf("first cell weekday = %s, want Monday", got)
	}

	// Cells must be strictly consecutive days.
	prev := weeks[0][0]
	for r := 0; r < len(weeks); r++ {
		for c := 0; c < 7; c++ {
			if r == 0 && c == 0 {
				continue
			}
			want := prev.AddDate(0, 0, 1)
			if !weeks[r][c].Equal(want) {
				t.Fatalf("cell [%d][%d] = %s, want %s", r, c,
					weeks[r][c].Format("2006-01-02"), want.Format("2006-01-02"))
			}
			prev = weeks[r][c]
		}
	}
}

func TestGridRespectsSundayStart(t *testing.T) {
	weeks := Grid(date(2026, time.September, 5), time.Sunday)
	if got := weeks[0][0].Weekday(); got != time.Sunday {
		t.Errorf("first cell weekday = %s, want Sunday", got)
	}
}

func TestWeekdayNamesOrder(t *testing.T) {
	names := WeekdayNames(time.Monday)
	if names[0] != "Mon" || names[6] != "Sun" {
		t.Errorf("names = %v, want Mon..Sun", names)
	}
}

func TestEventDaysTimed(t *testing.T) {
	start := time.Date(2026, time.September, 5, 9, 0, 0, 0, time.UTC)
	e := Event{Start: start, End: start.Add(90 * time.Minute)}

	days := e.Days()
	if len(days) != 1 || !days[0].Equal(date(2026, time.September, 5)) {
		t.Fatalf("days = %v, want [2026-09-05]", days)
	}
}

func TestEventDaysEndingAtMidnightStaysOnStartDay(t *testing.T) {
	start := time.Date(2026, time.September, 5, 22, 0, 0, 0, time.UTC)
	e := Event{Start: start, End: date(2026, time.September, 6)}

	if days := e.Days(); len(days) != 1 {
		t.Fatalf("days = %v, want a single day", days)
	}
}

func TestEventDaysAllDayUsesExclusiveEnd(t *testing.T) {
	e := Event{
		AllDay: true,
		Start:  date(2026, time.September, 14),
		End:    date(2026, time.September, 17), // exclusive
	}

	days := e.Days()
	if len(days) != 3 {
		t.Fatalf("days = %v, want 3 days (14th–16th)", days)
	}
	if !days[2].Equal(date(2026, time.September, 16)) {
		t.Errorf("last day = %s, want 2026-09-16", days[2].Format("2006-01-02"))
	}
}

func TestGroupByDaySpreadsMultiDayEvents(t *testing.T) {
	events := []Event{
		{ID: "a", AllDay: true, Start: date(2026, time.September, 14), End: date(2026, time.September, 16)},
		{ID: "b", Start: time.Date(2026, time.September, 15, 9, 0, 0, 0, time.UTC),
			End: time.Date(2026, time.September, 15, 10, 0, 0, 0, time.UTC)},
	}

	grouped := GroupByDay(events)
	if got := len(grouped[date(2026, time.September, 14)]); got != 1 {
		t.Errorf("14th has %d events, want 1", got)
	}
	if got := len(grouped[date(2026, time.September, 15)]); got != 2 {
		t.Errorf("15th has %d events, want 2", got)
	}
	if got := len(grouped[date(2026, time.September, 16)]); got != 0 {
		t.Errorf("16th has %d events, want 0 (exclusive end)", got)
	}
}

func TestSortHoistsAllDayEvents(t *testing.T) {
	day := date(2026, time.September, 15)
	events := []Event{
		{ID: "timed", Start: day.Add(9 * time.Hour), End: day.Add(10 * time.Hour)},
		{ID: "allday", AllDay: true, Start: day, End: day.AddDate(0, 0, 1)},
	}

	Sort(events)
	if events[0].ID != "allday" {
		t.Errorf("first event = %s, want allday", events[0].ID)
	}
}

func TestEventStateHelpers(t *testing.T) {
	now := time.Date(2026, time.September, 5, 12, 0, 0, 0, time.UTC)
	current := Event{Start: now.Add(-10 * time.Minute), End: now.Add(20 * time.Minute)}
	past := Event{Start: now.Add(-2 * time.Hour), End: now.Add(-1 * time.Hour)}

	if !current.IsNow(now) {
		t.Error("current event should report IsNow")
	}
	if current.IsPast(now) {
		t.Error("current event should not report IsPast")
	}
	if !past.IsPast(now) {
		t.Error("finished event should report IsPast")
	}
}

func TestTimeRange(t *testing.T) {
	day := date(2026, time.September, 5)
	timed := Event{Start: day.Add(9*time.Hour + 30*time.Minute), End: day.Add(10 * time.Hour)}

	if got := timed.TimeRange("15:04"); got != "09:30 – 10:00" {
		t.Errorf("TimeRange = %q", got)
	}
	allDay := Event{AllDay: true, Start: day, End: day.AddDate(0, 0, 1)}
	if got := allDay.TimeRange("15:04"); got != "all day" {
		t.Errorf("all-day TimeRange = %q", got)
	}
}
