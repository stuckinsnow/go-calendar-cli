package google

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	gapi "google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"

	"github.com/stuckinsnow/gcalendar-cli/internal/calendar"
	"github.com/stuckinsnow/gcalendar-cli/internal/config"
)

// maxPages bounds pagination per calendar as a safety valve.
const maxPages = 20

// New authorises against Google and loads the calendar list. Set interactive to
// allow the browser consent flow; pass false for non-interactive contexts.
func New(ctx context.Context, cfg config.Config, interactive bool) (*Client, error) {
	client, method, err := resolveClient(ctx, cfg, interactive)
	if err != nil {
		return nil, err
	}

	svc, err := gapi.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("create calendar service: %w", err)
	}

	c := &Client{svc: svc, cfg: cfg, method: method}
	if c.palette, err = loadPalette(ctx, svc); err != nil {
		return nil, err
	}
	if err := c.loadCalendars(ctx); err != nil {
		return nil, scopeHint(err)
	}
	return c, nil
}

// Name identifies the signed-in account for the status bar.
func (c *Client) Name() string {
	if c.account != "" {
		return c.account
	}
	return "google calendar"
}

// Method describes how the session authenticated.
func (c *Client) AuthMethod() string { return string(c.method) }

// Calendars returns the calendars being read, in display order.
func (c *Client) Calendars() []string {
	out := make([]string, 0, len(c.cals))
	for _, cal := range c.cals {
		out = append(out, cal.Summary)
	}
	return out
}

// loadCalendars fetches the calendar list, honouring any configured allow-list.
func (c *Client) loadCalendars(ctx context.Context) error {
	allowed := make(map[string]bool, len(c.cfg.Calendars))
	for _, id := range c.cfg.Calendars {
		allowed[id] = true
	}

	var pageToken string
	for page := 0; page < maxPages; page++ {
		call := c.svc.CalendarList.List().ShowHidden(false).MaxResults(250)
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		res, err := call.Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("list calendars: %w", err)
		}

		for _, item := range res.Items {
			if item.Selected == false && !item.Primary && len(allowed) == 0 {
				continue // respect the user's Google-side visibility choices
			}
			if len(allowed) > 0 && !allowed[item.Id] && !allowed[item.Summary] {
				continue
			}
			if item.Primary {
				c.account = item.Id
			}
			c.cals = append(c.cals, calendarRef{
				ID:      item.Id,
				Summary: displayName(item),
				Color:   firstNonEmpty(item.BackgroundColor, c.palette.calendar(item.ColorId)),
				Primary: item.Primary,
			})
		}

		if res.NextPageToken == "" {
			break
		}
		pageToken = res.NextPageToken
	}

	if len(c.cals) == 0 {
		return fmt.Errorf("no readable calendars found")
	}
	sort.SliceStable(c.cals, func(i, j int) bool {
		if c.cals[i].Primary != c.cals[j].Primary {
			return c.cals[i].Primary
		}
		return c.cals[i].Summary < c.cals[j].Summary
	})
	return nil
}

// Events fetches every event overlapping [from, to) across all calendars.
func (c *Client) Events(ctx context.Context, from, to time.Time) ([]calendar.Event, error) {
	var (
		wg      sync.WaitGroup
		mu      sync.Mutex
		out     []calendar.Event
		firstEr error
	)

	for _, ref := range c.cals {
		wg.Add(1)
		go func(ref calendarRef) {
			defer wg.Done()
			events, err := c.eventsFor(ctx, ref, from, to)

			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				if firstEr == nil {
					firstEr = err
				}
				return
			}
			out = append(out, events...)
		}(ref)
	}
	wg.Wait()

	if firstEr != nil {
		return nil, scopeHint(firstEr)
	}
	calendar.Sort(out)
	return out, nil
}

// eventsFor pages through a single calendar, expanding recurring events.
func (c *Client) eventsFor(ctx context.Context, ref calendarRef, from, to time.Time) ([]calendar.Event, error) {
	var (
		out       []calendar.Event
		pageToken string
	)

	for page := 0; page < maxPages; page++ {
		call := c.svc.Events.List(ref.ID).
			TimeMin(from.Format(time.RFC3339)).
			TimeMax(to.Format(time.RFC3339)).
			SingleEvents(true).
			OrderBy("startTime").
			MaxResults(2500)
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}

		res, err := call.Context(ctx).Do()
		if err != nil {
			return nil, fmt.Errorf("list events for %s: %w", ref.Summary, err)
		}

		for _, item := range res.Items {
			event, ok := convert(item, ref, c.palette)
			if !ok || (c.cfg.HideDeclined && event.Declined) {
				continue
			}
			out = append(out, event)
		}

		if res.NextPageToken == "" {
			break
		}
		pageToken = res.NextPageToken
	}
	return out, nil
}

func displayName(item *gapi.CalendarListEntry) string {
	if item.SummaryOverride != "" {
		return item.SummaryOverride
	}
	if item.Summary != "" {
		return item.Summary
	}
	return item.Id
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}
