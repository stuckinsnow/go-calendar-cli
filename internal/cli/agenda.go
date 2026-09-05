package cli

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"

	"github.com/stuckinsnow/gcalendar-cli/internal/calendar"
	"github.com/stuckinsnow/gcalendar-cli/internal/config"
	"github.com/stuckinsnow/gcalendar-cli/internal/ui/theme"
)

// maxAgendaDays caps the non-interactive agenda window.
const maxAgendaDays = 90

// runAgenda prints a styled agenda for the next N days.
func runAgenda(ctx context.Context, cfg config.Config, opts options) error {
	days := opts.days
	switch {
	case days < 1:
		days = 1
	case days > maxAgendaDays:
		days = maxAgendaDays
	}

	provider, err := resolveProvider(ctx, cfg, opts)
	if err != nil {
		return err
	}

	from := calendar.StartOfDay(time.Now())
	to := from.AddDate(0, 0, days)

	events, err := provider.Events(ctx, from, to)
	if err != nil {
		return err
	}

	fmt.Print(renderAgenda(theme.New(), cfg, events, from, days))
	return nil
}

// renderAgenda formats grouped events for plain stdout.
func renderAgenda(t theme.Theme, cfg config.Config, events []calendar.Event, from time.Time, days int) string {
	byDay := calendar.GroupByDay(events)
	layoutTime := cfg.TimeLayout()
	now := time.Now()

	var b strings.Builder
	for i := 0; i < days; i++ {
		day := from.AddDate(0, 0, i)
		dayEvents := byDay[day]
		if len(dayEvents) == 0 && days > 1 {
			continue
		}

		label := day.Format("Monday 2 January")
		if calendar.SameDay(day, now) {
			label += "  today"
		}
		b.WriteString(t.AgendaDate.Render(label) + "\n")

		if len(dayEvents) == 0 {
			b.WriteString("  " + t.AgendaEmpty.Render("Nothing scheduled.") + "\n")
		}
		for _, e := range dayEvents {
			b.WriteString(agendaLine(t, e, layoutTime, now) + "\n")
		}
		b.WriteString("\n")
	}
	return b.String()
}

// agendaLine renders one event: coloured bar, time range, title, then meta.
func agendaLine(t theme.Theme, e calendar.Event, layoutTime string, now time.Time) string {
	bar := lipgloss.NewStyle().Foreground(t.EventColor(e.Color)).Render("▌")

	timeStyle, titleStyle := t.EventTime, t.EventTitle
	switch {
	case e.IsNow(now):
		timeStyle, titleStyle = t.EventNow, titleStyle.Bold(true)
	case e.IsPast(now):
		timeStyle, titleStyle = t.EventPast, t.EventPast
	}

	line := fmt.Sprintf("  %s %s  %s", bar,
		timeStyle.Render(fmt.Sprintf("%-13s", e.TimeRange(layoutTime))),
		titleStyle.Render(e.Title))

	var meta []string
	if e.Location != "" {
		meta = append(meta, e.Location)
	}
	if e.Calendar != "" {
		meta = append(meta, e.Calendar)
	}
	if len(meta) > 0 {
		line += "  " + t.EventMeta.Render("· "+strings.Join(meta, " · "))
	}
	return line
}
