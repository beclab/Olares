package router

import (
	"context"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// `olares-cli router provider history <name|id>`
// `olares-cli router provider rollback <name|id> <version>`
// GET  /console/api/providers/:id/credential-history
// POST /console/api/providers/:id/rollback/:version
//
// Router keeps every credential a provider has held, encrypted, with who
// changed it and when. The values are never readable — not here and not in
// Router's own console — so what the history gives you is the shape of the
// change: how many rotations, by whom, and the notes attached.
//
// A rollback restores one of those stored versions without decrypting anything:
// it copies the old ciphertext back and takes a new version number. So rolling
// back to v2 leaves you at v5, not v2, and that v5 is itself in the history.
// Nothing is ever overwritten, which is why a rollback can be rolled back.

type credentialVersion struct {
	ID              string  `json:"id"`
	ProviderID      string  `json:"provider_id"`
	Version         int     `json:"version"`
	ChangedByUserID *string `json:"changed_by_user_id,omitempty"`
	Note            *string `json:"note,omitempty"`
	CreatedAt       string  `json:"created_at"`
}

func newProviderHistoryCommand(f *cmdutil.Factory) *cobra.Command {
	var output string
	cmd := &cobra.Command{
		Use:   "history <name|id>",
		Short: "the credential versions a provider has held, newest first",
		Long: `List a provider's credential versions.

Each rotation writes an entry: the version it produced, who did it, when, and
the note they left. The credentials themselves stay encrypted and are not
readable from any route, so this is a record of changes rather than of values.

Use it to answer "when was this key last rotated" and to find the version to
restore with "provider rollback".

A provider with an empty history has never had its credentials changed since it
was created.

Example:
  olares-cli router provider history openai-prod
`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			return runProviderHistory(c.Context(), f, args[0], output)
		},
	}
	addOutputFlag(cmd, &output)
	return cmd
}

func runProviderHistory(ctx context.Context, f *cmdutil.Factory, ref, outputRaw string) error {
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
	found, err := resolveProvider(ctx, pc, ref)
	if err != nil {
		return err
	}
	var env struct {
		Items []credentialVersion `json:"items"`
		Total int                 `json:"total"`
	}
	path := epProviderCredentialHistory(found.ID)
	if err := pc.router.doJSON(ctx, "GET", path, nil, &env); err != nil {
		return err
	}
	if format == FormatJSON {
		return printJSON(os.Stdout, env)
	}
	return renderCredentialHistory(os.Stdout, found, env.Items, userLabels(ctx, pc))
}

func renderCredentialHistory(w io.Writer, p *providerRow, items []credentialVersion, whoBy map[string]string) error {
	if _, err := fmt.Fprintf(w, "%s is at credential version %d\n\n", p.Name, p.CredentialsVersion); err != nil {
		return err
	}
	if len(items) == 0 {
		_, err := fmt.Fprintln(w, "no recorded changes — these are the credentials the provider was created with")
		return err
	}
	tw := tabwriter.NewWriter(w, 0, 4, 2, ' ', 0)
	if _, err := fmt.Fprintln(tw, "VERSION\tRECORDED\tCHANGED BY\tNOTE"); err != nil {
		return err
	}
	for i := range items {
		it := &items[i]
		marker := strconv.Itoa(it.Version)
		if it.Version == p.CredentialsVersion {
			marker += " (current)"
		}
		changedBy := derefOr(it.ChangedByUserID, "")
		if name := whoBy[changedBy]; name != "" {
			changedBy = name
		}
		if _, err := fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n",
			marker,
			nonEmpty(it.CreatedAt),
			nonEmpty(changedBy),
			nonEmpty(clip(derefOr(it.Note, ""), 60)),
		); err != nil {
			return err
		}
	}
	return tw.Flush()
}

func derefOr(s *string, fallback string) string {
	if s == nil {
		return fallback
	}
	return *s
}

func newProviderRollbackCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		output string
		note   string
	)
	cmd := &cobra.Command{
		Use:   "rollback <name|id> <version>",
		Short: "restore an earlier credential version",
		Long: `Restore the credentials a provider held at an earlier version.

The stored ciphertext for that version is copied back as the current
credentials, and the provider moves to a new version number rather than
returning to the old one. Rolling back to version 2 from version 4 leaves the
provider at version 5, whose history entry records where it came from. Nothing
is overwritten, so a rollback can itself be rolled back.

Router never decrypts anything to do this, which is why it works even for a
version whose key you no longer have.

Rolling back does not check that the restored credentials still work. Follow it
with "provider validate" — a key rotated because it was revoked will be just as
revoked after being restored.

"provider history" lists the versions available.

Example:
  olares-cli router provider rollback openai-prod 2 --note "new key was wrong"
`,
		Args: cobra.ExactArgs(2),
		RunE: func(c *cobra.Command, args []string) error {
			return runProviderRollback(c.Context(), f, args[0], args[1], note, output)
		},
	}
	cmd.Flags().StringVar(&note, "note", "", "reason for the rollback, recorded in credential history")
	addOutputFlag(cmd, &output)
	return cmd
}

func runProviderRollback(ctx context.Context, f *cmdutil.Factory, ref, versionRaw, note, outputRaw string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	format, err := parseFormat(outputRaw)
	if err != nil {
		return err
	}
	version, err := strconv.Atoi(strings.TrimSpace(versionRaw))
	if err != nil || version < 1 {
		return fmt.Errorf("version must be a positive number; `olares-cli router provider history %s` lists them", ref)
	}
	pc, err := prepare(ctx, f)
	if err != nil {
		return err
	}
	found, err := resolveProvider(ctx, pc, ref)
	if err != nil {
		return err
	}
	if found.isMarketSourced() {
		return marketOwnedErr(found, "roll back the credentials of")
	}

	var body any
	if strings.TrimSpace(note) != "" {
		body = map[string]string{"note": note}
	}
	var res struct {
		Provider    providerRow `json:"provider"`
		FromVersion int         `json:"from_version"`
		ToVersion   int         `json:"to_version"`
		NewVersion  int         `json:"new_version"`
	}
	path := epProviderRollback(found.ID, version)
	if err := pc.router.doJSON(ctx, "POST", path, body, &res); err != nil {
		return err
	}
	if format == FormatJSON {
		return printJSON(os.Stdout, res)
	}
	if _, err := fmt.Fprintf(os.Stdout,
		"%s restored the credentials from version %d; it was at version %d and is now at version %d\n\n",
		res.Provider.Name, res.ToVersion, res.FromVersion, res.NewVersion); err != nil {
		return err
	}
	if err := renderProviderRow(os.Stdout, &res.Provider); err != nil {
		return err
	}
	_, err = fmt.Fprintf(os.Stdout,
		"\nRestoring does not re-check the upstream: `olares-cli router provider validate %s`\n",
		res.Provider.Name)
	return err
}
