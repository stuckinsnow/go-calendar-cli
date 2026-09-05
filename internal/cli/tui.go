package cli

import (
	"context"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/stuckinsnow/gcalendar-cli/internal/config"
	"github.com/stuckinsnow/gcalendar-cli/internal/ui"
)

// runTUI opens the interactive calendar.
func runTUI(ctx context.Context, cfg config.Config, opts options) error {
	provider, err := resolveProvider(ctx, cfg, opts)
	if err != nil {
		return err
	}

	// Mouse reporting is deliberately not enabled: with it on, the terminal
	// hands drag events to us and the user can no longer select and copy text.
	// Scrolling is keyboard-driven, and `y` copies the selected event.
	program := tea.NewProgram(
		ui.New(cfg, provider),
		tea.WithAltScreen(),
		tea.WithContext(ctx),
	)
	if _, err := program.Run(); err != nil {
		return fmt.Errorf("run interface: %w", err)
	}
	return nil
}
