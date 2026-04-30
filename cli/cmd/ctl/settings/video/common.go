// Package video hosts `olares-cli settings video`. Mirrors the SPA's
// Settings -> Video page (a single Jellyfin-style encoding config blob).
// Backed by user-service's files.controller.ts /api/files/video/config,
// which proxies the files-service /system/configuration/encoding endpoint.
// user-service wraps with returnSucceed, so the CLI sees a uniform BFL
// envelope.
package video

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

type Doer interface {
	DoJSON(ctx context.Context, method, path string, body, out interface{}) error
}

type preparedClient struct {
	profile *credential.ResolvedProfile
	doer    Doer
}

func prepare(ctx context.Context, f *cmdutil.Factory) (*preparedClient, error) {
	if f == nil {
		return nil, fmt.Errorf("internal error: settings video not wired with cmdutil.Factory")
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

type bflEnvelope struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

// doGetEnvelopeRaw decodes the inner data into a json.RawMessage so the
// caller can either pretty-print it as-is or further process. The video
// config struct is large and provider-versioned, so we deliberately
// don't model it field-by-field at the CLI layer.
func doGetEnvelopeRaw(ctx context.Context, d Doer, path string) (json.RawMessage, error) {
	var env bflEnvelope
	if err := d.DoJSON(ctx, "GET", path, nil, &env); err != nil {
		return nil, err
	}
	switch env.Code {
	case 0, 200:
	default:
		msg := strings.TrimSpace(env.Message)
		if msg == "" {
			return nil, fmt.Errorf("GET %s: upstream returned code %d", path, env.Code)
		}
		return nil, fmt.Errorf("GET %s: upstream returned code %d: %s", path, env.Code, msg)
	}
	return env.Data, nil
}

func printJSONRaw(w io.Writer, data json.RawMessage) error {
	if w == nil {
		w = os.Stdout
	}
	if len(data) == 0 {
		_, err := fmt.Fprintln(w, "{}")
		return err
	}
	var pretty interface{}
	if err := json.Unmarshal(data, &pretty); err != nil {
		_, e := fmt.Fprintln(w, string(data))
		return e
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(pretty)
}
