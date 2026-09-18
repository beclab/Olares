package router

import (
	"context"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// `olares-cli router usage apps`
// GET /console/api/installed-apps
//
// The applications on this Olares, which is the question `usage --by
// caller_app` leaves open. That report names an application by its title when
// Router has one and by its appid when it does not, and either way says only
// what was spent. It cannot say whether an application that spent nothing is
// stopped, idle, or gone — and those are different situations with different
// fixes.
//
// This is the directory's own cache, so a row here means the application is
// installed. `STATE` is a snapshot as of the last sync rather than a live
// reading: an application can stop between two of them.
//
// Readable by any authenticated console user, unlike the caller_app dimension
// beside it. An install is not a secret; what an application spent is.

type installedApp struct {
	AppName string `json:"app_name"`
	Title   string `json:"title"`
	IconURL string `json:"icon_url,omitempty"`
	State   string `json:"state"`
	Shared  bool   `json:"is_shared"`
}

func newUsageAppsCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		output string
		state  string
	)
	cmd := &cobra.Command{
		Use:   "apps",
		Short: "the applications installed on this Olares",
		Long: `List the applications installed here.

"usage summary --by caller_app" says what each application spent, and names one
by its appid when Router has never had a title for it. This says what is
installed, which is what makes an empty bucket readable: an application with no
calls is stopped, or idle, or was uninstalled and left its spend behind.

STATE is a snapshot from the last directory sync, not a live reading. An
application can stop between two syncs, so treat a "running" here as recent
rather than current — "olares-cli market list --mine" is the live answer.

SHARED marks an application installed for everyone on this Olares rather than
for one person.

Unlike the rest of "usage", this is readable by any console user. An install is
not a secret; what an application spent is.

Examples:
  olares-cli router usage apps
  olares-cli router usage apps --state running
`,
		Args: cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			return runUsageApps(c.Context(), f, strings.TrimSpace(state), output)
		},
	}
	cmd.Flags().StringVar(&state, "state", "", "only applications in this state, e.g. running")
	addOutputFlag(cmd, &output)
	return cmd
}

func runUsageApps(ctx context.Context, f *cmdutil.Factory, state, outputRaw string) error {
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
	apps, err := collection[installedApp](ctx, pc, epInstalledApps)
	if err != nil {
		return err
	}
	kept, states := filterInstalledApps(apps, state)
	if state != "" && len(kept) == 0 && len(apps) > 0 {
		return missing{
			noun: "state", ref: state, known: states,
			have: "applications are in", none: "no application reports a state",
		}.err()
	}
	if format == FormatJSON {
		return printJSON(os.Stdout, kept)
	}
	return renderInstalledApps(os.Stdout, kept)
}

// filterInstalledApps also reports the states that were there, so a --state
// nothing matches can be refused with the vocabulary this Olares actually
// uses. The states come from the platform rather than from Router, so a list
// spelled out here would be a guess.
func filterInstalledApps(apps []installedApp, state string) ([]installedApp, []string) {
	seen := map[string]bool{}
	kept := make([]installedApp, 0, len(apps))
	for _, a := range apps {
		if s := strings.TrimSpace(a.State); s != "" {
			seen[s] = true
		}
		if state == "" || strings.EqualFold(a.State, state) {
			kept = append(kept, a)
		}
	}
	states := make([]string, 0, len(seen))
	for s := range seen {
		states = append(states, s)
	}
	sort.Strings(states)
	return kept, states
}

func renderInstalledApps(w io.Writer, apps []installedApp) error {
	if len(apps) == 0 {
		_, err := fmt.Fprintln(w, "no application is installed on this Olares.")
		return err
	}
	t := newTable(w, "APPLICATION", "TITLE", "STATE", "SHARED")
	for i := range apps {
		a := &apps[i]
		t.row(a.AppName, nonEmpty(a.Title), nonEmpty(a.State), boolStr(a.Shared))
	}
	if err := t.flush(); err != nil {
		return err
	}
	_, err := fmt.Fprintln(w, "\n`olares-cli router usage summary --by caller_app` says what each "+
		"of these spent; APPLICATION is what `--caller-app` takes.")
	return err
}
