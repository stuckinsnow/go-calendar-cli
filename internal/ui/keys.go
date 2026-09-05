package ui

import "github.com/charmbracelet/bubbles/key"

// KeyMap defines every binding, and doubles as the source for the help view.
type KeyMap struct {
	PrevDay    key.Binding
	NextDay    key.Binding
	PrevWeek   key.Binding
	NextWeek   key.Binding
	PrevMonth  key.Binding
	NextMonth  key.Binding
	Today      key.Binding
	SwitchPane key.Binding
	Open       key.Binding
	Close      key.Binding
	Refresh    key.Binding
	Help       key.Binding
	Quit       key.Binding
}

// DefaultKeyMap returns vim-flavoured bindings with arrow-key equivalents.
func DefaultKeyMap() KeyMap {
	return KeyMap{
		PrevDay: key.NewBinding(
			key.WithKeys("left", "h"),
			key.WithHelp("←/h", "prev day"),
		),
		NextDay: key.NewBinding(
			key.WithKeys("right", "l"),
			key.WithHelp("→/l", "next day"),
		),
		PrevWeek: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "up"),
		),
		NextWeek: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "down"),
		),
		PrevMonth: key.NewBinding(
			key.WithKeys("pgup", "[", "p"),
			key.WithHelp("[/p", "prev month"),
		),
		NextMonth: key.NewBinding(
			key.WithKeys("pgdown", "]", "n"),
			key.WithHelp("]/n", "next month"),
		),
		Today: key.NewBinding(
			key.WithKeys("t", "."),
			key.WithHelp("t", "today"),
		),
		SwitchPane: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "switch pane"),
		),
		Open: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "event details"),
		),
		Close: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "back"),
		),
		Refresh: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "refresh"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "help"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
	}
}

// ShortHelp is the single-line help shown by default. It uses compact synthetic
// bindings so the whole hint fits in a narrow terminal.
func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		key.NewBinding(key.WithKeys("left", "right"), key.WithHelp("←→", "day")),
		key.NewBinding(key.WithKeys("up", "down"), key.WithHelp("↑↓", "week")),
		key.NewBinding(key.WithKeys("[", "]"), key.WithHelp("[]", "month")),
		key.NewBinding(key.WithKeys("t"), key.WithHelp("t", "today")),
		key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "pane")),
		key.NewBinding(key.WithKeys("enter"), key.WithHelp("↵", "details")),
		k.Help,
		k.Quit,
	}
}

// FullHelp is the expanded help grid.
func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.PrevDay, k.NextDay, k.PrevWeek, k.NextWeek},
		{k.PrevMonth, k.NextMonth, k.Today, k.Refresh},
		{k.SwitchPane, k.Open, k.Close, k.Help, k.Quit},
	}
}
