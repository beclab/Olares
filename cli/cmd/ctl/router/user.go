package router

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// `olares-cli router user list` — GET /console/api/users
//
// Router's users are not Olares' users. A row appears here the first time
// somebody's session reaches Router, created from the X-BFL-USER header the edge
// injects, which means an Olares account that has never opened Router is absent
// and that absence says nothing about the account.
//
// There is nothing to write. Roles come from Olares — Router reads whether the
// account is an Olares admin — so promoting somebody is done in Olares Settings,
// not here.

func NewUserCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "user",
		Short: "the people Router knows",
		Long: `List the users Router has records for.

A record appears the first time somebody's session reaches Router, so this is the
people who have used it rather than everybody who could. Roles are Olares' to
decide and are read from there; there is nothing to change here.

Use "olares-cli router whoami" for your own record, which any user can read.

Admin only.
`,
	}
	cmd.SilenceUsage = true
	cmd.AddCommand(newUserListCommand(f))
	return cmd
}

func newUserListCommand(f *cmdutil.Factory) *cobra.Command {
	var output string
	cmd := &cobra.Command{
		Use:   "list",
		Short: "the users Router has records for",
		Long: `List every user Router knows, newest first.

The list is what other verbs mean by a user: "key issue --for-user" takes one of
these names, and "router usage --user" narrows spend to one of them.

Examples:
  olares-cli router user list
  olares-cli router user list -o json
`,
		Args: cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			return runUserList(c.Context(), f, output)
		},
	}
	addOutputFlag(cmd, &output)
	return cmd
}

func runUserList(ctx context.Context, f *cmdutil.Factory, outputRaw string) error {
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
	users, err := listUsers(ctx, pc)
	if err != nil {
		return err
	}
	if format == FormatJSON {
		return printJSON(os.Stdout, map[string]any{"items": users})
	}
	return renderUserList(os.Stdout, users)
}

func listUsers(ctx context.Context, pc *preparedClient) ([]consoleUser, error) {
	var env struct {
		Items []consoleUser `json:"items"`
	}
	if err := pc.router.doJSON(ctx, "GET", epUsers, nil, &env); err != nil {
		return nil, err
	}
	return env.Items, nil
}

func renderUserList(w io.Writer, users []consoleUser) error {
	if len(users) == 0 {
		_, err := fmt.Fprintln(w, "no users. A record appears the first time somebody opens Router.")
		return err
	}
	tw := tabwriter.NewWriter(w, 0, 4, 2, ' ', 0)
	if _, err := fmt.Fprintln(tw, "USER\tROLE\tSTATUS\tSOURCE\tMODELS\tID"); err != nil {
		return err
	}
	for i := range users {
		u := &users[i]
		models := "all"
		if len(u.AllowedModels) > 0 {
			models = strings.Join(u.AllowedModels, ", ")
		}
		if _, err := fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%s\n",
			nonEmpty(u.BflName), nonEmpty(u.Role), nonEmpty(u.Status),
			nonEmpty(u.Source), clip(models, 40), nonEmpty(u.ID)); err != nil {
			return err
		}
	}
	return tw.Flush()
}

// resolveUserID turns a name into the id the routes take. Names are what people
// know; the id appears only in Router's own rows.
func resolveUserID(ctx context.Context, pc *preparedClient, ref string) (string, error) {
	ref = strings.TrimSpace(ref)
	if ref == "" {
		return "", fmt.Errorf("a user name or id is required")
	}
	if providerIDPattern.MatchString(ref) {
		return ref, nil
	}
	users, err := listUsers(ctx, pc)
	if err != nil {
		return "", fmt.Errorf("look up user %q: %w", ref, err)
	}
	for i := range users {
		if strings.EqualFold(users[i].BflName, ref) || strings.EqualFold(users[i].OlaresID, ref) {
			return users[i].ID, nil
		}
	}
	known := make([]string, 0, len(users))
	for i := range users {
		known = append(known, users[i].BflName)
	}
	if len(known) == 0 {
		return "", fmt.Errorf("no user %q; Router has no user records yet, which happens before anybody has opened it", ref)
	}
	return "", fmt.Errorf("no user %q; Router knows %s. A user appears only after their first visit to Router",
		ref, strings.Join(known, ", "))
}
