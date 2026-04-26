// Package users hosts the `olares-cli settings users ...` subtree.
//
// common.go centralizes the per-area Doer / output plumbing in the same
// shape as cli/cmd/ctl/settings/me/common.go. We deliberately don't share
// the helpers across packages because the per-area transports may need
// per-area types in later phases (e.g. settings/apps will need ListResult
// awareness, settings/backup will need a different base path), and a
// duplicated 100-line common.go per package is a much smaller cost than
// teasing the shared abstraction out before we know what stays common.
//
// All `users` reads we ship here are gated server-side: app-service's
// /app-service/v1/users handler runs as the cluster controller and lists
// every user; user-service's /api/users/v2 wrapper applies a role-based
// filter so non-privileged callers only see themselves. We standardize on
// v2 for `users list` so the CLI degrades gracefully — a normal user gets
// a 1-row table instead of a 403, which matches the SPA's UX.
package users

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

// Format selects how a `users` read verb renders its result.
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

// Doer is the smallest contract the verbs need from the underlying HTTP
// client; *whoami.HTTPClient satisfies it (and has the desktop-ingress
// 401/403 reformatting we want), tests can supply a fake.
type Doer interface {
	DoJSON(ctx context.Context, method, path string, body, out interface{}) error
}

type preparedClient struct {
	profile *credential.ResolvedProfile
	doer    Doer
}

func prepare(ctx context.Context, f *cmdutil.Factory) (*preparedClient, error) {
	if f == nil {
		return nil, fmt.Errorf("internal error: settings users not wired with cmdutil.Factory")
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

// appServiceListResult mirrors framework/app-service/pkg/apiserver/utils.go's
// ListResult. Note this is NOT the BFL envelope shape — it's app-service's
// own:
//
//	{ "code": 200, "data": [...items], "totals": N }
//
// (no "message" field, code is 200 not 0). user-service forwards this body
// verbatim from /app-service/v1/users for /api/users; the SPA's response
// interceptor unwraps `data.data` because it sees `code` in {0,200}. We
// have to unwrap manually here.
type appServiceListResult[T any] struct {
	Code   int    `json:"code"`
	Data   []T    `json:"data"`
	Totals int    `json:"totals"`
	Error  string `json:"error,omitempty"`
}

// decodeListResult is doGetEnvelope's app-service-shaped sibling. Used by
// `users list` (and likely by any future settings verb that hits an
// endpoint user-service forwards from app-service).
func decodeListResult[T any](ctx context.Context, d Doer, path string, out *appServiceListResult[T]) error {
	if err := d.DoJSON(ctx, "GET", path, nil, out); err != nil {
		return err
	}
	switch out.Code {
	case 0, 200:
		return nil
	default:
		msg := strings.TrimSpace(out.Error)
		if msg == "" {
			msg = fmt.Sprintf("server returned code=%d", out.Code)
		}
		return fmt.Errorf("GET %s: %s", path, msg)
	}
}

func printJSON(w io.Writer, v interface{}) error {
	if w == nil {
		w = os.Stdout
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}
