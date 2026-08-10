// Package model hosts `olares-cli model` — Router (the Olares AI gateway,
// whose runtime name is still llm-gateway) together with the per-app Model
// Console runtime that serves locally installed models.
//
// Two things set this tree apart from its profile-authenticated siblings.
//
// Addressing. Router is a Market application rather than a system service,
// so it has no fixed subdomain the way files.<terminus> and
// settings.<terminus> do — app-service gives it a per-install host. Every
// verb therefore resolves the entrance at runtime; see discovery.go.
//
// Identity. The active profile's access token, exactly as elsewhere: the
// Olares edge turns it into the X-BFL-USER header that Router's console
// plane trusts. This tree owns no credential of its own, so authentication
// questions belong to the profile verbs, not here. The `sk-*` API keys it
// can issue are a resource to hand to OTHER applications — Router's /v1
// data plane accepts only `Authorization: Bearer sk-*` or a
// platform-injected x-caller-appid, and by design never a console session.
package model

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
	"github.com/beclab/Olares/cli/pkg/credential"
	"github.com/beclab/Olares/cli/pkg/whoami"
)

// minOlaresVersion is the oldest line whose Market carries Router. The
// `llmgatewayv3` listing declares `olares >= 1.12.6-0`; its successor
// `router` raises that to 1.12.7, but gating on the lower bound keeps the
// tree usable on an instance that has not switched listings yet.
const minOlaresVersion = "1.12.6"

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
type preparedClient struct {
	profile *credential.ResolvedProfile
	router  *routerClient
	found   *discoveredRouter
	desktop *whoami.HTTPClient
}

// prepare resolves the profile, checks the version floor, and locates Router.
// The discovery round-trip is paid once per command invocation — a single
// process runs a single verb, so there is nothing to memoize across calls.
func prepare(ctx context.Context, f *cmdutil.Factory) (*preparedClient, error) {
	if f == nil {
		return nil, fmt.Errorf("internal error: model not wired with cmdutil.Factory")
	}
	if err := cmdutil.RequireMinVersion(ctx, f, cmdutil.MinVersionGate{
		Verb:       "model",
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
