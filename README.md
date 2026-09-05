# VIBE CODED

## gcal

A Google Calendar client for the terminal, built with the Charm stack
(Bubble Tea, Lip Gloss, Bubbles): a month grid with per-day colour-coded event
dots beside a scrollable day agenda.

```
go run ./cmd/gcal -demo
```

## Running

| Command | What it does |
| --- | --- |
| `gcal` | interactive calendar |
| `gcal -demo` | interactive calendar with built-in sample data, no account needed |
| `gcal today` | print today's agenda |
| `gcal agenda -days 14` | print the next 14 days |
| `gcal auth` | sign in to Google |
| `gcal config` | show config paths and which logins are available |

Install as a binary:

```
go build -o ~/.local/bin/gcal ./cmd/gcal
```

## Keys

| Key | Action | Key | Action |
| --- | --- | --- | --- |
| `←`/`h` `→`/`l` | previous / next day | `[` `]` (or `p` `n`) | previous / next month |
| `↑`/`k` `↓`/`j` | previous / next week | `t` | jump to today |
| `tab` | switch pane | `enter` | event details |
| `r` | refresh | `?` | toggle full help |
| `esc` | back | `q` | quit |

When the agenda pane has focus, `↑`/`↓` move between that day's events instead
of between weeks.

## Signing in

Two options; `authMode: "auto"` tries them in this order.

**A. Reuse an existing gcloud login.** No Cloud console visit needed, but the
scope must be requested explicitly — a plain ADC login has no calendar access:

```
gcloud auth application-default login \
  --scopes=https://www.googleapis.com/auth/calendar.readonly,openid,https://www.googleapis.com/auth/userinfo.email
```

gcal picks up the resulting Application Default Credentials automatically. If
the Calendar API is not enabled on your ADC quota project, enable it once with
`gcloud services enable calendar-json.googleapis.com`.

**B. Your own OAuth client.** More clicks, but independent of gcloud and of any
Cloud project quota:

1. https://console.cloud.google.com/apis/credentials — enable the Google
   Calendar API.
2. Create an OAuth client ID, type **Desktop app**.
3. Save the JSON as `~/.config/gcal/credentials.json`.
4. `gcal auth` — opens a browser, then caches a token in
   `~/.config/gcal/token.json`.

Only the read-only scope `calendar.readonly` is ever requested. Tokens are
written with `0600` permissions and never leave the machine.

## Configuration

`~/.config/gcal/config.json` (override the directory with `GCAL_CONFIG_DIR`):

```json
{
  "authMode": "auto",
  "weekStart": 1,
  "twentyFourHour": true,
  "hideDeclined": true,
  "calendars": []
}
```

- `authMode` — `auto`, `client` (own OAuth client only) or `adc` (gcloud only)
- `weekStart` — `0` Sunday … `1` Monday
- `calendars` — restrict to specific calendar IDs or names; empty means all
  calendars selected in Google Calendar

## Layout

```
cmd/gcal              entry point
internal/calendar     domain model: Event, day/grid maths, event cache
internal/provider     demo (offline fixtures) and google (Calendar API)
internal/config       config paths and preferences
internal/cli          argument parsing and non-interactive commands
internal/ui           Bubble Tea root model, key map, layout
internal/ui/theme     all colours and styles
internal/ui/components  month, agenda, detail, header, legend, statusbar
```

Each package keeps its type declarations in `types.go` and its behaviour in
feature-named files. The UI components are pure views: the root model owns the
selected day and the event cache, and pushes state down.

## Development

```
go test ./...     # domain logic plus TUI render, navigation and layout tests
go vet ./...
gofmt -l .
```

The layout tests assert no rendered line exceeds the terminal width at five
terminal sizes, and that the grid rows stay aligned as the selection moves.
