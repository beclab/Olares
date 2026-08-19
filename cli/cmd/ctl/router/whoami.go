package router

import (
	"context"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// `olares-cli router whoami`
//
// Wraps GET /console/api/users/me — Router's own view of the caller, which is
// not the same question `olares-cli profile whoami` answers. The profile verb
// reports who the CLI is to Olares; this reports the user row Router created
// from the X-BFL-USER header the edge injected, including the role that
// decides whether the admin-only verbs in this tree will work at all.

// consoleUser mirrors Router's users row as it appears on the wire.
type consoleUser struct {
	ID            string   `json:"id"`
	BflName       string   `json:"bfl_name"`
	Role          string   `json:"role"`
	Status        string   `json:"status"`
	Source        string   `json:"source"`
	OlaresID      string   `json:"olares_id,omitempty"`
	AllowedModels []string `json:"allowed_models,omitempty"`
}

func (u *consoleUser) isAdmin() bool {
	return strings.EqualFold(u.Role, "admin") || strings.EqualFold(u.Role, "owner")
}

func NewWhoamiCommand(f *cmdutil.Factory) *cobra.Command {
	var output string
	cmd := &cobra.Command{
		Use:   "whoami",
		Short: "show the identity and role Router sees for the active profile",
		Long: `Report the user record Router holds for the active profile.

Router derives it from the X-BFL-USER header the Olares edge injects, so
this is the identity every "model" verb acts as. The role matters: most of
Router's management surface is admin-only, and a non-admin session is
rejected with "admin role required" rather than an authentication error.

Use "olares-cli profile whoami" for the complementary question of who the
CLI is to Olares itself.
`,
		Args: cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			return runWhoami(c.Context(), f, output)
		},
	}
	addOutputFlag(cmd, &output)
	return cmd
}

func runWhoami(ctx context.Context, f *cmdutil.Factory, outputRaw string) error {
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
	me, err := fetchConsoleUser(ctx, pc)
	if err != nil {
		return err
	}
	if format == FormatJSON {
		return printJSON(os.Stdout, me)
	}
	return renderWhoami(os.Stdout, me)
}

func fetchConsoleUser(ctx context.Context, pc *preparedClient) (*consoleUser, error) {
	var me consoleUser
	if err := pc.router.doJSON(ctx, "GET", epMe, nil, &me); err != nil {
		return nil, err
	}
	return &me, nil
}

// userLabels maps Router's user ids to the names a person recognises, for the
// columns that would otherwise print a UUID. Failing to read the list is not
// worth failing the caller over — the id is still printed — so this returns an
// empty map rather than an error.
func userLabels(ctx context.Context, pc *preparedClient) map[string]string {
	var env struct {
		Items []consoleUser `json:"items"`
	}
	if err := pc.router.doJSON(ctx, "GET", epUsers, nil, &env); err != nil {
		return map[string]string{}
	}
	out := make(map[string]string, len(env.Items))
	for i := range env.Items {
		u := &env.Items[i]
		if u.ID != "" && u.BflName != "" {
			out[u.ID] = u.BflName
		}
	}
	return out
}

func renderWhoami(w io.Writer, me *consoleUser) error {
	t := newTable(w)
	t.row("USER", nonEmpty(me.BflName))
	t.row("ROLE", nonEmpty(me.Role))
	t.row("ADMIN", boolStr(me.isAdmin()))
	t.row("STATUS", nonEmpty(me.Status))
	t.row("SOURCE", nonEmpty(me.Source))
	t.row("ID", nonEmpty(me.ID))
	if len(me.AllowedModels) > 0 {
		t.row("ALLOWED MODELS", strings.Join(me.AllowedModels, ", "))
	}
	return t.flush()
}
