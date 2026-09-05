package calendar

import (
	"strings"
	"testing"
)

func TestLinksPrefersConferenceThenLocationThenDescription(t *testing.T) {
	e := Event{
		URL:           "https://calendar.google.com/event?eid=abc",
		ConferenceURL: "https://meet.google.com/abc-defg-hij",
		Location:      "Join at https://zoom.example/j/123",
		Description:   "Agenda: https://docs.example/agenda, notes at https://notes.example/x.",
	}

	got := e.Links()
	want := []string{
		"https://meet.google.com/abc-defg-hij",
		"https://zoom.example/j/123",
		"https://docs.example/agenda",
		"https://notes.example/x",
	}

	if len(got) != len(want) {
		t.Fatalf("got %d links %v, want %d", len(got), got, len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("link %d = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestLinksExcludesOwnURLAndDuplicates(t *testing.T) {
	e := Event{
		URL:         "https://calendar.google.com/event?eid=abc",
		Description: "https://calendar.google.com/event?eid=abc and https://a.example and https://a.example",
	}

	got := e.Links()
	if len(got) != 1 || got[0] != "https://a.example" {
		t.Errorf("links = %v, want only https://a.example", got)
	}
}

func TestLinksCappedAtMaxLinks(t *testing.T) {
	var b strings.Builder
	for i := 0; i < MaxLinks+5; i++ {
		b.WriteString("https://example.com/")
		b.WriteString(string(rune('a' + i)))
		b.WriteString(" ")
	}

	if got := len((Event{Description: b.String()}).Links()); got != MaxLinks {
		t.Errorf("got %d links, want the cap of %d", got, MaxLinks)
	}
}

func TestLinksNoneFound(t *testing.T) {
	e := Event{Title: "Standup", Location: "Room 4B", Description: "no urls here"}
	if got := e.Links(); len(got) != 0 {
		t.Errorf("links = %v, want none", got)
	}
}
