package model

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cliutil"
	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// `olares-cli model caller …` — the applications that call Router.
//
// GET    /console/api/apps
// DELETE /console/api/apps/:id
//
// These are consumers, not the model applications `model app` installs. An
// Olares application calling the data plane is recognised by the platform — the
// edge injects its identity — so Router creates the row on first call and issues
// it an internal key nobody types. That is why a caller appears here without
// anybody adding it, and why the list is a record of what has called rather than
// a list of what may.
//
// Archiving is the only write. It stops the application calling and does not
// delete it: the row keeps its history, and an archived caller does not come
// back on its own.

type callerApp struct {
	ID              string     `json:"id"`
	OlaresAppName   string     `json:"olares_app_name"`
	DisplayName     string     `json:"display_name"`
	Source          string     `json:"source"`
	Status          string     `json:"status"`
	FirstSeenAt     time.Time  `json:"first_seen_at"`
	LastSeenAt      time.Time  `json:"last_seen_at"`
	ArchivedAt      *time.Time `json:"archived_at,omitempty"`
	DefaultAPIKeyID string     `json:"default_api_key_id"`
}

func callerAppLabel(a *callerApp) string {
	if s := strings.TrimSpace(a.DisplayName); s != "" {
		return s
	}
	return a.OlaresAppName
}

func NewCallerCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "caller",
		Short: "the applications that call Router",
		Long: `List and archive the applications calling models through Router.

These are consumers — Wise, an agent, anything on this Olares asking for a
completion — and not the model applications that serve models. "olares-cli model
app" is the other one.

A caller appears by calling. The platform vouches for an application on this
Olares, so Router creates the row on first use and issues it an internal key
nobody has to hand over. Nothing here grants access; the list records what has
used Router, and archiving is how access is taken away.

Subcommands:
  list             the applications that have called, and when they last did
  archive <app>    stop an application calling

Admin only.
`,
	}
	cmd.SilenceUsage = true
	cmd.AddCommand(newCallerListCommand(f))
	cmd.AddCommand(newCallerArchiveCommand(f))
	return cmd
}

func newCallerListCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		output      string
		includeGone bool
	)
	cmd := &cobra.Command{
		Use:   "list",
		Short: "the applications that have called Router",
		Long: `List the calling applications Router knows, newest first.

LAST CALLED is the useful column: an application that has not called in months is
usually one that has been uninstalled or reconfigured, and archiving it costs
nothing.

Archived rows are hidden unless --all is given.

Examples:
  olares-cli model caller list
  olares-cli model caller list --all -o json
`,
		Args: cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			return runCallerList(c.Context(), f, includeGone, output)
		},
	}
	cmd.Flags().BoolVar(&includeGone, "all", false, "include archived applications")
	addOutputFlag(cmd, &output)
	return cmd
}

func runCallerList(ctx context.Context, f *cmdutil.Factory, includeGone bool, outputRaw string) error {
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
	apps, err := listCallerApps(ctx, pc)
	if err != nil {
		return err
	}
	kept := make([]callerApp, 0, len(apps))
	archived := 0
	for i := range apps {
		if apps[i].Status == "archived" || apps[i].ArchivedAt != nil {
			archived++
			if !includeGone {
				continue
			}
		}
		kept = append(kept, apps[i])
	}
	if format == FormatJSON {
		return printJSON(os.Stdout, map[string]any{"items": kept})
	}
	return renderCallerList(os.Stdout, kept, archived, includeGone)
}

func listCallerApps(ctx context.Context, pc *preparedClient) ([]callerApp, error) {
	var env struct {
		Items []callerApp `json:"items"`
	}
	if err := pc.router.doJSON(ctx, "GET", consoleAPI+"/apps", nil, &env); err != nil {
		return nil, err
	}
	return env.Items, nil
}

