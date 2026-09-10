// Package apps hosts the `olares-cli settings apps ...` subtree.
//
// common.go centralizes the per-area Doer + output plumbing in the same
// shape as cli/cmd/ctl/settings/me/common.go and
// cli/cmd/ctl/settings/users/common.go. These tiny helpers stay per-area
// rather than being hoisted to a shared package because each area
// accumulates its own response-shape quirks; the cost of a small
// per-area common.go is much smaller than churning an over-abstracted
// shared helper.
package apps

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/bflenvelope"
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
		return nil, fmt.Errorf("internal error: settings apps not wired with cmdutil.Factory")
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

func doGetEnvelope(ctx context.Context, d Doer, path string, out interface{}) error {
	return doMutateEnvelope(ctx, d, "GET", path, nil, out)
}

// doMutateEnvelope is the POST/PUT/DELETE counterpart of doGetEnvelope.
func doMutateEnvelope(ctx context.Context, d Doer, method, path string, body, out interface{}) error {
	var env bflenvelope.Envelope
	if err := d.DoJSON(ctx, method, path, body, &env); err != nil {
		return err
	}
	return bflenvelope.Data(method, path, env, out)
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
