// Package vpn hosts `olares-cli settings vpn`. Mirrors the SPA's
// Settings -> VPN page. Three flavors of endpoints ride here:
//
//  1. Headscale proxy at /headscale/...                  (raw upstream JSON;
//                                                        no BFL envelope)
//  2. Network ACL / public-domain-policy at /api/...     (user-service
//                                                        forwards data.data
//                                                        from BFL — body
//                                                        already unwrapped)
//  3. Subroutes / SSH toggles at /api/acl/...            (no envelope on
//                                                        the read; opaque
//                                                        body on POST)
//  4. Per-app ACL at /api/acl/app/status                 (BFL envelope on
//                                                        both ends; GET
//                                                        treats code!=0
//                                                        as "no ACL
//                                                        configured" not
//                                                        a hard error)
//
// common.go centralizes the per-area Doer + output plumbing in the same
// shape as the other settings subpackages. We deliberately don't reach
// into a shared package because each area's wire envelope differs (BFL
// envelope vs raw vs app-service ListResult), and per-area helpers stay
// honest about which decoder maps to which path.
package vpn

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
// client. *whoami.HTTPClient satisfies it.
type Doer interface {
	DoJSON(ctx context.Context, method, path string, body, out interface{}) error
}

type preparedClient struct {
	profile *credential.ResolvedProfile
	doer    Doer
}

func prepare(ctx context.Context, f *cmdutil.Factory) (*preparedClient, error) {
	if f == nil {
		return nil, fmt.Errorf("internal error: settings vpn not wired with cmdutil.Factory")
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
		doer:    whoami.NewHTTPClient(hc, rp.SettingsURL, rp.OlaresID),
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

// bflEnvelope is the BFL response wrapper that comes back on the per-app
// ACL endpoints. Most other vpn paths either hit Headscale directly (raw
// JSON, no envelope) or a user-service route that already unwraps the
// envelope before responding (public-domain-policy, ssh status). The
// per-app ACL editor is the odd one out — both the GET and the POST
// round-trip the envelope verbatim.
type bflEnvelope struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

// doMutateEnvelope is the POST/PUT/DELETE counterpart of plain DoJSON
// for routes that surface a BFL envelope on the response. Treats
// `code: 0` and `code: 200` as success; anything else surfaces the
// upstream message verbatim. Used today only by the per-app ACL editor;
// the existing ssh/subroutes/public-domain-policy writes pass nil out
// because user-service unwraps the envelope server-side for those
// paths.
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

func joinNonEmpty(ss []string, sep string) string {
	if len(ss) == 0 {
		return "-"
	}
	out := ""
	for i, s := range ss {
		if i > 0 {
			out += sep
		}
		out += s
	}
	return out
}

