package cli

// options holds the flags shared by every subcommand.
type options struct {
	// demo selects the built-in offline calendar.
	demo bool
	// days is the window for the agenda command.
	days int
	// version prints the build version and exits.
	version bool
}
