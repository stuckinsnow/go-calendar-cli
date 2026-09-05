package ui

import (
	"context"
	"regexp"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/stuckinsnow/gcalendar-cli/internal/config"
	"github.com/stuckinsnow/gcalendar-cli/internal/provider/demo"
)

var ansi = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)

// plain strips ANSI styling so assertions can match on text.
func plain(s string) string { return ansi.ReplaceAllString(s, "") }

// loadedModel returns a model sized for a wide terminal with demo events
// already delivered, i.e. the state a user sees after startup.
func loadedModel(t *testing.T, width, height int) Model {
	t.Helper()

	provider := demo.New()
	m := New(config.Defaults(), provider)

	model, _ := m.Update(tea.WindowSizeMsg{Width: width, Height: height})
	m = model.(Model)

	from, to := gridWindow(m.cursor, m.cfg.WeekStart)
	events, err := provider.Events(context.Background(), from, to)
	if err != nil {
		t.Fatalf("demo provider: %v", err)
	}

	model, _ = m.Update(eventsLoadedMsg{from: from, to: to, events: events})
	return model.(Model)
}

// press feeds a key press to the model.
func press(t *testing.T, m Model, keys string) Model {
	t.Helper()
	model, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(keys)})
	return model.(Model)
}

func TestViewRendersHeaderGridAndAgenda(t *testing.T) {
	m := loadedModel(t, 120, 40)
	view := plain(m.View())

	now := time.Now()
	for _, want := range []string{
		now.Format("January 2006"), // header title
		"Mon", "Tue", "Sun",        // weekday row
		now.Format("Monday 2 January"), // agenda heading
		"demo calendar",                // provider label
		"tab pane",                     // help line
	} {
		if !strings.Contains(view, want) {
			t.Errorf("view is missing %q\n---\n%s", want, view)
		}
	}
}

func TestViewShowsTodaysEvents(t *testing.T) {
	m := loadedModel(t, 120, 40)

	events := m.store.Day(m.cursor)
	if len(events) == 0 {
		t.Fatal("demo calendar should populate today")
	}

	view := plain(m.View())
	if !strings.Contains(view, events[0].Title) {
		t.Errorf("agenda is missing %q\n---\n%s", events[0].Title, view)
	}
}

func TestNavigationMovesCursorAndMonth(t *testing.T) {
	m := loadedModel(t, 120, 40)
	start := m.cursor

	m = press(t, m, "l")
	if !m.cursor.Equal(start.AddDate(0, 0, 1)) {
		t.Errorf("after l cursor = %s, want next day", m.cursor.Format("2006-01-02"))
	}

	m = press(t, m, "j")
	if !m.cursor.Equal(start.AddDate(0, 0, 8)) {
		t.Errorf("after j cursor = %s, want +7 days", m.cursor.Format("2006-01-02"))
	}

	m = press(t, m, "]")
	if m.cursor.Month() == start.Month() {
		t.Error("] should move to the next month")
	}
	if !strings.Contains(plain(m.View()), m.cursor.Format("January 2006")) {
		t.Error("header should follow the cursor month")
	}

	m = press(t, m, "t")
	if !m.cursor.Equal(start) {
		t.Errorf("t should jump back to today, got %s", m.cursor.Format("2006-01-02"))
	}
}

func TestMonthShiftClampsShortMonths(t *testing.T) {
	// 31 March minus one month must land on 28 February, not 3 March.
	got := shiftMonth(time.Date(2026, time.March, 31, 0, 0, 0, 0, time.UTC), -1)
	want := time.Date(2026, time.February, 28, 0, 0, 0, 0, time.UTC)

	if !got.Equal(want) {
		t.Errorf("shiftMonth = %s, want %s", got.Format("2006-01-02"), want.Format("2006-01-02"))
	}
}

func TestTabFocusesAgendaAndEnterOpensDetail(t *testing.T) {
	m := loadedModel(t, 120, 40)

	model, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = model.(Model)
	if m.focus != focusAgenda {
		t.Fatal("tab should focus the agenda")
	}

	model, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = model.(Model)
	if !m.detail.Visible() {
		t.Fatal("enter should open the event detail overlay")
	}

	selected, ok := m.agenda.Selected()
	if !ok {
		t.Fatal("agenda should have a selected event")
	}
	if !strings.Contains(plain(m.View()), selected.Title) {
		t.Error("detail overlay should show the selected event title")
	}

	model, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if model.(Model).detail.Visible() {
		t.Error("esc should dismiss the overlay")
	}
}

func TestRefreshRefetchesAndClearsCache(t *testing.T) {
	m := loadedModel(t, 120, 40)

	model, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("r")})
	m = model.(Model)

	if cmd == nil {
		t.Fatal("refresh should return a fetch command")
	}
	if m.inFlight == 0 {
		t.Error("refresh should mark a fetch in flight")
	}
	if len(m.store.Day(m.cursor)) != 0 {
		t.Error("refresh should invalidate the cache")
	}

	// Running the returned command must deliver a fresh payload.
	msg := cmd()
	loaded, ok := msg.(eventsLoadedMsg)
	if !ok {
		t.Fatalf("command returned %T, want eventsLoadedMsg", msg)
	}
	if len(loaded.events) == 0 {
		t.Error("refetch returned no events")
	}
}

func TestNarrowTerminalStacksPanes(t *testing.T) {
	m := loadedModel(t, 60, 40)
	if !m.layout.stacked {
		t.Error("a 60-column terminal should stack the panes")
	}
	if plain(m.View()) == "" {
		t.Error("stacked layout rendered nothing")
	}
}

func TestTinyTerminalShowsGuidance(t *testing.T) {
	m := loadedModel(t, 30, 10)
	if !strings.Contains(plain(m.View()), "Terminal too small") {
		t.Error("tiny terminals should explain the minimum size")
	}
}

func TestFetchFailureSurfacesInStatusBar(t *testing.T) {
	m := loadedModel(t, 120, 40)

	model, _ := m.Update(fetchFailedMsg{err: errFake{}})
	if !strings.Contains(plain(model.(Model).View()), "boom") {
		t.Error("status bar should show the fetch error")
	}
}

type errFake struct{}

func (errFake) Error() string { return "boom" }
