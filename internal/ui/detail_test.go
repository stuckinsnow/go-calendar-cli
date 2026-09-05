package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// openDetail focuses the agenda and opens the detail view for its selection.
func openDetail(t *testing.T, m Model) Model {
	t.Helper()

	model, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	model, _ = model.(Model).Update(tea.KeyMsg{Type: tea.KeyEnter})

	m = model.(Model)
	if !m.detail.Visible() {
		t.Fatal("detail should be visible after enter")
	}
	return m
}

func TestDetailRendersAsSidePaneOnWideTerminals(t *testing.T) {
	m := openDetail(t, loadedModel(t, 160, 44))

	if !m.layout.showDetail {
		t.Fatal("a 160-column terminal should host the detail column")
	}
	if m.layout.detail.Width < 38 {
		t.Errorf("detail pane is %d columns, expected at least 38", m.layout.detail.Width)
	}

	// Three panes plus two gaps must exactly fill the terminal.
	if got := lipgloss.Width(m.panes()); got != 160 {
		t.Errorf("panes span %d columns, want 160", got)
	}
	assertFits(t, m.View(), 160, 44)
}

func TestDetailFallsBackToOverlayWhenNarrow(t *testing.T) {
	m := openDetail(t, loadedModel(t, 80, 24))

	if m.layout.showDetail {
		t.Error("an 80-column terminal has no room for a third column")
	}
	selected, _ := m.agenda.Selected()
	if !strings.Contains(plain(m.View()), selected.Title) {
		t.Error("overlay should still show the event")
	}
}

func TestDetailPaneFollowsAgendaCursor(t *testing.T) {
	m := openDetail(t, loadedModel(t, 160, 44))

	first := m.detail.Event()
	m = press(t, m, "j")
	second := m.detail.Event()

	if first.ID == second.ID {
		t.Fatal("moving the agenda cursor should update the detail pane")
	}
	if selected, _ := m.agenda.Selected(); selected.ID != second.ID {
		t.Errorf("detail shows %q but agenda has %q selected", second.Title, selected.Title)
	}
}

func TestNavigationStillWorksWithDetailPaneOpen(t *testing.T) {
	m := openDetail(t, loadedModel(t, 160, 44))
	start := m.cursor

	m = press(t, m, "l")
	if !m.cursor.Equal(start.AddDate(0, 0, 1)) {
		t.Error("the detail pane must not swallow day navigation")
	}
}

func TestEscapeClosesDetailPaneAndReclaimsWidth(t *testing.T) {
	m := openDetail(t, loadedModel(t, 160, 44))
	withDetail := m.layout.agenda.Width

	model, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = model.(Model)

	if m.detail.Visible() {
		t.Fatal("esc should close the detail pane")
	}
	if m.layout.agenda.Width <= withDetail {
		t.Errorf("agenda should reclaim width: %d then %d", withDetail, m.layout.agenda.Width)
	}
	if got := lipgloss.Width(m.panes()); got != 160 {
		t.Errorf("panes span %d columns after closing, want 160", got)
	}
}

func TestFooterPutsSyncStateOnTheShortcutLine(t *testing.T) {
	m := loadedModel(t, 120, 40)

	footer := plain(m.footer())
	if lines := strings.Count(footer, "\n"); lines != 0 {
		t.Fatalf("footer should be one line, got %d", lines+1)
	}
	if !strings.Contains(footer, "synced") {
		t.Error("footer is missing the sync state")
	}
	if !strings.Contains(footer, "quit") {
		t.Error("footer is missing the shortcuts")
	}
}

func TestHeaderHasNoRule(t *testing.T) {
	m := loadedModel(t, 120, 40)

	header := plain(m.header.View())
	if strings.Count(header, "\n") != 0 {
		t.Error("header should be a single line with no rule beneath it")
	}
	if strings.Contains(header, "───") {
		t.Error("header should not draw a horizontal rule")
	}
}

func TestGridKeepsItsWidthOnWideTerminals(t *testing.T) {
	closed := loadedModel(t, 200, 44)
	opened := openDetail(t, closed)

	if !opened.layout.showDetail {
		t.Fatal("a 200-column terminal should host the detail column")
	}
	if opened.layout.month.Width != closed.layout.month.Width {
		t.Errorf("grid shrank from %d to %d columns when the details opened; "+
			"wide terminals have room to spare",
			closed.layout.month.Width, opened.layout.month.Width)
	}
}

func TestGridYieldsWidthOnlyWhenNeeded(t *testing.T) {
	// 130 columns cannot fit a full-size grid, the agenda and the details, so
	// here the grid is expected to give ground rather than lose the pane.
	closed := loadedModel(t, 130, 40)
	opened := openDetail(t, closed)

	if !opened.layout.showDetail {
		t.Fatal("130 columns should still manage three panes with a narrow grid")
	}
	if opened.layout.month.Width >= closed.layout.month.Width {
		t.Errorf("grid should have narrowed: %d then %d",
			closed.layout.month.Width, opened.layout.month.Width)
	}
	assertFits(t, opened.View(), 130, 40)
}

func TestDetailPaneListsCopyableLinks(t *testing.T) {
	m := openDetail(t, loadedModel(t, 160, 44))

	// Walk the day's events to find one carrying links.
	for i := 0; i < 6; i++ {
		if event := m.detail.Event(); len(event.Links()) > 0 {
			view := plain(m.View())
			if !strings.Contains(view, "Links") {
				t.Fatalf("detail pane omits the Links section for %q", event.Title)
			}
			if !strings.Contains(view, event.Links()[0]) {
				t.Errorf("detail pane omits the link %q", event.Links()[0])
			}
			return
		}
		m = press(t, m, "j")
	}
	t.Skip("no demo event with links on today's agenda")
}

func TestCopyLinkReportsMissingIndex(t *testing.T) {
	m := openDetail(t, loadedModel(t, 160, 44))

	m.copyLink(9) // demo events never have nine links
	if !strings.Contains(plain(m.status.View()), "no link 9") {
		t.Errorf("status should explain the missing link: %q", plain(m.status.View()))
	}
}
