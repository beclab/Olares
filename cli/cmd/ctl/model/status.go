package model

import (
	"context"
	"fmt"
	"io"
	"os"
	"sort"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// `olares-cli model status`
//
// The verb to run first, and the one to run when something else in this tree
// failed. It answers the three questions that precede every other model
// operation: is Router installed and where, is its backend answering, and does
// the active profile have the role its management surface demands.
//
// Each probe is reported independently. A 403 from the identity read or a dead
// backend is exactly what this verb exists to show, so a failing probe is
// recorded and the rest still run.

type statusReport struct {
	Router      *discoveredRouter      `json:"router"`
	Health      map[string]interface{} `json:"health,omitempty"`
	HealthError string                 `json:"health_error,omitempty"`
	Identity    *consoleUser           `json:"identity,omitempty"`
	IdentityErr string                 `json:"identity_error,omitempty"`
}

func NewStatusCommand(f *cmdutil.Factory) *cobra.Command {
	var output string
	cmd := &cobra.Command{
		Use:   "status",
		Short: "show where Router lives, whether it is healthy, and your role",
		Long: `Report the state of the AI gateway for the active profile.

Three independent probes:

  router      which Market app id serves Router, its entrance and base URL
  health      Router's own /healthz — process, database, migrations, and the
              optional LLDAP / NATS / Olares-sync subsystems
  identity    the user and role Router sees, since most management verbs
              require admin

A probe that fails is reported rather than aborting the command: "the
backend is unreachable" and "you are not an admin" are answers, not errors,
and this is the verb that should surface them.
`,
		Args: cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			return runStatus(c.Context(), f, output)
		},
	}
	addOutputFlag(cmd, &output)
	return cmd
}

func runStatus(ctx context.Context, f *cmdutil.Factory, outputRaw string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	format, err := parseFormat(outputRaw)
	if err != nil {
		return err
	}
	// Discovery failing is terminal: without an address there is nothing left
	// to probe, and its error already says whether Router is missing or the
	// profile cannot list apps.
	pc, err := prepare(ctx, f)
	if err != nil {
		return err
	}

	report := statusReport{Router: pc.found}
	var health map[string]interface{}
	if herr := pc.router.doJSON(ctx, "GET", "/healthz", nil, &health); herr != nil {
		report.HealthError = herr.Error()
	} else {
		report.Health = health
	}
	if me, merr := fetchConsoleUser(ctx, pc); merr != nil {
		report.IdentityErr = merr.Error()
	} else {
		report.Identity = me
	}

	if format == FormatJSON {
		return printJSON(os.Stdout, report)
	}
	return renderStatus(os.Stdout, report)
}

func renderStatus(w io.Writer, r statusReport) error {
	tw := tabwriter.NewWriter(w, 0, 4, 2, ' ', 0)
	rows := [][2]string{
		{"APP", nonEmpty(r.Router.AppName)},
		{"TITLE", nonEmpty(r.Router.Title)},
		{"STATE", nonEmpty(r.Router.State)},
		{"ENTRANCE", nonEmpty(r.Router.EntranceName)},
		{"BASE URL", nonEmpty(r.Router.BaseURL)},
	}

	switch {
	case r.HealthError != "":
		rows = append(rows, [2]string{"HEALTH", "unreachable: " + r.HealthError})
	default:
		keys := make([]string, 0, len(r.Health))
		for k := range r.Health {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			rows = append(rows, [2]string{"HEALTH " + k, fmt.Sprintf("%v", r.Health[k])})
		}
	}

	switch {
	case r.IdentityErr != "":
		rows = append(rows, [2]string{"IDENTITY", "unavailable: " + r.IdentityErr})
	case r.Identity != nil:
		rows = append(rows,
			[2]string{"USER", nonEmpty(r.Identity.BflName)},
			[2]string{"ROLE", nonEmpty(r.Identity.Role)},
			[2]string{"ADMIN", boolStr(r.Identity.isAdmin())},
		)
	}

	for _, row := range rows {
		if _, err := fmt.Fprintf(tw, "%s\t%s\n", row[0], row[1]); err != nil {
			return err
		}
	}
	return tw.Flush()
}
