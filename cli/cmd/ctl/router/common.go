// Package router hosts `olares-cli router` — Router (the Olares AI gateway,
// whose runtime name is still llm-gateway) together with the per-app Model
// Console runtime that serves locally installed models.
//
// Two things set this tree apart from its profile-authenticated siblings.
//
// Addressing. Router is a Market application rather than a system service,
// so it has no fixed subdomain the way files.<terminus> and
// settings.<terminus> do — app-service gives it a per-install host. Every
// verb therefore resolves the entrance at runtime; see discovery.go. The
// verbs that reach into a model application address a second host the same
// way, because it serves its own Model Console on its own entrance.
//
// Identity. Configuration runs on the active profile's access token, exactly
// as elsewhere: the Olares edge turns it into the X-BFL-USER header that
// Router's console plane trusts, so authentication questions there belong to
// the profile verbs rather than here.
//
// Calling a model is the exception, and the reason this tree keeps a secret of
// its own. Router's /v1 data plane accepts an `sk-*` key or a
// platform-injected caller identity and by design never a console session, so
// `router call` presents whichever of the two it can get and mints a key when
// it can get neither. See dataplane.go for the order and why. Keys issued with
// `router key` remain what they were — a credential to hand to other software.
package router

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
	"github.com/beclab/Olares/cli/pkg/credential"
	"github.com/beclab/Olares/cli/pkg/whoami"
)

// minOlaresVersion is the oldest line this tree speaks to: the `router`
// application declares `olares >= 1.12.7-0`. The Model Console verbs share the
// floor, because a Model Console with no Router in front of it is not a shape
// this tree addresses.
const minOlaresVersion = "1.12.7"

type Format string

const (
	FormatTable Format = "table"
	FormatJSON  Format = "json"
)

func parseFormat(s string) (Format, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "", string(FormatTable):
		return FormatTable, nil
	case string(FormatJSON):
		return FormatJSON, nil
	default:
		return "", fmt.Errorf("unsupported --output %q (allowed: table, json)", s)
	}
}

func addOutputFlag(cmd *cobra.Command, target *string) {
	cmd.Flags().StringVarP(target, "output", "o", "table", "output format: table, json")
}

// preparedClient is what every verb needs: the resolved profile (for Desktop
// addressing and diagnostics), where Router turned out to live, a client
// pointed at it, and the Desktop client discovery used — kept because Olares
// itself, not Router, is the authority on which applications are installed.
// The session itself is kept alongside them: it carries the profile's token
// and its refresh, and a model application's Model Console is a third host
// reached over the same one.
type preparedClient struct {
	profile *credential.ResolvedProfile
	router  *routerClient
	found   *discoveredRouter
	desktop *whoami.HTTPClient
	hc      *http.Client

	// collections holds what has already been read this invocation; see
	// collection in resolve.go for what invalidates it.
	collections map[string]memoizedCollection
}

// prepare resolves the profile, checks the version floor, and locates Router.
// The discovery round-trip is paid once per command invocation — a single
// process runs a single verb, so there is nothing to memoize across calls.
func prepare(ctx context.Context, f *cmdutil.Factory) (*preparedClient, error) {
	if f == nil {
		return nil, fmt.Errorf("internal error: router not wired with cmdutil.Factory")
	}
	if err := cmdutil.RequireMinVersion(ctx, f, cmdutil.MinVersionGate{
		Verb:       "router",
		MinVersion: minOlaresVersion,
		Reason:     "Router, the AI gateway Market app",
	}); err != nil {
		return nil, err
	}
	rp, err := f.ResolveProfile(ctx)
	if err != nil {
		return nil, err
	}
	hc, err := f.HTTPClient(ctx)
	if err != nil {
		return nil, err
	}
	desktop := whoami.NewHTTPClient(hc, rp.DesktopURL, rp.OlaresID)
	found, err := discoverRouter(ctx, desktop, rp)
	if err != nil {
		return nil, err
	}
	return &preparedClient{
		profile: rp,
		router:  newRouterClient(hc, found.BaseURL, rp.OlaresID),
		found:   found,
		desktop: desktop,
		hc:      hc,
	}, nil
}

func printJSON(w io.Writer, v interface{}) error {
	if w == nil {
		w = os.Stdout
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

func nonEmpty(s string) string {
	if s == "" {
		return "-"
	}
	return s
}

func strDeref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// i18nText picks one string out of a locale map.
//
// Three unrelated wire shapes arrive as one of these: a vendor label out of the
// provider catalog and a model card's label, both keyed the Dify way (`en_US`),
// and a Market application's title and description, keyed the way the
// application's own i18n bundle is keyed (`en-US`, `zh-CN`). Underscore and
// hyphen name the same locale, so both spellings are tried.
//
// English first, because these are read on a terminal by whoever runs the
// command and the CLI has no locale of its own to consult. The last resort is
// the alphabetically first non-empty entry rather than whatever the map happens
// to yield: Go randomises that, and a column whose text changes between two runs
// of the same command reads as a machine that changed.
func i18nText(m map[string]string) string {
	if len(m) == 0 {
		return ""
	}
	for _, k := range []string{"en_US", "en-US", "en", "zh_Hans", "zh-CN", "zh_CN", "zh"} {
		if v := strings.TrimSpace(m[k]); v != "" {
			return v
		}
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		if v := strings.TrimSpace(m[k]); v != "" {
			return v
		}
	}
	return ""
}

func boolStr(b bool) string {
	if b {
		return "yes"
	}
	return "no"
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "...(truncated)"
}

// humanBytes is for confirming a file was written. Powers of two, because that
// is what a file manager will say about the same file a moment later.
func humanBytes(n int64) string {
	const unit = 1024
	if n < unit {
		return fmt.Sprintf("%d B", n)
	}
	value := float64(n)
	for _, suffix := range []string{"KiB", "MiB", "GiB", "TiB"} {
		value /= unit
		if value < unit {
			return fmt.Sprintf("%.1f %s", value, suffix)
		}
	}
	return fmt.Sprintf("%.1f PiB", value/unit)
}

// clip shortens a table cell. Unlike truncate it says so with one character:
// inside a column, "...(truncated)" is longer than most of the values around it
// and turns a readable row into a wall of the same word.
func clip(s string, n int) string {
	if len([]rune(s)) <= n {
		return s
	}
	r := []rune(s)
	if n <= 1 {
		return "…"
	}
	return string(r[:n-1]) + "…"
}
