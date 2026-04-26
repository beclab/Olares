// Package network hosts `olares-cli settings network`. Mirrors the SPA's
// Settings -> Network page (Reverse Proxy, FRP, External Network, SSL,
// Hosts File). The wire formats here come from two backends:
//
//  1. user-service /api/network.controller.ts → BFL endpoints under
//     /bfl/settings/v1alpha1/...
//     Some handlers (`/api/reverse-proxy`, `/api/external-network`,
//     `/api/launcher-public-domain-access-policy`) forward `data.data`
//     verbatim to the client, leaving the BFL envelope intact.
//     Others (`/api/frp-servers`) wrap with `returnSucceed(...)` —
//     also a BFL-shaped envelope. Either way, the CLI sees a uniform
//     {code, message, data} response and uses doGetEnvelope.
//
//  2. user-service /api/terminusd.controller.ts → /system/* on the
//     olaresd daemon. Same outer BFL envelope (returnSucceed of the
//     daemon's data.data).
//
// We deliberately keep a per-area common.go (matching settings/me,
// settings/users, settings/apps, settings/vpn) instead of pulling a
// shared package — each area's response decoder can drift independently
// without leaking abstractions across areas.
package network

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

type Format string

const (
	FormatTable Format = "table"
	FormatJSON  Format = "json"
)

func parseFormat(s string) (Format, error) {
	v := strings.ToLower(strings.TrimSpace(s))
	switch v {
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

// Doer is the smallest contract verbs need from the underlying HTTP
// client. *whoami.HTTPClient satisfies it; tests can supply a fake.
type Doer interface {
	DoJSON(ctx context.Context, method, path string, body, out interface{}) error
}

type preparedClient struct {
	profile *credential.ResolvedProfile
	doer    Doer
}

func prepare(ctx context.Context, f *cmdutil.Factory) (*preparedClient, error) {
	if f == nil {
		return nil, fmt.Errorf("internal error: settings network not wired with cmdutil.Factory")
	}
	rp, err := f.ResolveProfile(ctx)
	if err != nil {
		return nil, err
	}
	hc, err := f.HTTPClient(ctx)
	if err != nil {
		return nil, err
	}
	return &preparedClient{
		profile: rp,
		doer:    whoami.NewHTTPClient(hc, rp.DesktopURL, rp.OlaresID),
	}, nil
}

// bflEnvelope mirrors what user-service forwards from BFL or wraps
// itself via returnSucceed.
type bflEnvelope struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

// doGetEnvelope GETs a path that returns a BFL-shaped envelope and
// decodes the inner data into `out`. A code of 0 (or 200, which a
// handful of olaresd-backed handlers use) is treated as success.
func doGetEnvelope(ctx context.Context, d Doer, path string, out interface{}) error {
	var env bflEnvelope
	if err := d.DoJSON(ctx, "GET", path, nil, &env); err != nil {
		return err
	}
	switch env.Code {
	case 0, 200:
		// success
	default:
		msg := strings.TrimSpace(env.Message)
		if msg == "" {
			return fmt.Errorf("GET %s: upstream returned code %d", path, env.Code)
		}
		return fmt.Errorf("GET %s: upstream returned code %d: %s", path, env.Code, msg)
	}
	if out == nil || len(env.Data) == 0 {
		return nil
	}
	if err := json.Unmarshal(env.Data, out); err != nil {
		return fmt.Errorf("GET %s: decode data: %w", path, err)
	}
	return nil
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
