package me

import (
	"context"
	"os"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
	"github.com/beclab/Olares/cli/pkg/whoami"
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
// The olares-info round-trip is shared with the version-cache layer via
// whoami.FetchOlaresInfo, so there is one olares-info implementation in the
// CLI rather than a per-call copy.
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
`,
		Args: cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			return runVersion(c.Context(), f, output)
		},
	}
	addOutputFlag(cmd, &output)
	return cmd
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

	info, err := whoami.FetchOlaresInfo(ctx, pc.doer)
	if err != nil {
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
