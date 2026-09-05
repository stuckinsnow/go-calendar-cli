// Package ui wires the Bubble Tea application together: it owns the selected
// day, orchestrates provider fetches and composes the view components.
package ui

import (
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/stuckinsnow/gcalendar-cli/internal/calendar"
	"github.com/stuckinsnow/gcalendar-cli/internal/config"
	"github.com/stuckinsnow/gcalendar-cli/internal/ui/components/agenda"
	"github.com/stuckinsnow/gcalendar-cli/internal/ui/components/detail"
	"github.com/stuckinsnow/gcalendar-cli/internal/ui/components/header"
	"github.com/stuckinsnow/gcalendar-cli/internal/ui/components/legend"
	"github.com/stuckinsnow/gcalendar-cli/internal/ui/components/month"
	"github.com/stuckinsnow/gcalendar-cli/internal/ui/components/statusbar"
	"github.com/stuckinsnow/gcalendar-cli/internal/ui/theme"
)

// New builds the root model for a provider.
func New(cfg config.Config, provider calendar.Provider) Model {
	t := theme.New()
	store := calendar.NewStore()

	sp := spinner.New(spinner.WithSpinner(spinner.Dot))
	sp.Style = lipgloss.NewStyle().Foreground(t.Palette.Primary)

	hp := help.New()
	hp.Styles = t.Help

	m := Model{
		cfg:      cfg,
		provider: provider,
		store:    store,
		theme:    t,
		keys:     DefaultKeyMap(),
		help:     hp,
		spinner:  sp,
		header:   header.New(t, provider.Name()),
		month:    month.New(t, cfg.WeekStart, store),
		legend:   legend.New(t),
		agenda:   agenda.New(t, cfg.TimeLayout()),
		detail:   detail.New(t, cfg.TimeLayout()),
		status:   statusbar.New(t),
		cursor:   calendar.StartOfDay(time.Now()),
		inFlight: 1, // the initial fetch issued by Init
	}
	m.status.SetLoading(true, sp.View())
	m.syncPanes()
	return m
}

// Init starts the spinner, the clock and the first fetch.
func (m Model) Init() tea.Cmd {
	from, to := gridWindow(m.cursor, m.cfg.WeekStart)
	return tea.Batch(m.spinner.Tick, clockTick(), loadEvents(m.provider, from, to))
}

// Update handles all incoming messages.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.ready = true
		m.applyLayout()
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		m.status.SetLoading(m.inFlight > 0, m.spinner.View())
		return m, cmd

	case clockTickMsg:
		m.month.SetToday(time.Time(msg))
		m.syncPanes()
		return m, clockTick()

	case eventsLoadedMsg:
		m.inFlight = decrement(m.inFlight)
		m.store.Replace(msg.from, msg.to, msg.events)
		m.status.MarkSynced(time.Now())
		m.status.SetLoading(m.inFlight > 0, m.spinner.View())
		m.syncPanes()
		return m, nil

	case fetchFailedMsg:
		m.inFlight = decrement(m.inFlight)
		m.status.SetError(msg.err)
		m.status.SetLoading(m.inFlight > 0, m.spinner.View())
		return m, nil
	}

	var cmd tea.Cmd
	m.agenda, cmd = m.agenda.Update(msg)
	return m, cmd
}

// handleKey routes key presses. The detail pane is a normal column, so unlike a
// modal it does not swallow navigation keys; only the overlay fallback does.
func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if key.Matches(msg, m.keys.Quit) {
		return m, tea.Quit
	}

	if m.detail.Visible() && !m.layout.showDetail {
		if key.Matches(msg, m.keys.Close, m.keys.Open) {
			m.detail.Hide()
			m.applyLayout()
		}
		return m, nil
	}

	switch {
	case key.Matches(msg, m.keys.Help):
		m.help.ShowAll = !m.help.ShowAll
		m.applyLayout()

	case key.Matches(msg, m.keys.SwitchPane):
		m.toggleFocus()

	case key.Matches(msg, m.keys.Open):
		return m.openSelection()

	case key.Matches(msg, m.keys.Close):
		if m.detail.Visible() {
			m.detail.Hide()
			m.applyLayout()
			return m, nil
		}
		m.setFocus(focusMonth)

	case key.Matches(msg, m.keys.Copy):
		m.copySelection()

	case key.Matches(msg, m.keys.CopyLink):
		if len(msg.Runes) > 0 {
			m.copyLink(int(msg.Runes[0] - '0'))
		}

	case key.Matches(msg, m.keys.Refresh):
		return m.refresh()

	case key.Matches(msg, m.keys.Today):
		return m.moveTo(time.Now())

	case key.Matches(msg, m.keys.PrevDay):
		return m.moveTo(m.cursor.AddDate(0, 0, -1))

	case key.Matches(msg, m.keys.NextDay):
		return m.moveTo(m.cursor.AddDate(0, 0, 1))

	case key.Matches(msg, m.keys.PrevWeek):
		if m.focus == focusAgenda {
			m.agenda.MoveCursor(-1)
			m.trackSelection()
			return m, nil
		}
		return m.moveTo(m.cursor.AddDate(0, 0, -7))

	case key.Matches(msg, m.keys.NextWeek):
		if m.focus == focusAgenda {
			m.agenda.MoveCursor(1)
			m.trackSelection()
			return m, nil
		}
		return m.moveTo(m.cursor.AddDate(0, 0, 7))

	case key.Matches(msg, m.keys.PrevMonth):
		return m.moveTo(shiftMonth(m.cursor, -1))

	case key.Matches(msg, m.keys.NextMonth):
		return m.moveTo(shiftMonth(m.cursor, 1))
	}
	return m, nil
}

