// Package google implements the calendar.Provider interface on top of the
// Google Calendar API.
package google

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"golang.org/x/oauth2"
	googleauth "golang.org/x/oauth2/google"

	"github.com/stuckinsnow/gcalendar-cli/internal/config"
)

// authTimeout bounds how long the browser consent step may take.
const authTimeout = 3 * time.Minute

// oauthClient returns a client authorised with our own OAuth client ID, running
// the browser consent flow only when no usable cached token exists.
func oauthClient(ctx context.Context, cfg config.Config, interactive bool) (*http.Client, error) {
	oauthCfg, err := loadOAuthConfig(cfg)
	if err != nil {
		return nil, err
	}

	token, err := readToken(cfg.TokenPath())
	switch {
	case err != nil && !errors.Is(err, fs.ErrNotExist):
		return nil, err
	case errors.Is(err, fs.ErrNotExist):
		if !interactive {
			return nil, fmt.Errorf("not authorised yet: run `gcal auth`")
		}
		if token, err = authorize(ctx, oauthCfg); err != nil {
			return nil, err
		}
		if err := writeToken(cfg.TokenPath(), token); err != nil {
			return nil, err
		}
	}

	src := &persistingSource{
		src:  oauthCfg.TokenSource(ctx, token),
		path: cfg.TokenPath(),
		last: token,
	}
	return oauth2.NewClient(ctx, src), nil
}

// loadOAuthConfig reads the client secret written by the Google Cloud console.
func loadOAuthConfig(cfg config.Config) (*oauth2.Config, error) {
	data, err := os.ReadFile(cfg.CredentialsPath())
	if errors.Is(err, fs.ErrNotExist) {
		return nil, fmt.Errorf("%w at %s", ErrNoCredentials, cfg.CredentialsPath())
	}
	if err != nil {
		return nil, fmt.Errorf("read credentials: %w", err)
	}

	oauthCfg, err := googleauth.ConfigFromJSON(data, Scope)
	if err != nil {
		return nil, fmt.Errorf("parse credentials: %w", err)
	}
	return oauthCfg, nil
}

// authorize runs the loopback OAuth flow: it serves a one-shot callback on
// localhost, opens the consent page, then exchanges the returned code.
func authorize(ctx context.Context, oauthCfg *oauth2.Config) (*oauth2.Token, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, fmt.Errorf("start callback listener: %w", err)
	}
	defer listener.Close()

	// Desktop OAuth clients may use any loopback port.
	cfgCopy := *oauthCfg
	cfgCopy.RedirectURL = fmt.Sprintf("http://127.0.0.1:%d/callback", listener.Addr().(*net.TCPAddr).Port)

	state, err := randomState()
	if err != nil {
		return nil, err
	}

	type result struct {
		code string
		err  error
	}
	results := make(chan result, 1)

	srv := &http.Server{
		ReadHeaderTimeout: 10 * time.Second,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			q := r.URL.Query()
			if got := q.Get("state"); got != state {
				http.Error(w, "state mismatch", http.StatusBadRequest)
				results <- result{err: errors.New("oauth state mismatch")}
				return
			}
			if errParam := q.Get("error"); errParam != "" {
				http.Error(w, errParam, http.StatusBadRequest)
				results <- result{err: fmt.Errorf("authorisation denied: %s", errParam)}
				return
			}
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			_, _ = w.Write([]byte(successPage))
			results <- result{code: q.Get("code")}
		}),
	}
	go func() { _ = srv.Serve(listener) }()
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutdownCtx)
	}()

	authURL := cfgCopy.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.ApprovalForce)
	fmt.Fprintln(os.Stderr, "Opening your browser to authorise access to Google Calendar.")
	fmt.Fprintln(os.Stderr, "If it does not open, visit:\n\n  "+authURL+"\n")
	openBrowser(authURL)

	ctx, cancel := context.WithTimeout(ctx, authTimeout)
	defer cancel()

	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("waiting for authorisation: %w", ctx.Err())
	case res := <-results:
		if res.err != nil {
			return nil, res.err
		}
		token, err := cfgCopy.Exchange(ctx, res.code)
		if err != nil {
			return nil, fmt.Errorf("exchange authorisation code: %w", err)
		}
		return token, nil
	}
}

// randomState produces an unguessable CSRF state parameter.
func randomState() (string, error) {
	buf := make([]byte, 24)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("generate oauth state: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

// openBrowser makes a best-effort attempt to launch the consent page.
func openBrowser(url string) {
	var cmd string
	switch runtime.GOOS {
	case "darwin":
		cmd = "open"
	case "windows":
		cmd = "rundll32"
		_ = exec.Command(cmd, "url.dll,FileProtocolHandler", url).Start()
		return
	default:
		cmd = "xdg-open"
	}
	_ = exec.Command(cmd, url).Start()
}

// readToken loads a cached OAuth token.
func readToken(path string) (*oauth2.Token, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var token oauth2.Token
	if err := json.Unmarshal(data, &token); err != nil {
		return nil, fmt.Errorf("parse cached token: %w", err)
	}
	return &token, nil
}

// writeToken persists a token with owner-only permissions, creating the config
// directory if this is the first sign-in.
func writeToken(path string, token *oauth2.Token) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}
	data, err := json.Marshal(token)
	if err != nil {
		return fmt.Errorf("encode token: %w", err)
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("write token: %w", err)
	}
	return nil
}

// persistingSource writes refreshed tokens back to disk so the consent flow is
// only ever needed once.
type persistingSource struct {
	src  oauth2.TokenSource
	path string
	last *oauth2.Token
}

func (s *persistingSource) Token() (*oauth2.Token, error) {
	token, err := s.src.Token()
	if err != nil {
		return nil, err
	}
	if s.last == nil || token.AccessToken != s.last.AccessToken {
		s.last = token
		_ = writeToken(s.path, token)
	}
	return token, nil
}

const successPage = `<!doctype html>
<html lang="en"><head><meta charset="utf-8"><title>gcal authorised</title>
<style>
 body{font-family:ui-monospace,SFMono-Regular,Menlo,monospace;background:#1a1a26;color:#e6e6f0;
      display:grid;place-items:center;height:100vh;margin:0}
 .card{border:1px solid #7d56f4;border-radius:10px;padding:2rem 2.5rem;text-align:center}
 h1{color:#ff5faf;font-size:1.1rem;margin:0 0 .5rem}
 p{margin:0;opacity:.75;font-size:.9rem}
</style></head>
<body><main class="card"><h1>Authorised</h1><p>You can close this tab and return to the terminal.</p></main></body></html>`
