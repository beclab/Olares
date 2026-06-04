package me

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// `olares-cli settings me version`
//
// Maps to the SPA's Person -> Version page (and the "Version" line on the
// Person home page). Both call axios.get('/api/init') and read
// `terminusInfo.osVersion`, but `/api/init` is a heavy aggregate
// (terminusInfo + userInfo + applicationData + secrets + devices +
// wallpaper). For a CLI verb that only wants the version, we hit the
// dedicated `/api/olares-info` endpoint instead — same data shape under
// `data` (an OlaresInfo), no extra payload.
//
// Backend handler: user-service/src/init.controller.ts:46-50
//   handler delegates to OlaresService.updateOlaresInfo(), which proxies
//   bfl /bfl/backend/v1/olares-info (framework/bfl/.../handler.go:296)
//   and returns a `Result<OlaresInfo>` envelope.
//
// Output:
//   table  -> tidy 2-column "Field: Value" rendering of the most useful
//             OlaresInfo fields (OS version + identity bits)
//   json   -> the full OlaresInfo struct so jq/yq scripting works without
//             re-running other verbs
//
// Role: any authenticated user can fetch their own OS version, so no
// PreflightRole check.
func NewVersionCommand(f *cmdutil.Factory) *cobra.Command {
	var output string
	var dispatch bool
	cmd := &cobra.Command{
		Use:   "version",
		Short: "show the installed Olares OS version (Settings -> Person -> Version)",
		Long: `Show the running Olares OS version on the target instance, plus a
handful of related identity fields exposed by the same /api/olares-info
endpoint the SPA uses.

In table mode (default) you get the OS version, terminus identity, and
the wizard / reverse-proxy / TailScale flags the Person page surfaces.
In JSON mode you get the full OlaresInfo struct verbatim — useful for
scripting against future fields without waiting on a CLI bump.

--dispatch shows the version-compat view instead: the backend version
the CLI will dispatch remote/API commands on, where it came from
(--olares-version flag / per-profile cache / a fresh /api/olares-info
read), when the cache was last refreshed, and which versioned client
implementation handles your commands. Combine with --refresh-version to
force a fresh read, or --olares-version to preview a specific version's
dispatch.
`,
		Args: cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			if dispatch {
				return runVersionDispatch(c.Context(), f, output)
			}
			return runVersion(c.Context(), f, output)
		},
	}
	addOutputFlag(cmd, &output)
	cmd.Flags().BoolVar(&dispatch, "dispatch", false, "show the backend version the CLI dispatches commands on (source, cache age, selected client) instead of the live OS version")
	return cmd
}

// olaresInfoResp mirrors framework/bfl/pkg/apis/backend/v1/model.go's
// OlaresInfo. We embed json tags rather than importing the type because
// (a) the bfl package isn't a stable public surface for CLI consumers
// and (b) we want to be explicit about which fields the table rendering
// promises to surface.
type olaresInfoResp struct {
	OlaresID           string `json:"olaresId"`
	WizardStatus       string `json:"wizardStatus"`
	EnableReverseProxy bool   `json:"enableReverseProxy"`
	TailScaleEnable    bool   `json:"tailScaleEnable"`
	OsVersion          string `json:"osVersion"`
	LoginBackground    string `json:"loginBackground"`
	Avatar             string `json:"avatar"`
	ID                 string `json:"id"`
	UserDID            string `json:"did"`
	Olaresd            string `json:"olaresd"`
	Style              string `json:"style"`
}

func runVersion(ctx context.Context, f *cmdutil.Factory, outputRaw string) error {
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

	var info olaresInfoResp
	if err := doGetEnvelope(ctx, pc.doer, "/api/olares-info", &info); err != nil {
		return err
	}

	switch format {
	case FormatJSON:
		return printJSON(os.Stdout, info)
	default:
		return printKV(os.Stdout, [][2]string{
			{"OS Version", nonEmpty(info.OsVersion)},
			{"Olares ID", nonEmpty(info.OlaresID)},
			{"Wizard Status", nonEmpty(info.WizardStatus)},
			{"Reverse Proxy", boolStr(info.EnableReverseProxy)},
			{"TailScale", boolStr(info.TailScaleEnable)},
			{"Olaresd", nonEmpty(info.Olaresd)},
		}, 14)
	}
}

// runVersionDispatch renders the version-compat dispatch view: which backend
// version the CLI resolves for remote/API commands, its provenance, the cache
// timestamp, and the selected client implementation. Honors --refresh-version
// and --olares-version (both persistent root flags) via the Factory.
func runVersionDispatch(ctx context.Context, f *cmdutil.Factory, outputRaw string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	format, err := parseFormat(outputRaw)
	if err != nil {
		return err
	}

	info, err := f.OlaresBackendVersionInfo(ctx)
	if err != nil {
		return err
	}

	if format == FormatJSON {
		out := struct {
			Version        string `json:"version"`
			Source         string `json:"source"`
			CachedVersion  string `json:"cachedVersion,omitempty"`
			RefreshedAt    int64  `json:"refreshedAt,omitempty"`
			RefreshedAtRFC string `json:"refreshedAtRFC,omitempty"`
			TTLSeconds     int64  `json:"ttlSeconds"`
			Implementation string `json:"implementation"`
		}{
			Version:        versionString(info.Version),
			Source:         string(info.Source),
			CachedVersion:  info.CachedVersion,
			RefreshedAt:    info.RefreshedAt,
			TTLSeconds:     int64(info.TTL.Seconds()),
			Implementation: info.Implementation,
		}
		if info.RefreshedAt > 0 {
			out.RefreshedAtRFC = time.Unix(info.RefreshedAt, 0).Format(time.RFC3339)
		}
		return printJSON(os.Stdout, out)
	}

	refreshed := "-"
	if info.RefreshedAt > 0 {
		t := time.Unix(info.RefreshedAt, 0)
		refreshed = fmt.Sprintf("%s (%s ago)", t.Format("2006-01-02 15:04:05"), durShort(time.Since(t)))
	}
	return printKV(os.Stdout, [][2]string{
		{"Backend Version", versionString(info.Version)},
		{"Source", string(info.Source)},
		{"Cached Value", nonEmpty(info.CachedVersion)},
		{"Refreshed At", refreshed},
		{"TTL", info.TTL.String()},
		{"Dispatches To", nonEmpty(info.Implementation) + " client"},
	}, 18)
}

func versionString(v interface{ String() string }) string {
	if v == nil {
		return "-"
	}
	return v.String()
}

// durShort renders a duration as a coarse "12m" / "3h" / "2d" string for the
// "refreshed N ago" hint. Sub-minute ages collapse to "0m" to avoid noisy
// seconds in a human-facing freshness indicator.
func durShort(d time.Duration) string {
	if d < time.Minute {
		return "0m"
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%dh", int(d.Hours()))
	}
	return fmt.Sprintf("%dd", int(d.Hours()/24))
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
