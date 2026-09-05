package google

import (
	"context"
	"fmt"
	"strings"
	"time"

	gapi "google.golang.org/api/calendar/v3"

	"github.com/stuckinsnow/gcalendar-cli/internal/calendar"
)

// convert maps a Google API event onto the domain model. It reports false for
// entries that cannot be placed on the grid (cancelled or undated).
func convert(item *gapi.Event, ref calendarRef, pal palette) (calendar.Event, bool) {
	if item == nil || item.Status == "cancelled" || item.Start == nil || item.End == nil {
		return calendar.Event{}, false
	}

	start, allDay, err := parseEventTime(item.Start)
	if err != nil {
		return calendar.Event{}, false
	}
	end, _, err := parseEventTime(item.End)
	if err != nil {
		return calendar.Event{}, false
	}

	title := item.Summary
	if title == "" {
		title = "(no title)"
	}

	return calendar.Event{
		ID:            fmt.Sprintf("%s/%s", ref.ID, item.Id),
		Title:         title,
		Start:         start,
		End:           end,
		AllDay:        allDay,
		Location:      strings.TrimSpace(item.Location),
		Description:   cleanDescription(item.Description),
		Calendar:      ref.Summary,
		Color:         firstNonEmpty(pal.event(item.ColorId), ref.Color),
		URL:           item.HtmlLink,
		ConferenceURL: conferenceURL(item),
		Declined:      declined(item),
	}, true
}

// conferenceURL finds the joining link for a meeting, preferring the modern
// conferenceData entry points over the legacy hangoutLink.
func conferenceURL(item *gapi.Event) string {
	if item.ConferenceData != nil {
		for _, ep := range item.ConferenceData.EntryPoints {
			if ep.EntryPointType == "video" && ep.Uri != "" {
				return ep.Uri
			}
		}
		for _, ep := range item.ConferenceData.EntryPoints {
			if ep.Uri != "" {
				return ep.Uri
			}
		}
	}
	return item.HangoutLink
}

// parseEventTime handles both timed events and all-day dates.
func parseEventTime(t *gapi.EventDateTime) (moment time.Time, allDay bool, err error) {
	if t.DateTime != "" {
		moment, err = time.Parse(time.RFC3339, t.DateTime)
		if err != nil {
			return time.Time{}, false, fmt.Errorf("parse event time %q: %w", t.DateTime, err)
		}
		return moment.Local(), false, nil
	}
	if t.Date == "" {
		return time.Time{}, false, fmt.Errorf("event has neither date nor dateTime")
	}
	moment, err = time.ParseInLocation("2006-01-02", t.Date, time.Local)
	if err != nil {
		return time.Time{}, false, fmt.Errorf("parse event date %q: %w", t.Date, err)
	}
	return moment, true, nil
}

// declined reports whether the signed-in user declined the invitation.
func declined(item *gapi.Event) bool {
	for _, a := range item.Attendees {
		if a.Self && a.ResponseStatus == "declined" {
			return true
		}
	}
	return false
}

// cleanDescription strips the HTML that Google sometimes embeds, keeping the
// text readable in a terminal.
func cleanDescription(s string) string {
	if s == "" {
		return ""
	}
	replacer := strings.NewReplacer(
		"<br>", "\n", "<br/>", "\n", "<br />", "\n",
		"</p>", "\n", "&nbsp;", " ", "&amp;", "&", "&lt;", "<", "&gt;", ">", "&quot;", `"`,
	)
	s = replacer.Replace(s)

	var b strings.Builder
	var inTag bool
	for _, r := range s {
		switch {
		case r == '<':
			inTag = true
		case r == '>':
			inTag = false
		case !inTag:
			b.WriteRune(r)
		}
	}
	return strings.TrimSpace(b.String())
}

// palette lookups; a missing ID simply yields no colour.
func (p palette) event(id string) string    { return p.events[id] }
func (p palette) calendar(id string) string { return p.calendars[id] }

// loadPalette fetches the colour definitions once per session. A failure here
// is not fatal: the UI falls back to its own theme colours.
func loadPalette(ctx context.Context, svc *gapi.Service) (palette, error) {
	pal := palette{events: map[string]string{}, calendars: map[string]string{}}

	res, err := svc.Colors.Get().Context(ctx).Do()
	if err != nil {
		return pal, nil //nolint:nilerr // colours are cosmetic
	}
	for id, def := range res.Event {
		pal.events[id] = def.Background
	}
	for id, def := range res.Calendar {
		pal.calendars[id] = def.Background
	}
	return pal, nil
}
