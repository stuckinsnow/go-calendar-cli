// Package theme centralises every colour and style used by the UI, so the look
// of the app can be adjusted in one place.
package theme

import (
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/lipgloss"
)

// Palette is the semantic colour set. Adaptive colours pick a variant based on
// the terminal background.
type Palette struct {
	Primary   lipgloss.AdaptiveColor // headings, focused borders
	Secondary lipgloss.AdaptiveColor // accents, today marker
	Text      lipgloss.AdaptiveColor // body text
	Subtle    lipgloss.AdaptiveColor // secondary text
	Faint     lipgloss.AdaptiveColor // out-of-month days, past events
	Border    lipgloss.AdaptiveColor // unfocused borders
	Surface   lipgloss.AdaptiveColor // selection background
	Success   lipgloss.AdaptiveColor
	Warning   lipgloss.AdaptiveColor
	Error     lipgloss.AdaptiveColor
	Inverted  lipgloss.AdaptiveColor // text on a coloured background
}

// DefaultPalette is a Charm-flavoured purple and pink scheme.
func DefaultPalette() Palette {
	return Palette{
		Primary:   lipgloss.AdaptiveColor{Light: "#5A3FD6", Dark: "#7D56F4"},
		Secondary: lipgloss.AdaptiveColor{Light: "#D6297E", Dark: "#FF5FAF"},
		Text:      lipgloss.AdaptiveColor{Light: "#1A1A26", Dark: "#E6E6F0"},
		Subtle:    lipgloss.AdaptiveColor{Light: "#5C5C70", Dark: "#A0A0B8"},
		Faint:     lipgloss.AdaptiveColor{Light: "#A8A8B8", Dark: "#5C5C70"},
		Border:    lipgloss.AdaptiveColor{Light: "#C8C8D8", Dark: "#3A3A50"},
		Surface:   lipgloss.AdaptiveColor{Light: "#E8E4FF", Dark: "#2E2545"},
		Success:   lipgloss.AdaptiveColor{Light: "#1F9254", Dark: "#43BF6D"},
		Warning:   lipgloss.AdaptiveColor{Light: "#B26B00", Dark: "#F5A623"},
		Error:     lipgloss.AdaptiveColor{Light: "#C4314B", Dark: "#FF5F87"},
		Inverted:  lipgloss.AdaptiveColor{Light: "#FFFFFF", Dark: "#11111B"},
	}
}

// Theme bundles the palette with ready-made styles.
type Theme struct {
	Palette Palette

	// Chrome
	Title    lipgloss.Style
	Subtitle lipgloss.Style
	Panel    lipgloss.Style
	Status   lipgloss.Style
	Key      lipgloss.Style
	Value    lipgloss.Style
	Badge    lipgloss.Style
	ErrorMsg lipgloss.Style

	// Month grid
	Weekday      lipgloss.Style
	WeekNumber   lipgloss.Style
	Day          lipgloss.Style
	DayOutside   lipgloss.Style
	DayWeekend   lipgloss.Style
	DayToday     lipgloss.Style
	DaySelected  lipgloss.Style
	DayFocusMark lipgloss.Style

	// Agenda and detail
	AgendaDate     lipgloss.Style
	AgendaEmpty    lipgloss.Style
	EventTitle     lipgloss.Style
	EventTime      lipgloss.Style
	EventMeta      lipgloss.Style
	EventPast      lipgloss.Style
	EventNow       lipgloss.Style
	EventSelected  lipgloss.Style
	DetailBox      lipgloss.Style
	DetailLabel    lipgloss.Style
	DetailBodyText lipgloss.Style

	Help help.Styles
}

// New builds the default theme.
func New() Theme { return FromPalette(DefaultPalette()) }

// FromPalette derives every style from a palette.
func FromPalette(p Palette) Theme {
	base := lipgloss.NewStyle()

	t := Theme{
		Palette: p,

		Title:    base.Foreground(p.Secondary).Bold(true),
		Subtitle: base.Foreground(p.Subtle),
		Panel: base.Border(lipgloss.RoundedBorder()).
			BorderForeground(p.Border).
			Padding(0, 1),
		Status:   base.Foreground(p.Subtle),
		Key:      base.Foreground(p.Primary).Bold(true),
		Value:    base.Foreground(p.Text),
		Badge:    base.Foreground(p.Inverted).Background(p.Primary).Padding(0, 1).Bold(true),
		ErrorMsg: base.Foreground(p.Error).Bold(true),

		Weekday:      base.Foreground(p.Subtle).Bold(true).Align(lipgloss.Center),
		WeekNumber:   base.Foreground(p.Faint),
		Day:          base.Foreground(p.Text),
		DayOutside:   base.Foreground(p.Faint),
		DayWeekend:   base.Foreground(p.Subtle),
		DayToday:     base.Foreground(p.Secondary).Bold(true),
		DaySelected:  base.Foreground(p.Text).Background(p.Surface),
		DayFocusMark: base.Foreground(p.Primary).Bold(true),

		AgendaDate:     base.Foreground(p.Primary).Bold(true),
		AgendaEmpty:    base.Foreground(p.Faint).Italic(true),
		EventTitle:     base.Foreground(p.Text),
		EventTime:      base.Foreground(p.Subtle),
		EventMeta:      base.Foreground(p.Faint),
		EventPast:      base.Foreground(p.Faint).Strikethrough(false),
		EventNow:       base.Foreground(p.Success).Bold(true),
		EventSelected:  base.Background(p.Surface),
		DetailBox:      base.Border(lipgloss.RoundedBorder()).BorderForeground(p.Secondary).Padding(1, 2),
		DetailLabel:    base.Foreground(p.Subtle),
		DetailBodyText: base.Foreground(p.Text),
	}

	t.Help = help.Styles{
		ShortKey:       base.Foreground(p.Subtle),
		ShortDesc:      base.Foreground(p.Faint),
		ShortSeparator: base.Foreground(p.Border),
		Ellipsis:       base.Foreground(p.Faint),
		FullKey:        base.Foreground(p.Subtle),
		FullDesc:       base.Foreground(p.Faint),
		FullSeparator:  base.Foreground(p.Border),
	}
	return t
}

// FocusedPanel returns the panel border styled to show keyboard focus.
func (t Theme) FocusedPanel() lipgloss.Style {
	return t.Panel.BorderForeground(t.Palette.Primary)
}

// EventColor resolves a provider-supplied hex colour, falling back to the
// theme's primary when the provider gave none.
func (t Theme) EventColor(hex string) lipgloss.TerminalColor {
	if hex == "" {
		return t.Palette.Primary
	}
	return lipgloss.Color(hex)
}
