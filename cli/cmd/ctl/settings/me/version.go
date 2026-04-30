package me

import (
	"context"
	"os"

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
