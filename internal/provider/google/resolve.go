package google

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"golang.org/x/oauth2"
	googleauth "golang.org/x/oauth2/google"
	gapi "google.golang.org/api/calendar/v3"

	"github.com/stuckinsnow/gcalendar-cli/internal/config"
)

// Scope is the only permission this tool ever asks for: read-only calendars.
const Scope = gapi.CalendarReadonlyScope

// ErrNoCredentials is returned when neither an OAuth client secret nor
// Application Default Credentials are available.
var ErrNoCredentials = errors.New("no Google credentials found")

// resolveClient picks an authentication method and returns an authorised client.
//
// Resolution order, matching config.AuthMode:
//  1. our own OAuth client (credentials.json + cached token.json)
//  2. Application Default Credentials, i.e. an existing gcloud login or a
//     service account pointed to by GOOGLE_APPLICATION_CREDENTIALS
//
// interactive allows the browser consent flow; pass false for scripts.
func resolveClient(ctx context.Context, cfg config.Config, interactive bool) (*http.Client, authMethod, error) {
	mode := cfg.AuthMode

	if mode == config.AuthClient || mode == config.AuthAuto {
		if cfg.HasCredentials() {
			client, err := oauthClient(ctx, cfg, interactive)
			return client, methodOAuthClient, err
		}
		if mode == config.AuthClient {
			return nil, "", fmt.Errorf("%w at %s", ErrNoCredentials, cfg.CredentialsPath())
		}
	}

	if mode == config.AuthADC || mode == config.AuthAuto {
		client, err := adcClient(ctx)
		if err == nil {
			return client, methodADC, nil
		}
		if mode == config.AuthADC {
			return nil, "", err
		}
	}

	return nil, "", ErrNoCredentials
}

// adcClient reuses Application Default Credentials already present on the
// machine. This is what makes `gcloud auth application-default login` work as a
// login method: gcloud writes a user token we can borrow.
//
// The well-known file is checked first so we never probe GCE metadata on a
// laptop, which would only add latency before failing.
func adcClient(ctx context.Context) (*http.Client, error) {
	if !hasADC() {
		return nil, fmt.Errorf("no application default credentials on this machine")
	}

	creds, err := googleauth.FindDefaultCredentials(ctx, Scope)
	if err != nil {
		return nil, fmt.Errorf("load application default credentials: %w", err)
	}
	return oauth2.NewClient(ctx, creds.TokenSource), nil
}

// hasADC reports whether ADC are configured, without making any network calls.
func hasADC() bool {
	if path := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"); path != "" {
		if _, err := os.Stat(path); err == nil {
			return true
		}
	}
	if _, err := os.Stat(ADCPath()); err == nil {
		return true
	}
	return false
}

// ADCPath is the well-known location gcloud writes user credentials to.
func ADCPath() string {
	if runtime.GOOS == "windows" {
		return filepath.Join(os.Getenv("APPDATA"), "gcloud", "application_default_credentials.json")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".config", "gcloud", "application_default_credentials.json")
}

// ADCAvailable reports whether Application Default Credentials exist, so the
// CLI can suggest them in help text.
func ADCAvailable() bool { return hasADC() }

// GcloudLoginHint is the one-liner that grants gcloud's ADC the calendar scope.
const GcloudLoginHint = "gcloud auth application-default login " +
	"--scopes=https://www.googleapis.com/auth/calendar.readonly,openid,https://www.googleapis.com/auth/userinfo.email"

// scopeHint recognises the "insufficient scope" failure that occurs when ADC
// exist but were minted without calendar access.
func scopeHint(err error) error {
	if err == nil {
		return nil
	}
	msg := err.Error()
	for _, marker := range []string{"insufficient", "ACCESS_TOKEN_SCOPE_INSUFFICIENT", "insufficientPermissions"} {
		if strings.Contains(msg, marker) {
			return fmt.Errorf("%w\n\nYour Google login does not include calendar access. Re-run:\n\n  %s",
				err, GcloudLoginHint)
		}
	}
	return err
}
