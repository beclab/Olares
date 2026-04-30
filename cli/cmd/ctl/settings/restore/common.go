// Package restore hosts `olares-cli settings restore`. Mirrors the
// SPA's Settings -> Restore page (restore plans created from a
// snapshot or from a restic-style URL).
//
// Wire format notes:
//   - Plans live alongside backup plans on the BFL backup-server,
//     under `/apis/backup/v1/plans/restore/...`. Same BFL envelope
//     handling as `settings backup`.
//   - The URL pre-flight (`/apis/backup/v1/plans/restore/checkurl`)
//     is a POST that takes a restic-style URL + password and lists
//     candidate snapshots, exposed as `plans check-url`.
package restore

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

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

type Doer interface {
	DoJSON(ctx context.Context, method, path string, body, out interface{}) error
}

type preparedClient struct {
	profile *credential.ResolvedProfile
	doer    Doer
}

func prepare(ctx context.Context, f *cmdutil.Factory) (*preparedClient, error) {
	if f == nil {
		return nil, fmt.Errorf("internal error: settings restore not wired with cmdutil.Factory")
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

type bflEnvelope struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

func doGetEnvelope(ctx context.Context, d Doer, path string, out interface{}) error {
	return doMutateEnvelope(ctx, d, "GET", path, nil, out)
}

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
			return fmt.Errorf("%s %s: upstream returned code %d", method, path, env.Code)
		}
		return fmt.Errorf("%s %s: upstream returned code %d: %s", method, path, env.Code, msg)
	}
	if out == nil || len(env.Data) == 0 {
		return nil
	}
	if err := json.Unmarshal(env.Data, out); err != nil {
		return fmt.Errorf("%s %s: decode data: %w", method, path, err)
	}
	return nil
}

// readPasswordOnce returns the password either from --password (literal),
// from --password-stdin (read once from stdin), or from a TTY prompt
// without echo. Used for the restore check-url and create flows where
// the user has to provide the repo password.
func readPasswordOnce(literal string, fromStdin bool, promptLabel string) (string, error) {
	if literal != "" {
		return literal, nil
	}
	if fromStdin {
		buf, err := io.ReadAll(os.Stdin)
		if err != nil {
			return "", fmt.Errorf("read --password-stdin: %w", err)
		}
		return strings.TrimRight(string(buf), "\n\r"), nil
	}
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		return "", fmt.Errorf("stdin is not a terminal — pass --password or --password-stdin")
	}
	if _, err := fmt.Fprint(os.Stderr, promptLabel); err != nil {
		return "", err
	}
	defer fmt.Fprintln(os.Stderr)
	pw, err := term.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		return "", fmt.Errorf("read password: %w", err)
	}
	return string(pw), nil
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

func fmtUnix(sec int64) string {
	if sec <= 0 {
		return "-"
	}
	return time.Unix(sec, 0).Format(time.RFC3339)
}

// formatProgressBP renders backup-server's progress field as a 0–100
// integer percent. The wire value is *basis points* (0–10000 where
// 10000 = 100.00%) — see backup-server's handler_snapshot.go and the
// SPA's RestoreDetail.vue / RestoreItem.vue, both of which divide by
// 10000 before feeding Quasar's progress bar. Duplicated here from
// settings/backup/common.go to keep area packages independent.
func formatProgressBP(bp int) string {
	switch {
	case bp <= 0:
		return "0%"
	case bp >= 10000:
		return "100%"
	default:
		return fmt.Sprintf("%d%%", bp/100)
	}
}
