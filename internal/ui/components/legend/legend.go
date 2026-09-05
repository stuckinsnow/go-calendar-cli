// Package legend lists the calendars in view with their colours, plus a key to
// the symbols used in the month grid.
package legend

import (
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/stuckinsnow/gcalendar-cli/internal/ui/theme"
)

// New creates an empty legend.
func New(t theme.Theme) Model { return Model{theme: t} }

// SetCalendars replaces the listed calendars, given a name-to-colour map.
func (m *Model) SetCalendars(calendars map[string]string) {
	m.entries = m.entries[:0]
	for name, color := range calendars {
		m.entries = append(m.entries, entry{name: name, color: color})
	}
	sort.Slice(m.entries, func(i, j int) bool { return m.entries[i].name < m.entries[j].name })
}

// SetSize records the space available inside the panel.
func (m *Model) SetSize(width, height int) { m.width, m.height = width, height }

// Empty reports whether there is anything to show.
func (m Model) Empty() bool { return len(m.entries) == 0 }

// View renders the calendar list followed by the grid symbol key.
func (m Model) View() string {
	lines := []string{m.theme.DetailLabel.Render("Calendars")}

	// Leave room for the symbol key.
	budget := m.height - 3
	for i, e := range m.entries {
		if i >= budget && budget > 0 {
			lines = append(lines, m.theme.EventMeta.Render("…"))
			break
		}
		swatch := lipgloss.NewStyle().Foreground(m.theme.EventColor(e.color)).Render("■")
		lines = append(lines, swatch+" "+m.theme.EventTitle.Render(truncate(e.name, m.width-2)))
	}

	lines = append(lines, "", m.theme.EventMeta.Render("• timed   ▪ all day   · today"))
	return strings.Join(lines, "\n")
}

func truncate(s string, width int) string {
	if width < 4 {
		width = 4
	}
	runes := []rune(s)
	if len(runes) <= width {
		return s
	}
	return string(runes[:width-1]) + "…"
}
