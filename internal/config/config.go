// Package config resolves on-disk locations and user preferences.
package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"time"
)

const appName = "gcal"

// File names inside the config directory.
const (
	credentialsFile = "credentials.json"
	tokenFile       = "token.json"
	settingsFile    = "config.json"
)

// Config holds user preferences. Zero values are filled in by Defaults.
type Config struct {
	// AuthMode selects how to authenticate: auto, client or adc.
	AuthMode AuthMode `json:"authMode"`
	// WeekStart is the leftmost weekday of the month grid (0 = Sunday).
	WeekStart time.Weekday `json:"weekStart"`
	// TwentyFourHour selects between 15:04 and 3:04pm time labels.
	TwentyFourHour bool `json:"twentyFourHour"`
	// Calendars optionally restricts which calendar IDs are fetched.
	// An empty list means "every calendar the account can see".
	Calendars []string `json:"calendars,omitempty"`
	// HideDeclined omits events the user has declined.
	HideDeclined bool `json:"hideDeclined"`

	dir string
}

// AuthMode selects the authentication strategy.
type AuthMode string

const (
	// AuthAuto prefers our own OAuth client, then falls back to Application
	// Default Credentials (an existing gcloud login).
	AuthAuto AuthMode = "auto"
	// AuthClient requires credentials.json and a cached token.
	AuthClient AuthMode = "client"
	// AuthADC requires Application Default Credentials.
	AuthADC AuthMode = "adc"
)

// Defaults returns a sensible configuration.
func Defaults() Config {
	return Config{
		AuthMode:       AuthAuto,
		WeekStart:      time.Monday,
		TwentyFourHour: true,
		HideDeclined:   true,
	}
}

// TimeLayout is the Go layout matching the user's clock preference.
func (c Config) TimeLayout() string {
	if c.TwentyFourHour {
		return "15:04"
	}
	return "3:04pm"
}

// Dir is the directory holding credentials, token and settings.
func (c Config) Dir() string { return c.dir }

// CredentialsPath is where the Google OAuth client secret is expected.
func (c Config) CredentialsPath() string { return filepath.Join(c.dir, credentialsFile) }

// TokenPath is where the cached OAuth token is written.
func (c Config) TokenPath() string { return filepath.Join(c.dir, tokenFile) }

// SettingsPath is the settings file backing this config.
func (c Config) SettingsPath() string { return filepath.Join(c.dir, settingsFile) }

// Dir resolves the config directory, honouring GCAL_CONFIG_DIR and XDG.
func Dir() (string, error) {
	if dir := os.Getenv("GCAL_CONFIG_DIR"); dir != "" {
		return dir, nil
	}
	base, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("resolve user config dir: %w", err)
	}
	return filepath.Join(base, appName), nil
}

// Load reads settings from disk, falling back to defaults when absent.
func Load() (Config, error) {
	dir, err := Dir()
	if err != nil {
		return Config{}, err
	}

	cfg := Defaults()
	cfg.dir = dir

	data, err := os.ReadFile(filepath.Join(dir, settingsFile))
	switch {
	case errors.Is(err, fs.ErrNotExist):
		return cfg, nil
	case err != nil:
		return cfg, fmt.Errorf("read settings: %w", err)
	}

	if err := json.Unmarshal(data, &cfg); err != nil {
		return cfg, fmt.Errorf("parse %s: %w", settingsFile, err)
	}
	cfg.dir = dir
	if cfg.AuthMode == "" {
		cfg.AuthMode = AuthAuto
	}
	return cfg, nil
}

// Save persists settings, creating the config directory if needed.
func (c Config) Save() error {
	if err := os.MkdirAll(c.dir, 0o700); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("encode settings: %w", err)
	}
	if err := os.WriteFile(c.SettingsPath(), append(data, '\n'), 0o600); err != nil {
		return fmt.Errorf("write settings: %w", err)
	}
	return nil
}

// HasCredentials reports whether a Google client secret has been installed.
func (c Config) HasCredentials() bool {
	info, err := os.Stat(c.CredentialsPath())
	return err == nil && !info.IsDir() && info.Size() > 0
}
