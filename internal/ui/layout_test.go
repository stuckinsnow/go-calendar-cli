package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// terminalSizes covers wide, ordinary, narrow (stacked) and tiny terminals.
var terminalSizes = []struct {
	name          string
	width, height int
}{
	{"wide", 160, 48},
	{"ordinary", 100, 32},
	{"small", 80, 24},
	{"narrow", 60, 30},
	{"tiny", 30, 10},
}

// assertFits checks no rendered line is wider or the view taller than the
// terminal, which is what would otherwise cause wrapping and scroll artefacts.
func assertFits(t *testing.T, view string, width, height int) {
	t.Helper()

	lines := strings.Split(view, "\n")
	if len(lines) > height {
		t.Errorf("view is %d lines, terminal is %d", len(lines), height)
	}
	for i, line := range lines {
		if w := lipgloss.Width(line); w > width {
			t.Errorf("line %d is %d columns, terminal is %d: %q", i, w, width, plain(line))
		}
	}
}

func TestViewFitsEveryTerminalSize(t *testing.T) {
	for _, size := range terminalSizes {
		t.Run(size.name, func(t *testing.T) {
			m := loadedModel(t, size.width, size.height)
			assertFits(t, m.View(), size.width, size.height)
		})
	}
}

func TestViewFitsWithFullHelpAndOverlay(t *testing.T) {
	for _, size := range terminalSizes {
		t.Run(size.name, func(t *testing.T) {
			m := loadedModel(t, size.width, size.height)

			model, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("?")})
			m = model.(Model)
			assertFits(t, m.View(), size.width, size.height)

			model, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
			model, _ = model.(Model).Update(tea.KeyMsg{Type: tea.KeyEnter})
			assertFits(t, model.(Model).View(), size.width, size.height)
		})
	}
}

func TestGridRowsStayAlignedWhenSelectionMoves(t *testing.T) {
	m := loadedModel(t, 100, 32)

	widthsFor := func(m Model) []int {
		var widths []int
		for _, line := range strings.Split(m.month.View(), "\n") {
			widths = append(widths, lipgloss.Width(line))
		}
		return widths
	}

	before := widthsFor(m)
	m = press(t, m, "l") // move the selection to the next day
	after := widthsFor(m)

	if len(before) != len(after) {
		t.Fatalf("grid changed line count: %d then %d", len(before), len(after))
	}
	for i := range before {
		if before[i] != after[i] {
			t.Errorf("grid line %d width changed from %d to %d when the cursor moved",
				i, before[i], after[i])
		}
	}
	for i, w := range after {
		if w != after[0] {
			t.Errorf("grid line %d is %d wide, line 0 is %d: rows are ragged", i, w, after[0])
		}
	}
}

func TestWideLayoutShowsLegend(t *testing.T) {
	m := loadedModel(t, 120, 40)
	if !m.layout.showLegend {
		t.Fatal("a 40-row terminal should have room for the legend")
	}
	if !strings.Contains(plain(m.View()), "Calendars") {
		t.Error("legend heading missing from the view")
	}
}

func TestPanesFillTerminalWidth(t *testing.T) {
	m := loadedModel(t, 100, 32)

	if got := lipgloss.Width(m.panes()); got != 100 {
		t.Errorf("panes are %d columns wide, want the full 100", got)
	}
}