func renderCallerList(w io.Writer, apps []callerApp, archived int, includeGone bool) error {
	if len(apps) == 0 {
		msg := "no application has called Router yet."
		if archived > 0 && !includeGone {
			msg = fmt.Sprintf("no application is calling Router; %d archived one(s) are hidden, and --all shows them.", archived)
		}
		_, err := fmt.Fprintln(w, msg)
		return err
	}
	tw := tabwriter.NewWriter(w, 0, 4, 2, ' ', 0)
	if _, err := fmt.Fprintln(tw, "APPLICATION\tSTATUS\tFIRST CALLED\tLAST CALLED\tHAS KEY"); err != nil {
		return err
	}
	for i := range apps {
		a := &apps[i]
		if _, err := fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\n",
			nonEmpty(callerAppLabel(a)), nonEmpty(a.Status),
			a.FirstSeenAt.Local().Format("2006-01-02 15:04"),
			a.LastSeenAt.Local().Format("2006-01-02 15:04"),
			boolStr(a.DefaultAPIKeyID != "")); err != nil {
			return err
		}
	}
	if err := tw.Flush(); err != nil {
		return err
	}
	if archived > 0 && !includeGone {
		if _, err := fmt.Fprintf(w, "\n%d archived application(s) are hidden; --all shows them.\n", archived); err != nil {
			return err
		}
	}
	_, err := fmt.Fprintln(w, "\nHAS KEY is the internal key Router issued to the application itself. "+
		"It is not readable and is not one of `olares-cli model key list`.")
	return err
}

func newCallerArchiveCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		output    string
		assumeYes bool
	)
	cmd := &cobra.Command{
		Use:   "archive <app>",
		Short: "stop an application calling Router",
		Long: `Archive a calling application.

The application starts being refused, and its internal key is disabled with it.
Nothing is deleted: the row and everything it spent survive, and "usage" still
attributes its past calls.

An archived caller stays archived. Calling again does not revive it — which is
the point, and also the thing to know before archiving something still deployed:
it will keep failing until somebody notices.

Confirmation is required. --yes skips the prompt and is mandatory when stdin is
not a terminal.

Examples:
  olares-cli model caller archive wise --yes
`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			return runCallerArchive(c.Context(), f, args[0], assumeYes, output)
		},
	}
	cmd.Flags().BoolVar(&assumeYes, "yes", false, "skip the confirmation prompt (required when stdin is not a terminal)")
	addOutputFlag(cmd, &output)
	return cmd
}

func runCallerArchive(ctx context.Context, f *cmdutil.Factory, ref string, assumeYes bool, outputRaw string) error {
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
	found, err := resolveCallerApp(ctx, pc, ref)
	if err != nil {
		return err
	}
	if found.Status == "archived" || found.ArchivedAt != nil {
		_, werr := fmt.Fprintf(os.Stdout, "%s is already archived.\n", callerAppLabel(found))
		return werr
	}
	if !assumeYes {
		if err := cliutil.ConfirmDestructive(os.Stderr, os.Stdin,
			fmt.Sprintf("Archive %q? It last called %s, and it will be refused from now on without being told why.",
				callerAppLabel(found), found.LastSeenAt.Local().Format("2006-01-02 15:04")),
			false); err != nil {
			return err
		}
	}
	var archived callerApp
	path := consoleAPI + "/apps/" + url.PathEscape(found.ID)
	if err := pc.router.doJSON(ctx, "DELETE", path, nil, &archived); err != nil {
		return err
	}
	if format == FormatJSON {
		return printJSON(os.Stdout, archived)
	}
	_, err = fmt.Fprintf(os.Stdout, "%s is archived and its internal key is disabled; its usage history is kept\n",
		nonEmpty(callerAppLabel(&archived)))
	return err
}

// resolveCallerApp accepts the Olares application name, its display name, or the
// row id. The application name is what appears in a manifest and in a log, so it
// is the one anybody has to hand.
func resolveCallerApp(ctx context.Context, pc *preparedClient, ref string) (*callerApp, error) {
	ref = strings.TrimSpace(ref)
	if ref == "" {
		return nil, fmt.Errorf("an application name or id is required")
	}
	apps, err := listCallerApps(ctx, pc)
	if err != nil {
		return nil, err
	}
	for i := range apps {
		a := &apps[i]
		if a.ID == ref || strings.EqualFold(a.OlaresAppName, ref) || strings.EqualFold(a.DisplayName, ref) {
			return a, nil
		}
	}
	known := make([]string, 0, len(apps))
	for i := range apps {
		known = append(known, apps[i].OlaresAppName)
	}
	if len(known) == 0 {
		return nil, fmt.Errorf("no application %q; nothing has called Router yet, and a row appears only on the first call", ref)
	}
	return nil, fmt.Errorf("no application %q; the ones that have called are %s", ref, strings.Join(known, ", "))
}

func resolveCallerAppID(ctx context.Context, pc *preparedClient, ref string) (string, error) {
	found, err := resolveCallerApp(ctx, pc, ref)
	if err != nil {
		return "", err
	}
	return found.ID, nil
}
