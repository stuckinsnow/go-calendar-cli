package calendar

import (
	"regexp"
	"strings"
)

// MaxLinks caps how many links a single event exposes, so they stay addressable
// by a single keypress in the UI.
const MaxLinks = 9

// urlPattern matches bare http(s) URLs in free text.
var urlPattern = regexp.MustCompile(`https?://[^\s<>"'\)\]]+`)

// trailingPunctuation is stripped from URLs harvested from prose.
const trailingPunctuation = ".,;:!?"

// Links returns the URLs found inside the event: its conference link first,
// then any in the location, then any in the description. Duplicates and the
// event's own calendar link are omitted, and the list is capped at MaxLinks.
func (e Event) Links() []string {
	var (
		out  []string
		seen = map[string]bool{e.URL: true}
	)

	add := func(candidates ...string) {
		for _, raw := range candidates {
			link := strings.TrimRight(strings.TrimSpace(raw), trailingPunctuation)
			if link == "" || seen[link] || len(out) >= MaxLinks {
				continue
			}
			seen[link] = true
			out = append(out, link)
		}
	}

	add(e.ConferenceURL)
	add(urlPattern.FindAllString(e.Location, -1)...)
	add(urlPattern.FindAllString(e.Description, -1)...)
	return out
}
