// Package download hosts `olares-cli knowledge download` — the download-server
// task centre via Settings edge `https://settings.<terminus>/download/...`.
// Distinct from top-level `download` (installer packages) and `files download`.
package download

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
	"github.com/beclab/Olares/cli/pkg/credential"
	"github.com/beclab/Olares/cli/pkg/whoami"
)

const (
	minOlaresVersion = "1.12.7"
	defaultApp       = "wise"
	// ytdlpQualityValues is the server-accepted --quality enum
	// (download-server services.IsValidYtdlpQuality). Kept here so the
	// --help text stays in sync with what the backend validates.
	ytdlpQualityValues = "best, 2160p, 1080p, 720p, 480p, 360p, audio"
	// defaultDownloadPath mirrors the wise front-end default landing
	// directory (Termipass collect-site downloadFile: opts.path ||
	// 'Downloads/', normalised to drive/Home/Downloads/). The manager
	// itself applies no default (empty task.Path lands at the PVC root),
	// so the CLI seeds this to match what wise users see.
	defaultDownloadPath = "drive/Home/Downloads/"
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

func addAppFlag(cmd *cobra.Command, target *string) {
	cmd.Flags().StringVar(target, "app", defaultApp, "app namespace for the download task (default: wise)")
}

// Doer is the JSON transport used by download verbs (whoami.HTTPClient).
type Doer interface {
	DoJSON(ctx context.Context, method, path string, body, out interface{}) error
}

type preparedClient struct {
	profile *credential.ResolvedProfile
	doer    Doer
}

// edgeBase is SettingsURL + "/download" (user-service DownloadController strip prefix).
func edgeBase(rp *credential.ResolvedProfile) string {
	if rp == nil {
		return ""
	}
	return strings.TrimRight(rp.SettingsURL, "/") + "/download"
}

func prepare(ctx context.Context, f *cmdutil.Factory) (*preparedClient, error) {
	if f == nil {
		return nil, fmt.Errorf("internal error: download not wired with cmdutil.Factory")
	}
	if err := cmdutil.RequireMinVersion(ctx, f, cmdutil.MinVersionGate{
		Verb:       "knowledge download",
		MinVersion: minOlaresVersion,
		Reason:     "settings /download edge + download provider",
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
	return &preparedClient{
		profile: rp,
		doer:    whoami.NewHTTPClient(hc, edgeBase(rp), rp.OlaresID),
	}, nil
}

// dsEnvelope is download-server's response shape: success code 200 (or 0),
// single object in data, list+total (or list+has_more) at the top level.
type dsEnvelope struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
	List    json.RawMessage `json:"list"`
	Total   *int64          `json:"total"`
	HasMore *bool           `json:"has_more"`
	Exist   *bool           `json:"exist"`
}

func doGet(ctx context.Context, d Doer, path string, out interface{}) error {
	return doMutate(ctx, d, "GET", path, nil, out)
}

func doMutate(ctx context.Context, d Doer, method, path string, body, out interface{}) error {
	var env dsEnvelope
	if err := d.DoJSON(ctx, method, path, body, &env); err != nil {
		return err
	}
	switch env.Code {
	case 0, 200:
	default:
		msg := strings.TrimSpace(env.Message)
		if msg == "" {
			return fmt.Errorf("%s %s: code %d", method, path, env.Code)
		}
		return fmt.Errorf("%s %s: code %d: %s", method, path, env.Code, msg)
	}
	if out == nil {
		return nil
	}
	// List-shaped responses: decode into *ListResult (or any type that
	// expects {"list","total","has_more"}).
	if lr, ok := out.(*ListResult); ok {
		if len(env.List) > 0 {
			if err := json.Unmarshal(env.List, &lr.List); err != nil {
				return fmt.Errorf("%s %s: decode list: %w", method, path, err)
			}
		}
		if env.Total != nil {
			lr.Total = *env.Total
		}
		return nil
	}
	if cl, ok := out.(*CookieListResult); ok {
		if len(env.List) > 0 {
			if err := json.Unmarshal(env.List, &cl.List); err != nil {
				return fmt.Errorf("%s %s: decode list: %w", method, path, err)
			}
		}
		if env.Total != nil {
			cl.Total = *env.Total
		}
		return nil
	}
	if fc, ok := out.(*FileCheckResult); ok {
		if env.Exist != nil {
			fc.Exist = *env.Exist
		}
		return nil
	}
	// Sync-shaped responses: top-level {list, has_more}. Same "list" slot
	// as the list endpoint, so decode from env.List (not env.Data) plus the
	// has_more flag; the composite cursor is derived client-side.
	if sr, ok := out.(*SyncResult); ok {
		if len(env.List) > 0 {
			if err := json.Unmarshal(env.List, &sr.Items); err != nil {
				return fmt.Errorf("%s %s: decode list: %w", method, path, err)
			}
		}
		if env.HasMore != nil {
			sr.HasMore = *env.HasMore
		}
		return nil
	}
	if len(env.Data) == 0 || string(env.Data) == "null" {
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

func encodeQuery(values url.Values) string {
	if len(values) == 0 {
		return ""
	}
	return "?" + values.Encode()
}

func displayName(t DownloadTask) string {
	if strings.TrimSpace(t.FileName) != "" {
		return t.FileName
	}
	if strings.TrimSpace(t.URL) != "" {
		return t.URL
	}
	return "-"
}
