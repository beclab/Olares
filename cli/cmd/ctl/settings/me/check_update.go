package me

import (
	"context"
	"os"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// `olares-cli settings me check-update`
//
// Wraps user-service's GET /api/checkLastOsVersion (init.controller.ts:77)
// which proxies Olares Space (init.service.ts:32 → POST
// olaresSpaceUrl/v1/resource/lastVersions). The handler returns a
// `Result<...>` with three fields the SPA's upgrade store reads:
//
//	current_version  string  - OS version installed on this Olares
//	new_version      string  - latest version Olares Space knows about
//	is_new           bool    - whether new_version is actually newer
//
// The SPA exposes this in apps/.../stores/settings/upgrade.ts but no
// active Settings page currently calls it (the UI is in flight). We
// expose it here because it's the cheapest way to answer "should I
// upgrade?" from a CLI without going through the Olares Space mobile
// path (which lives in apps/.../stores/mdns.ts and uses a different
// origin entirely).
//
// Role: any authenticated user; no PreflightRole.
func NewCheckUpdateCommand(f *cmdutil.Factory) *cobra.Command {
	var output string
	cmd := &cobra.Command{
		Use:   "check-update",
		Short: "check whether a newer Olares OS release is available (Settings -> upgrade)",
		Long: `Ask Olares Space (via user-service) whether a newer OS version is
available than the one currently installed on the target Olares.

Returns three fields:
  current_version  the OS version this Olares is running
  new_version      the latest version Olares Space knows about
  is_new           whether new_version is strictly newer

The CLI's exit code is always 0 on a successful query — even when
is_new is false. Use --output json + jq if you need a non-zero exit
in CI when there's an update available.
`,
		Args: cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			return runCheckUpdate(c.Context(), f, output)
		},
	}
	addOutputFlag(cmd, &output)
	return cmd
}

// checkUpdateResp mirrors the body returned by InitService.getOSVersion
// (user-service/src/init.service.ts:32-79). The SPA's upgrade store reads
// these exact field names, and we keep them verbatim in JSON output so
// existing scripts written against the SPA's network responses keep
// working.
type checkUpdateResp struct {
	CurrentVersion string `json:"current_version"`
	NewVersion     string `json:"new_version"`
	IsNew          bool   `json:"is_new"`
}

func runCheckUpdate(ctx context.Context, f *cmdutil.Factory, outputRaw string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	format, err := parseFormat(outputRaw)
	if err != nil {
		return err
	}

	pc, err := prepare(ctx, f)
	if err != nil {
		return err
	}

	var resp checkUpdateResp
	if err := doGetEnvelope(ctx, pc.doer, "/api/checkLastOsVersion", &resp); err != nil {
		return err
	}

	switch format {
	case FormatJSON:
		return printJSON(os.Stdout, resp)
	default:
		availability := "up to date"
		if resp.IsNew {
			availability = "update available"
		}
		return printKV(os.Stdout, [][2]string{
			{"Current Version", nonEmpty(resp.CurrentVersion)},
			{"Latest Version", nonEmpty(resp.NewVersion)},
			{"Status", availability},
		}, 16)
	}
}
