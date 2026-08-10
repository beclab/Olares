package model

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// `olares-cli model provider update <name|id>`
// PATCH /console/api/providers/:id
//
// Everything is a merge: a flag you do not pass leaves that field alone. Two
// details are worth knowing before rotating a key.
//
// Credentials merge per field, not wholesale. Sending only the field you are
// changing keeps the rest of them, because Router resolves a secret field that
// is absent to the value already stored. So rotating one key on a vendor that
// wants three does not require re-typing the other two.
//
// A rotation always bumps the credential version and writes a history row, even
// when the new value equals the old one. That history is what `provider
// rollback` restores from, and the note you attach is the only place the reason
// for a rotation is recorded.

type updateProviderRequest struct {
	Name                 *string         `json:"name,omitempty"`
	ProviderDisplayTitle *string         `json:"provider_display_title,omitempty"`
	BaseURL              *string         `json:"base_url,omitempty"`
	Status               *string         `json:"status,omitempty"`
	Credentials          *map[string]any `json:"credentials,omitempty"`
	CredentialsNote      *string         `json:"credentials_note,omitempty"`
}

func newProviderUpdateCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		output    string
		name      string
		title     string
		clearTitl bool
		baseURL   string
		status    string
		credPairs []string
		credJSON  string
		credNote  string
	)
	cmd := &cobra.Command{
		Use:   "update <name|id>",
		Short: "rename, re-point, disable, or rotate a provider's credentials",
		Long: `Change one or more fields on a provider.

A field you do not name is left as it is.

Credentials merge field by field. To rotate one key on a vendor that asks for
several, pass only that key: Router keeps the stored value for every secret
field you omit. Prefer --credentials-json - so the new secret does not end up in
your shell history.

Every credential change bumps the version and records a history entry, whether
or not the value actually differs. --credentials-note is written to that entry
and is the only record of why; "provider history" reads them back and
"provider rollback" restores from them.

--status disabled takes the provider out of routing without deleting anything,
which is the reversible way to stop traffic to a suspect upstream. Models,
quotas and defaults stay as they are, so re-enabling restores the previous
behaviour exactly.

The provider type cannot change. An upstream that speaks a different dialect is
a different provider, and creating one is safer than reinterpreting the
credentials already stored.

A provider registered for a model application cannot be edited here at all —
the application owns its configuration.

Examples:
  olares-cli model provider update openai-prod --status disabled
  printf '{"openai_api_key":"%s"}' "$NEW_KEY" |
    olares-cli model provider update openai-prod --credentials-json - \
      --credentials-note "quarterly rotation"
`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			req := updateProviderRequest{}
			flags := c.Flags()
			if flags.Changed("name") {
				v := strings.TrimSpace(name)
				req.Name = &v
			}
			if flags.Changed("title") {
				v := strings.TrimSpace(title)
				req.ProviderDisplayTitle = &v
			}
			if clearTitl {
				empty := ""
				req.ProviderDisplayTitle = &empty
			}
			if flags.Changed("base-url") {
				v := strings.TrimSpace(baseURL)
				req.BaseURL = &v
			}
			if flags.Changed("status") {
				v := strings.ToLower(strings.TrimSpace(status))
				req.Status = &v
			}
			if flags.Changed("credentials-note") {
				v := credNote
				req.CredentialsNote = &v
			}
			return runProviderUpdate(c.Context(), f, args[0], req, credPairs, credJSON, output)
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "new unique name")
	cmd.Flags().StringVar(&title, "title", "", "new display label")
	cmd.Flags().BoolVar(&clearTitl, "clear-title", false, "drop the display label so the name is shown instead")
	cmd.Flags().StringVar(&baseURL, "base-url", "", "new API base URL")
	cmd.Flags().StringVar(&status, "status", "", "active or disabled")
	cmd.Flags().StringArrayVar(&credPairs, "credential", nil, credentialFlagUsage+"; omitted secret fields keep their stored value")
	cmd.Flags().StringVar(&credJSON, "credentials-json", "", credentialsJSONFlagUsage)
	cmd.Flags().StringVar(&credNote, "credentials-note", "", "reason for the rotation, recorded in credential history")
	addOutputFlag(cmd, &output)
	return cmd
}

func runProviderUpdate(ctx context.Context, f *cmdutil.Factory, ref string, req updateProviderRequest, credPairs []string, credJSON, outputRaw string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	format, err := parseFormat(outputRaw)
	if err != nil {
		return err
	}
	if req.Status != nil && *req.Status != "active" && *req.Status != "disabled" {
		return fmt.Errorf("--status must be active or disabled, not %q", *req.Status)
	}
	creds, err := collectCredentials(credPairs, credJSON)
	if err != nil {
		return err
	}
	if creds != nil {
		req.Credentials = &creds
	}
	if isEmptyUpdate(req) {
		return fmt.Errorf("nothing to change; pass at least one of --name, --title, --clear-title, --base-url, --status, --credential, --credentials-json")
	}
	if req.Credentials == nil && req.CredentialsNote != nil {
		return fmt.Errorf("--credentials-note only means something alongside a credential change")
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
		return marketOwnedErr(found, "update")
	}

	var updated providerRow
	path := consoleAPI + "/providers/" + url.PathEscape(found.ID)
	if err := pc.router.doJSON(ctx, "PATCH", path, req, &updated); err != nil {
		return err
	}
	if format == FormatJSON {
		return printJSON(os.Stdout, updated)
	}
	if err := renderProviderRow(os.Stdout, &updated); err != nil {
		return err
	}
	if req.Credentials != nil {
		_, err := fmt.Fprintf(os.Stdout,
			"\ncredentials now at version %d. Confirm the upstream accepts them: "+
				"`olares-cli model provider validate %s`\n",
			updated.CredentialsVersion, updated.Name)
		return err
	}
	return nil
}

// isEmptyUpdate mirrors what Router treats as an empty patch, so an invocation
// with no effective flags is refused here rather than travelling to be refused
// there.
func isEmptyUpdate(req updateProviderRequest) bool {
	raw, err := json.Marshal(req)
	if err != nil {
		return false
	}
	var fields map[string]any
	if err := json.Unmarshal(raw, &fields); err != nil {
		return false
	}
	delete(fields, "credentials_note")
	return len(fields) == 0
}
