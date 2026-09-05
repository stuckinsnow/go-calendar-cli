package calendar

import (
	"context"
	"time"
)

// Provider fetches events in the half-open interval [from, to).
//
// Implementations live under internal/provider. The UI only ever talks to this
// interface, which is what makes the offline demo calendar possible.
type Provider interface {
	// Name identifies the backing source, shown in the status bar.
	Name() string
	// Events returns every event overlapping [from, to).
	Events(ctx context.Context, from, to time.Time) ([]Event, error)
}

// ProviderFunc adapts a plain function to the Provider interface.
type ProviderFunc struct {
	Label string
	Fetch func(ctx context.Context, from, to time.Time) ([]Event, error)
}

func (p ProviderFunc) Name() string { return p.Label }

func (p ProviderFunc) Events(ctx context.Context, from, to time.Time) ([]Event, error) {
	return p.Fetch(ctx, from, to)
}
