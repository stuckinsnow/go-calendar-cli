// Command gcal is a pretty terminal UI for Google Calendar.
package main

import (
	"fmt"
	"os"

	"github.com/stuckinsnow/gcalendar-cli/internal/cli"
)

func main() {
	if err := cli.Run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "gcal: "+err.Error())
		os.Exit(1)
	}
}
