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

// `olares-cli router app installed` — what is on this Olares.
//
// GET /console/api/installed-apps
//
// Three verbs in this tree answer questions that sound alike and are not. The
// catalog says what may be installed, from the Market. The provider list says
// what Router can call, which is only the applications that serve models. This
// says what is installed, model application or not — the application directory's
// own cache, and the only place a consumer like Wise appears at all.
//
// A state here is a snapshot as of the last sync, sharing the vocabulary of a
// provider's olares_status. An application can stop between two syncs, so a row
// saying `running` is a report of a moment rather than a promise about now.

// installedApp is one row of the directory cache. It is deliberately narrower
// than the stored record: the appid the platform derives and the list of people
// who installed a copy are a spend report's business, not a page asking whether
// something is here.
type installedApp struct {
	AppName string `json:"app_name"`
	Title   string `json:"title"`
	IconURL string `json:"icon_url,omitempty"`
	State   string `json:"state"`
	// Shared reports one installation serving everybody on this Olares, as
	// against a copy per person. It is what decides whether uninstalling it is
	// somebody else's problem too.
	Shared bool `json:"is_shared"`
}

func (a *installedApp) label() string {
	if s := strings.TrimSpace(a.Title); s != "" {
		return s
	}
	return a.AppName
}

func newAppInstalledCommand(f *cmdutil.Factory) *cobra.Command {
	var output string
	cmd := &cobra.Command{
		Use:   "installed",
		Short: "every application installed on this Olares",
		Long: `List the applications installed on this Olares.

This is the whole machine, not just the ones that serve models: an application
that calls Router appears here and nowhere else, since Router keeps no row for a
caller. "olares-cli router app catalog" is what may be installed, and
"olares-cli router provider list" is what Router can call.

STATE is as of the directory's last sync, in the same words a provider's status
uses. An application can stop a moment after it was read.

SHARED means one installation serves everybody here rather than a copy per
person, which is worth knowing before uninstalling one.

Readable by anyone with a console session.

Examples:
  olares-cli router app installed
  olares-cli router app installed -o json
`,
		Args: cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			return runAppInstalled(c.Context(), f, output)
		},
	}
	addOutputFlag(cmd, &output)
	return cmd
}

func runAppInstalled(ctx context.Context, f *cmdutil.Factory, outputRaw string) error {
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
	apps, err := listInstalledApps(ctx, pc)
	if err != nil {
		return err
	}
	if format == FormatJSON {
		return printJSON(os.Stdout, map[string]any{"items": apps})
	}
	return renderInstalledApps(os.Stdout, apps)
}

func listInstalledApps(ctx context.Context, pc *preparedClient) ([]installedApp, error) {
	apps, err := collection[installedApp](ctx, pc, epInstalledApps)
	if err != nil {
		return nil, err
	}
	// The repository orders by title, which is not the order a reader typed
	// anything in; sorting on the name keeps two runs comparable.
	sort.Slice(apps, func(i, j int) bool { return apps[i].AppName < apps[j].AppName })
	return apps, nil
}

func renderInstalledApps(w io.Writer, apps []installedApp) error {
	if len(apps) == 0 {
		_, err := fmt.Fprintln(w, "no application is installed, or the application directory has not "+
			"synced yet. \"olares-cli router status\" says whether it is running.")
		return err
	}
	t := newTable(w, "APP", "TITLE", "STATE", "SHARED")
	for i := range apps {
		a := &apps[i]
		t.row(a.AppName, nonEmpty(a.label()), nonEmpty(a.State), boolStr(a.Shared))
	}
	return t.flush()
}
