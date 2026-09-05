package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// View composes the header, panes, status bar and help. On terminals too narrow
// for a detail column, the details are laid over the top instead.
func (m Model) View() string {
	if !m.ready {
		return m.theme.Subtitle.Render("starting…")
	}
	if m.layout.tooSmall {
		return m.tooSmallView()
	}

	screen := lipgloss.JoinVertical(lipgloss.Left,
		m.header.View(),
		m.panes(),
		m.footer(),
	)
	if m.layout.showDetail {
		return screen
	}
	return m.detail.Overlay(screen)
}

// footer is the bottom row: shortcuts on the left, sync state on the right. In
// expanded help mode the help grid sits above that row.
func (m Model) footer() string {
	hints := m.help.View(m.keys)
	if m.help.ShowAll {
		return lipgloss.JoinVertical(lipgloss.Left,
			lipgloss.NewStyle().MaxWidth(m.width).Render(hints),
			m.status.View(),
		)
	}
	return lipgloss.NewStyle().MaxWidth(m.width).Render(m.status.Footer(hints))
}

// panes renders the calendar column, the agenda and, when it fits, the event
// details; or grid above agenda when the terminal is narrow.
func (m Model) panes() string {
	columns := []string{m.calendarColumn()}

	columns = append(columns,
		m.panelFor(focusAgenda, m.layout.agenda).Render(clip(m.agenda.View(), m.layout.agenda)))

	if m.layout.showDetail {
		columns = append(columns,
			m.theme.DetailPanel().
				Width(m.layout.detail.Width).
				Height(m.layout.detail.Height).
				Render(clip(m.detail.Pane(), m.layout.detail)))
	}

	if m.layout.stacked {
		return lipgloss.JoinVertical(lipgloss.Left, columns...)
	}
	return joinColumns(columns)
}

// calendarColumn is the month grid, with the legend beneath it when it fits.
func (m Model) calendarColumn() string {
	grid := m.panelFor(focusMonth, m.layout.month).Render(clip(m.month.View(), m.layout.month))
	if !m.layout.showLegend || m.legend.Empty() {
		return grid
	}
	legend := m.theme.Panel.
		Width(m.layout.legend.Width).
		Height(m.layout.legend.Height).
		Render(clip(m.legend.View(), m.layout.legend))

	return lipgloss.JoinVertical(lipgloss.Left, grid, legend)
}

// clip guarantees a component cannot push its panel past the geometry it was
// given, whatever it renders.
func clip(content string, p pane) string {
	return lipgloss.NewStyle().
		MaxWidth(p.InnerWidth()).
		MaxHeight(p.Height).
		Render(content)
}

// joinColumns places panes side by side with a single-column gap between them.
func joinColumns(columns []string) string {
	spaced := make([]string, 0, len(columns)*2-1)
	for i, col := range columns {
		if i > 0 {
			spaced = append(spaced, strings.Repeat(" ", paneGap))
		}
		spaced = append(spaced, col)
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, spaced...)
}

// panelFor returns a sized panel style, highlighting the focused pane.
func (m Model) panelFor(area focusArea, p pane) lipgloss.Style {
	style := m.theme.Panel
	if m.focus == area {
		style = m.theme.FocusedPanel()
	}
	return style.Width(p.Width).Height(p.Height)
}

// helpView renders the key hints, hard-clipped to the terminal width because
// bubbles/help can overflow when there is no room for its ellipsis.
func (m Model) helpView() string {
	return lipgloss.NewStyle().MaxWidth(m.width).Render(m.help.View(m.keys))
}

// tooSmallView explains what the terminal needs to be.
func (m Model) tooSmallView() string {
	msg := lipgloss.JoinVertical(lipgloss.Center,
		m.theme.Title.Render("Terminal too small"),
		m.theme.Subtitle.Render(fmt.Sprintf("needs at least %d×%d", minWidth, minHeight)),
	)
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, msg)
}
