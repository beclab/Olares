// Package backup hosts `olares-cli settings backup`. Mirrors the SPA's
// Settings -> Backup page (Backup plans + per-plan snapshots + the
// repository password).
//
// Wire format notes:
//   - Plans + snapshots ride a different ingress prefix than the
//     other settings areas: `/apis/backup/v1/plans/backup/...` (BFL's
//     backup-server). The Olares desktop ingress forwards these to
//     backup-server intact, and the responses are BFL-shaped envelopes
//     (`{code, message, data}`). The SPA's global axios interceptor
//     already unwraps `data.data`, which is why upstream code reads
//     `{ backups: [...] }` directly. The CLI does the same via
//     `doGetEnvelope`.
//   - The repository password endpoint goes through user-service at
//     `/api/backup/password/:name` (not used in Phase 1 — Phase 6
//     adds password get/set).
//   - Phase 6 lands the write verbs (create / update / delete plan,
//     create / cancel snapshot, password set). Phase 1 ships the two
//     read-only verbs that exercise the BFL prefix end-to-end and
//     populate cli/skills/olares-settings/SKILL.md.
package backup

import (
	"bufio"
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
		return nil, fmt.Errorf("internal error: settings backup not wired with cmdutil.Factory")
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

// bflEnvelope mirrors the {code, message, data} BFL backup-server
// returns under /apis/backup/v1/... Some handlers use code 0 for
// success and others (occasionally) code 200; we tolerate both, as
// settings/network does.
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

// fmtUnix renders an epoch-seconds timestamp in local-time RFC3339
// (or "-" if the value is zero / negative). Backup-server timestamps
// are seconds, not milliseconds.
func fmtUnix(sec int64) string {
	if sec <= 0 {
		return "-"
	}
	return time.Unix(sec, 0).Format(time.RFC3339)
}

// confirmDestructive guards plan / snapshot deletion behind a y/N
// prompt unless --yes was passed. Mirrors the shape used in
// settings/vpn/common.go: non-TTY stdin without --yes is a hard error
// rather than an implicit yes.
func confirmDestructive(prompt io.Writer, in io.Reader, message string) error {
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
