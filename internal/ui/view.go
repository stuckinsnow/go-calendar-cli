package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// View composes the header, panes, status bar and help, then lays the detail
// overlay on top.
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
		m.status.View(),
		m.helpView(),
	)
	return m.detail.Overlay(screen)
}

// panes renders the calendar column beside the agenda, or stacked when narrow.
func (m Model) panes() string {
	left := m.calendarColumn()
	agenda := m.panelFor(focusAgenda, m.layout.agenda).Render(m.agenda.View())

	if m.layout.stacked {
		return lipgloss.JoinVertical(lipgloss.Left, left, agenda)
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, left, strings.Repeat(" ", paneGap), agenda)
}

// calendarColumn is the month grid, with the legend beneath it when it fits.
func (m Model) calendarColumn() string {
	grid := m.panelFor(focusMonth, m.layout.month).Render(m.month.View())
	if !m.layout.showLegend || m.legend.Empty() {
		return grid
	}
	legend := m.theme.Panel.
		Width(m.layout.legend.Width).
		Height(m.layout.legend.Height).
		Render(m.legend.View())

	return lipgloss.JoinVertical(lipgloss.Left, grid, legend)
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
