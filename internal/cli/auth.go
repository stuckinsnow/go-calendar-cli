package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/stuckinsnow/gcalendar-cli/internal/config"
	"github.com/stuckinsnow/gcalendar-cli/internal/provider/google"
	"github.com/stuckinsnow/gcalendar-cli/internal/ui/theme"
)

// runAuth signs in and reports the calendars found.
func runAuth(ctx context.Context, cfg config.Config) error {
	t := theme.New()

	if err := os.MkdirAll(cfg.Dir(), 0o700); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}
	if !cfg.HasCredentials() && !google.ADCAvailable() {
		return fmt.Errorf("%w\n\n%s", google.ErrNoCredentials, setupHint(cfg))
	}

	client, err := google.New(ctx, cfg, true)
	if err != nil {
		return err
	}

	fmt.Println(t.Title.Render("Signed in") + " " + t.Subtitle.Render(client.Name()))
	fmt.Println(t.Subtitle.Render("via " + client.AuthMethod()))
	for _, name := range client.Calendars() {
		fmt.Println("  " + t.Key.Render("•") + " " + t.Value.Render(name))
	}
	return nil
}

// runConfigInfo prints where things live and which login methods are available.
func runConfigInfo(cfg config.Config) error {
	t := theme.New()

	rows := [][2]string{
		{"config dir", cfg.Dir()},
		{"settings", cfg.SettingsPath()},
		{"auth mode", string(cfg.AuthMode)},
		{"oauth client", statusOf(cfg.CredentialsPath())},
		{"cached token", statusOf(cfg.TokenPath())},
		{"gcloud adc", statusOf(google.ADCPath())},
		{"week starts", cfg.WeekStart.String()},
		{"clock", clockLabel(cfg)},
	}

	fmt.Println(t.Title.Render("gcal configuration"))
	for _, row := range rows {
		fmt.Printf("  %-14s %s\n", t.Key.Render(row[0]), t.Value.Render(row[1]))
	}

	if !cfg.HasCredentials() && !google.ADCAvailable() {
		fmt.Println("\n" + t.Subtitle.Render(loginReminder(cfg)))
	}
	return nil
}

// loginReminder suggests the next step for the configured auth mode.
func loginReminder(cfg config.Config) string {
	if cfg.AuthMode == config.AuthADC {
		return "No Google login yet. Run:\n  " + google.GcloudLoginHint
	}
	return "No Google login yet. Save an OAuth client (Desktop app) as\n  " +
		cfg.CredentialsPath() + "\nthen run: gcal auth"
}

// statusOf annotates a path with whether it exists.
func statusOf(path string) string {
	if path == "" {
		return "(unknown location)"
	}
	if _, err := os.Stat(path); err != nil {
		return path + "  (missing)"
	}
	return path + "  (present)"
}

func clockLabel(cfg config.Config) string {
	if cfg.TwentyFourHour {
		return "24-hour"
	}
	return "12-hour"
}
