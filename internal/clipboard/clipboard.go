// Package clipboard copies text to the system clipboard, falling back to the
// OSC 52 terminal escape when no helper binary is available (which also makes
// copying work over SSH).
package clipboard

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
)

// helpers are tried in order; the first one present on PATH wins.
var helpers = []struct {
	name string
	args []string
}{
	{"wl-copy", nil},
	{"xclip", []string{"-selection", "clipboard"}},
	{"xsel", []string{"--clipboard", "--input"}},
	{"pbcopy", nil},
}

// Copy places text on the clipboard and reports the mechanism used.
func Copy(text string) (string, error) {
	for _, h := range helpers {
		path, err := exec.LookPath(h.name)
		if err != nil {
			continue
		}
		cmd := exec.Command(path, h.args...)
		cmd.Stdin = bytes.NewReader([]byte(text))
		if err := cmd.Run(); err != nil {
			return "", fmt.Errorf("%s: %w", h.name, err)
		}
		return h.name, nil
	}

	if err := osc52(text); err != nil {
		return "", err
	}
	return "terminal", nil
}

// osc52 asks the terminal itself to set the clipboard.
func osc52(text string) error {
	payload := base64.StdEncoding.EncodeToString([]byte(text))
	if _, err := fmt.Fprintf(os.Stdout, "\x1b]52;c;%s\x07", payload); err != nil {
		return fmt.Errorf("write clipboard escape: %w", err)
	}
	return nil
}
