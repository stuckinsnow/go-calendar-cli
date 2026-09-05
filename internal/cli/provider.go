package cli

import (
	"context"
	"errors"
	"fmt"

	"github.com/stuckinsnow/gcalendar-cli/internal/calendar"
	"github.com/stuckinsnow/gcalendar-cli/internal/config"
	"github.com/stuckinsnow/gcalendar-cli/internal/provider/demo"
	"github.com/stuckinsnow/gcalendar-cli/internal/provider/google"
)

// resolveProvider picks the demo calendar or a live Google client.
func resolveProvider(ctx context.Context, cfg config.Config, opts options) (calendar.Provider, error) {
	if opts.demo {
		return demo.New(), nil
	}

	if !cfg.HasCredentials() && !google.ADCAvailable() {
		return nil, fmt.Errorf("%w\n\n%s", google.ErrNoCredentials, setupHint(cfg))
	}

	client, err := google.New(ctx, cfg, true)
	if err != nil {
		if errors.Is(err, google.ErrNoCredentials) {
			return nil, fmt.Errorf("%w\n\n%s", err, setupHint(cfg))
		}
		return nil, err
	}
	return client, nil
}

// setupHint explains how to sign in, tailored to the configured auth mode.
func setupHint(cfg config.Config) string {
	if cfg.AuthMode == config.AuthADC {
		return fmt.Sprintf(`Reuse an existing gcloud login:

  %s

That writes Application Default Credentials to
  %s
The -scopes flag is required: a plain ADC login has no calendar permission.`,
			google.GcloudLoginHint, google.ADCPath())
	}

	hint := fmt.Sprintf(`Sign in with your own OAuth client (read-only, one-time setup):

  1. Sign in to https://console.cloud.google.com with the Google account whose
     calendar you want to read, and create a project.
  2. Enable the Google Calendar API for it.
  3. APIs & Services → Credentials → Create OAuth client ID → "Desktop app".
  4. Save the downloaded JSON as:
       %s
  5. Run: gcal auth

gcal only ever requests %s`, cfg.CredentialsPath(), google.Scope)

	if cfg.AuthMode == config.AuthAuto {
		hint += fmt.Sprintf(`

Alternatively, reuse a gcloud login (signs in as that account):

  %s`, google.GcloudLoginHint)
	}

	return hint + `

Or explore the interface with no account at all:

  gcal -demo`
}