// openSelection moves focus into the agenda, or opens the selected event.
func (m Model) openSelection() (tea.Model, tea.Cmd) {
	if m.focus == focusMonth {
		if len(m.store.Day(m.cursor)) == 0 {
			m.status.SetMessage("nothing scheduled on " + m.cursor.Format("2 Jan"))
			return m, nil
		}
		m.setFocus(focusAgenda)
		m.trackSelection()
		return m, nil
	}
	if event, ok := m.agenda.Selected(); ok {
		m.detail.Show(event)
		m.applyLayout()
	}
	return m, nil
}

// trackSelection keeps an open detail view on the currently selected event, so
// the side pane follows the agenda cursor instead of freezing on one event.
func (m *Model) trackSelection() {
	if !m.detail.Visible() {
		return
	}
	if event, ok := m.agenda.Selected(); ok {
		m.detail.Show(event)
	}
}

// refresh drops the cache and refetches the visible window.
func (m Model) refresh() (tea.Model, tea.Cmd) {
	m.store.Invalidate()
	m.inFlight++
	m.status.SetMessage("refreshing")
	m.status.SetLoading(true, m.spinner.View())

	from, to := gridWindow(m.cursor, m.cfg.WeekStart)
	return m, loadEvents(m.provider, from, to)
}

// moveTo selects a new day, fetching its month when not yet cached.
func (m Model) moveTo(day time.Time) (tea.Model, tea.Cmd) {
	previous := m.cursor
	m.cursor = calendar.StartOfDay(day)
	m.syncPanes()

	if m.cursor.Month() == previous.Month() && m.cursor.Year() == previous.Year() {
		return m, nil
	}
	if m.store.HasMonth(m.cursor) {
		return m, nil
	}

	m.inFlight++
	m.status.SetLoading(true, m.spinner.View())
	from, to := gridWindow(m.cursor, m.cfg.WeekStart)
	return m, loadEvents(m.provider, from, to)
}

// toggleFocus swaps between the grid and the agenda.
func (m *Model) toggleFocus() {
	if m.focus == focusMonth {
		m.setFocus(focusAgenda)
		return
	}
	m.setFocus(focusMonth)
}

// setFocus applies a focus area to the child components.
func (m *Model) setFocus(area focusArea) {
	m.focus = area
	m.month.Focus(area == focusMonth)
	m.agenda.Focus(area == focusAgenda)
}

// syncPanes pushes the current selection into the child components.
func (m *Model) syncPanes() {
	m.month.SetCursor(m.cursor)
	m.header.SetMonth(m.cursor)
	m.agenda.SetDay(m.cursor, m.store.Day(m.cursor))
	m.legend.SetCalendars(m.store.Calendars())
	m.month.Focus(m.focus == focusMonth)
	m.agenda.Focus(m.focus == focusAgenda)
	m.trackSelection()
}

// applyLayout recomputes geometry and resizes every component.
//
// The grid quantises its width to whole cells, so any slack is handed to the
// agenda; that keeps the panes flush with the terminal edge.
func (m *Model) applyLayout() {
	chrome := m.header.Height() + m.footerHeight()
	m.layout = computeLayout(m.width, m.height, chrome, m.month.Height(), m.detail.Visible())

	m.header.SetWidth(m.width)
	m.status.SetWidth(m.width)
	m.help.Width = m.width

	m.month.SetWidth(m.layout.month.InnerWidth())
	if slack := m.layout.month.InnerWidth() - m.month.Width(); slack > 0 && !m.layout.stacked {
		m.layout.month.Width -= slack
		m.layout.legend.Width -= slack
		m.layout.agenda.Width += slack
	}

	m.legend.SetSize(m.layout.legend.InnerWidth(), m.layout.legend.InnerHeight())
	m.agenda.SetSize(m.layout.agenda.InnerWidth(), m.layout.agenda.InnerHeight()-1) // 1 = date heading
	m.detail.SetPaneWidth(m.layout.detail.InnerWidth())
	m.detail.SetSize(m.width, m.height)
	m.syncPanes()
}

// footerHeight measures the bottom chrome: one combined row of shortcuts and
// sync state, or the help grid plus a status row when help is expanded.
func (m Model) footerHeight() int {
	if m.help.ShowAll {
		return lipgloss.Height(m.help.View(m.keys)) + m.status.Height()
	}
	return 1
}

// shiftMonth moves by n months, clamping the day to the target month's length.
func shiftMonth(t time.Time, n int) time.Time {
	first := calendar.StartOfMonth(t).AddDate(0, n, 0)
	last := first.AddDate(0, 1, -1).Day()

	day := t.Day()
	if day > last {
		day = last
	}
	return time.Date(first.Year(), first.Month(), day, 0, 0, 0, 0, t.Location())
}

func decrement(n int) int {
	if n <= 0 {
		return 0
	}
	return n - 1
}
