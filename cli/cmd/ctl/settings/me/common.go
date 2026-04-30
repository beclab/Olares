// Package me hosts the `olares-cli settings me ...` self-service subtree.
//
// common.go centralizes:
//   - The BFL response envelope helper (every endpoint we hit here returns
//     {code, message, data}).
//   - The minimal output-format flag plumbing shared by the read verbs
//     (version / check-update / sso list). We can't reuse
//     settings.SettingsOptions directly because that type is in package
//     settings and pulling it in here creates an import cycle (settings
//     imports me, so me cannot import settings back).
//
// Transport reuses cli/pkg/whoami.HTTPClient — a misnomer for historical
// reasons (it was first built for the whoami endpoint), but in practice
// it's already a generic Doer pointed at the desktop ingress with the
// X-Authorization header injected and 401/403 reformatted into the
// canonical "run profile login" CTA. Sharing it here keeps the auth-error
// story consistent between profile whoami and the four me reads.
//
// All `me` verbs are callable by any authenticated user — owner, admin, or
// normal. None of them call PreflightRole because empty / unknown role in
// the cache is treated as "let the server decide" and these endpoints
// don't actually role-gate.
package me

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

// Format selects how a `me` read verb renders its result.
type Format string

const (
	FormatTable Format = "table"
	FormatJSON  Format = "json"
)

// parseFormat normalizes -o / --output into a Format. Empty string maps to
// table to match the CLI-wide default.
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

// addOutputFlag wires a single string variable as the standard output
// selector. Read verbs only need this one — adding --quiet / --no-headers
// without honoring them would be a footgun.
func addOutputFlag(cmd *cobra.Command, target *string) {
	cmd.Flags().StringVarP(target, "output", "o", "table", "output format: table, json")
}

// Doer is the smallest contract the verbs need from the underlying HTTP
// client. *whoami.HTTPClient satisfies this; tests can supply a fake.
type Doer interface {
	DoJSON(ctx context.Context, method, path string, body, out interface{}) error
}

// preparedClient bundles the resolved profile with the Doer pointed at
// its desktop ingress. Verbs call prepare() at the top of their RunE so
// they can keep their bodies focused on path/result handling.
type preparedClient struct {
	profile *credential.ResolvedProfile
	doer    Doer
}

// prepare resolves the active profile and constructs a desktop-ingress
// Doer. The factory's HTTPClient already injects X-Authorization via its
// refreshingTransport (and auto-rotates expired access_tokens), so we just
// hand it to whoami.NewHTTPClient — which uses the request transport's
// existing header injection rather than setting its own token.
func prepare(ctx context.Context, f *cmdutil.Factory) (*preparedClient, error) {
	if f == nil {
		return nil, fmt.Errorf("internal error: settings me not wired with cmdutil.Factory")
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

// bflEnvelope is the universal user-service / BFL response wrapper. Every
// /api/* and /bfl/backend/* endpoint we hit in `me` returns this shape:
//
//	{ "code": 0, "message": "ok", "data": <typed payload> }
//
// We unmarshal data lazily via json.RawMessage so each verb can decode
// into its own typed struct without re-reading the response.
type bflEnvelope struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

// doGetEnvelope sends a GET, validates the BFL envelope, and decodes
// data into out. Verbs that only care about the unwrapped payload should
// call this; verbs that need code/message visibility (none yet) can call
// d.DoJSON directly with their own envelope type.
//
// Errors fall into three buckets:
//   - HTTP / network → wrapped by the underlying Doer (already
//     CTA-formatted for 401/403).
//   - Envelope code != 0 → wrapped here so the user sees the server's
//     own message rather than a generic "decode failure".
//   - JSON decode of data → wrapped here.
func doGetEnvelope(ctx context.Context, d Doer, path string, out interface{}) error {
	return doMutateEnvelope(ctx, d, "GET", path, nil, out)
}

// doMutateEnvelope is the POST/PUT/DELETE counterpart of doGetEnvelope.
// user-service routes either return BFL `{code: 0}` (the most common
// shape — returnSucceed wraps everything) or, for endpoints that proxy
// through additional layers, `{code: 200}`. Both are treated as success
// here.
func doMutateEnvelope(ctx context.Context, d Doer, method, path string, body, out interface{}) error {
	var env bflEnvelope
	if err := d.DoJSON(ctx, method, path, body, &env); err != nil {
		return err
	}
	switch env.Code {
	case 0, 200:
	default:
		msg := strings.TrimSpace(env.Message)
		if msg == "" {
			msg = fmt.Sprintf("server returned code=%d", env.Code)
		}
		return fmt.Errorf("%s %s: %s", method, path, msg)
	}
	if out == nil || len(env.Data) == 0 {
		return nil
	}
	if err := json.Unmarshal(env.Data, out); err != nil {
		return fmt.Errorf("%s %s: decode data: %w", method, path, err)
	}
	return nil
}

// printJSON pretty-prints v to stdout — used by every `me` verb when the
// caller passes -o json. Indentation matches the rest of the CLI (2
// spaces).
func printJSON(w io.Writer, v interface{}) error {
	if w == nil {
		w = os.Stdout
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

// printKV writes a "Key: Value" table for the simple flat-record verbs
// (version / check-update). Each row aligns the key column to keyWidth
// so the output looks tabular without dragging in tabwriter for two
// values.
func printKV(w io.Writer, rows [][2]string, keyWidth int) error {
	for _, r := range rows {
		if _, err := fmt.Fprintf(w, "%-*s  %s\n", keyWidth, r[0]+":", r[1]); err != nil {
			return err
		}
	}
	return nil
}
