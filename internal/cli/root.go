// Package cli parses arguments and dispatches to the subcommands.
package cli

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/stuckinsnow/gcalendar-cli/internal/config"
)

// Version is overridden at build time with -ldflags "-X ...cli.Version=v1.2.3".
var Version = "dev"

// defaultCommand runs when no subcommand is given.
const defaultCommand = "tui"

// Run parses args and executes the requested subcommand.
func Run(args []string) error {
	var opts options

	fs := flag.NewFlagSet("gcal", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	fs.BoolVar(&opts.demo, "demo", false, "use the built-in demo calendar instead of Google")
	fs.IntVar(&opts.days, "days", 7, "number of days for the agenda command")
	fs.BoolVar(&opts.version, "version", false, "print the version and exit")

	command, err := parse(fs, args)
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			printUsage()
			return nil
		}
		return err
	}

	if opts.version {
		fmt.Println("gcal " + Version)
		return nil
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	return dispatch(ctx, cfg, command, opts)
}

// parse extracts the subcommand, allowing flags on either side of it.
//
// The standard flag package stops at the first positional argument, so
// "gcal today -demo" needs a second pass over the remaining arguments.
func parse(fs *flag.FlagSet, args []string) (string, error) {
	if err := fs.Parse(args); err != nil {
		return "", err
	}

	command := defaultCommand
	if fs.NArg() > 0 {
		command = fs.Arg(0)
		if err := fs.Parse(fs.Args()[1:]); err != nil {
			return "", err
		}
	}
	return command, nil
}

// dispatch maps a command name to its handler.
func dispatch(ctx context.Context, cfg config.Config, command string, opts options) error {
	switch command {
	case defaultCommand, "calendar":
		return runTUI(ctx, cfg, opts)
	case "auth", "login":
		return runAuth(ctx, cfg)
	case "today":
		opts.days = 1
		return runAgenda(ctx, cfg, opts)
	case "agenda", "week":
		return runAgenda(ctx, cfg, opts)
	case "config", "where":
		return runConfigInfo(cfg)
	case "help":
		printUsage()
		return nil
	default:
		printUsage()
		return fmt.Errorf("unknown command %q", command)
	}
}

func printUsage() {
	fmt.Print(`gcal — a pretty Google Calendar in your terminal

Usage:
  gcal [flags]            open the interactive calendar
  gcal today              print today's agenda
  gcal agenda [-days N]   print the next N days (default 7)
  gcal auth               authorise access to Google Calendar
  gcal config             show config paths and setup status
  gcal help               show this help

Flags:
  -demo        use the built-in demo calendar (no Google account needed)
  -days N      day count for the agenda command
  -version     print the version

Keys inside the interactive view:
  ←/h →/l      previous / next day        [ / ]   previous / next month
  ↑/k ↓/j      previous / next week       t       jump to today
  tab          switch pane                enter   event details
  r            refresh                    ?       toggle help
  q            quit
`)
}
