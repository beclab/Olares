// Package vpn hosts `olares-cli settings vpn`. Mirrors the SPA's
// Settings -> VPN page. Three flavors of endpoints ride here:
//
//  1. Headscale proxy at /headscale/...                  (raw upstream JSON;
//                                                        no BFL envelope)
//  2. Network ACL / public-domain-policy at /api/...     (user-service
//                                                        forwards data.data
//                                                        from BFL — body
//                                                        already unwrapped)
//  3. (future) Subroutes / SSH at /api/... (Phase 3)
//
// common.go centralizes the per-area Doer + output plumbing in the same
// shape as the other settings subpackages. We deliberately don't reach
// into a shared package because each area's wire envelope differs (BFL
// envelope vs raw vs app-service ListResult), and per-area helpers stay
// honest about which decoder maps to which path.
package vpn

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/term"

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
		doer:    whoami.NewHTTPClient(hc, rp.DesktopURL, rp.OlaresID),
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

// confirmDestructive guards destructive verbs (device delete) behind a
// y/N prompt unless --yes was passed. Non-TTY stdin without --yes is a
// hard error rather than an implicit yes — we'd rather break a script
// than silently delete something the operator didn't review.
//
// The prompt body should describe what's about to happen ("delete
// device X"). The success criterion is a literal "y" or "yes" answer
// (case-insensitive); anything else aborts with a clear error.
func confirmDestructive(prompt io.Writer, in io.Reader, message string, assumeYes bool) error {
	if assumeYes {
		return nil
	}
	if f, ok := in.(*os.File); ok {
		if !term.IsTerminal(int(f.Fd())) {
			return fmt.Errorf("stdin is not a terminal — pass --yes to confirm: %s", message)
		}
	}
	if _, err := fmt.Fprintf(prompt, "%s [y/N]: ", message); err != nil {
		return err
	}
	rd := bufio.NewReader(in)
	line, err := rd.ReadString('\n')
	if err != nil && err != io.EOF {
		return fmt.Errorf("read confirmation: %w", err)
	}
	switch strings.ToLower(strings.TrimSpace(line)) {
	case "y", "yes":
		return nil
	default:
		return fmt.Errorf("aborted by user")
	}
}
