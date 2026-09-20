package router

import (
	"context"
	"fmt"
	"os"
	"sort"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// `olares-cli router provider credentials <name|id>`
// GET /console/api/providers/:id/credentials-form
//
// Which credential fields a provider has stored, with the secrets masked. What
// this answers is "what is configured here" — field names, and the plain values
// of the fields that are not secret, such as an endpoint or an organization id.
// It cannot answer "what is the key", by design: no Router route returns a
// stored secret, and the masked marker is also the value you send back
// unchanged when patching a different field.

// hiddenSentinel is the marker Router substitutes for a secret. Recognising it
// lets the output say "stored, not shown", which is a different statement from
// a field whose value happens to be that string.
const hiddenSentinel = "[__HIDDEN__]"

func newProviderCredentialsCommand(f *cmdutil.Factory) *cobra.Command {
	var output string
	cmd := &cobra.Command{
		Use:   "credentials <name|id>",
		Short: "the credential fields a provider has stored, with secrets masked",
		Long: `Show which credential fields are stored against a provider.

Secrets are never returned by Router, so they appear here as "stored, not
shown". Fields the vendor does not treat as secret — an endpoint URL, an
organization or deployment id — are shown as they are, which is often the thing
you actually want to check.

This is how to tell an empty field from an unreadable one before rotating
anything: a field listed as unset was never given a value, while one listed as
stored has one you cannot read back.

If Router cannot decrypt what it stored, the request fails rather than
answering with blanks — a provider whose credentials cannot be read needs a
rotation, not a redacted display.

Example:
  olares-cli router provider credentials openai-prod
`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			return runProviderCredentials(c.Context(), f, args[0], output)
		},
	}
	addOutputFlag(cmd, &output)
	return cmd
}

func runProviderCredentials(ctx context.Context, f *cmdutil.Factory, ref, outputRaw string) error {
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
		Credentials map[string]any `json:"credentials"`
	}
	path := epProviderCredentialsForm(found.ID)
	if err := pc.router.doJSON(ctx, "GET", path, nil, &env); err != nil {
		return err
	}
	if format == FormatJSON {
		return printJSON(os.Stdout, env)
	}

	if len(env.Credentials) == 0 {
		_, err := fmt.Fprintf(os.Stdout, "%s has no credential fields stored\n", found.Name)
		return err
	}
	keys := make([]string, 0, len(env.Credentials))
	for k := range env.Credentials {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	if _, err := fmt.Fprintf(os.Stdout, "%s (credential version %d)\n\n", found.Name, found.CredentialsVersion); err != nil {
		return err
	}
	t := newTable(os.Stdout, "FIELD", "VALUE")
	for _, k := range keys {
		value := maskCredentialValue(env.Credentials[k])
		if s, ok := env.Credentials[k].(string); ok && s == hiddenSentinel {
			value = "(stored, not shown)"
		}
		t.row(k, value)
	}
	return t.flush()
}
