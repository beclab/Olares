package wizard

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"

	wizardpkg "github.com/beclab/Olares/cli/pkg/wizard"
	"github.com/spf13/cobra"
)

// `olares-cli wizard frp <olaresId>`
//
// Public, unauthenticated lookup against the Olares-tunnel registry —
// the same call the activation wizard's "select tunnel" step makes
// before the user has finished binding (see
// TermiPass/packages/app/src/stores/wizard-step.ts:getFrpList).
//
// Useful for picking the `--host` value to feed into
// `olares-cli wizard activate ... --enable-tunnel --host <host>` from
// scripts.
func NewCmdFrp() *cobra.Command {
	var (
		outputFormat string
		locale       string
		environment  string
	)

	cmd := &cobra.Command{
		Use:   "frp {Olares ID (e.g., user@example.com)}",
		Short: "list public Olares-tunnel (FRP) servers for an Olares ID",
		Long: `Fetch the public Olares-tunnel (FRP) server registry for an Olares ID
without requiring authentication.

The endpoint is auto-selected from the olaresId suffix:
  - *.olares.cn  → https://api.olares.cn/frp/v2/servers
  - everything else → https://api.olares.com/frp/v2/servers

Use --env=cn|en to override.

Default output is a table with REGION / NAME / HOST columns; use
--output=json for the raw response (multilingual name map + every
machine host).`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			olaresID := strings.TrimSpace(args[0])
			if olaresID == "" {
				return fmt.Errorf("olaresId is required")
			}

			format, err := parseFRPOutputFormat(outputFormat)
			if err != nil {
				return err
			}

			env, err := parseFRPEnvironment(environment)
			if err != nil {
				return err
			}

			list, err := wizardpkg.FetchFrpList(cmd.Context(), olaresID, wizardpkg.FrpListOptions{
				Environment: env,
			})
			if err != nil {
				return err
			}

			switch format {
			case frpFormatJSON:
				return writeFRPJSON(cmd.OutOrStdout(), list)
			default:
				return writeFRPTable(cmd.OutOrStdout(), list, locale)
			}
		},
	}

	cmd.SilenceUsage = true
	cmd.Flags().StringVarP(&outputFormat, "output", "o", "table", "output format: table | json")
	cmd.Flags().StringVar(&locale, "lang", "en-US", "locale to render the NAME column (e.g. en-US, zh-CN)")
	cmd.Flags().StringVar(&environment, "env", "", "force API environment: cn | en (default: auto-derived from olaresId)")
	return cmd
}

type frpOutputFormat int

const (
	frpFormatTable frpOutputFormat = iota
	frpFormatJSON
)

func parseFRPOutputFormat(s string) (frpOutputFormat, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "", "table":
		return frpFormatTable, nil
	case "json":
		return frpFormatJSON, nil
	default:
		return frpFormatTable, fmt.Errorf("unsupported output format %q (want table | json)", s)
	}
}

func parseFRPEnvironment(s string) (wizardpkg.FrpEnvironment, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "":
		return "", nil
	case "cn":
		return wizardpkg.FrpEnvCN, nil
	case "en", "com":
		return wizardpkg.FrpEnvEN, nil
	default:
		return "", fmt.Errorf("unsupported --env %q (want cn | en)", s)
	}
}

func writeFRPJSON(w io.Writer, list []wizardpkg.FrpServer) error {
	if list == nil {
		list = []wizardpkg.FrpServer{}
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(list)
}

func writeFRPTable(w io.Writer, list []wizardpkg.FrpServer, locale string) error {
	if len(list) == 0 {
		_, err := fmt.Fprintln(w, "no FRP servers")
		return err
	}
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	if _, err := fmt.Fprintln(tw, "REGION\tNAME\tHOST"); err != nil {
		return err
	}
	for _, s := range list {
		region := nonEmptyStr(s.Region)
		name := nonEmptyStr(s.LocalizedName(locale))
		hosts := collectHosts(s.Machine)
		if _, err := fmt.Fprintf(tw, "%s\t%s\t%s\n", region, name, hosts); err != nil {
			return err
		}
	}
	return tw.Flush()
}

func collectHosts(ms []wizardpkg.FrpMachine) string {
	if len(ms) == 0 {
		return "-"
	}
	hosts := make([]string, 0, len(ms))
	for _, m := range ms {
		if h := strings.TrimSpace(m.Host); h != "" {
			hosts = append(hosts, h)
		}
	}
	if len(hosts) == 0 {
		return "-"
	}
	return strings.Join(hosts, ",")
}

func nonEmptyStr(s string) string {
	if strings.TrimSpace(s) == "" {
		return "-"
	}
	return s
}
