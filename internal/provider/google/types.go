package google

import (
	gapi "google.golang.org/api/calendar/v3"

	"github.com/stuckinsnow/gcalendar-cli/internal/config"
)

// authMethod records how the session authenticated, for display in the UI.
type authMethod string

const (
	// methodOAuthClient used our own OAuth client plus a cached user token.
	methodOAuthClient authMethod = "oauth client"
	// methodADC reused Application Default Credentials, e.g. the login left
	// behind by `gcloud auth application-default login`.
	methodADC authMethod = "application default credentials"
)

// Client is a calendar.Provider backed by the Google Calendar API.
type Client struct {
	svc     *gapi.Service
	cfg     config.Config
	method  authMethod
	account string
	cals    []calendarRef
	palette palette
}

// calendarRef is a calendar the account can read.
type calendarRef struct {
	ID      string
	Summary string
	Color   string
	Primary bool
}

// palette resolves Google's numeric colour IDs to hex values.
type palette struct {
	events    map[string]string
	calendars map[string]string
}
