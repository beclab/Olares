package router

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// `olares-cli router provider create` — POST /console/api/providers
// `olares-cli router provider register <app>` — the same route, source=olares
//
// Creating a provider stores credentials and nothing else: Router deliberately
// attaches no models, because pulling a vendor's whole catalog onto a key the
// moment it is entered leaves no way to offer a subset. Models are a separate,
// explicit step afterwards.
//
// The credentials are encrypted before they land, and no route ever returns
// them again — the credentials-form route answers with the secrets masked.
// A rejected key is therefore something you find out from `provider validate`
// rather than from this verb, which only reports that the row was written.

type createProviderRequest struct {
	Name                 string         `json:"name"`
	ProviderType         string         `json:"provider_type"`
	BaseURL              string         `json:"base_url"`
	Credentials          map[string]any `json:"credentials"`
	ProviderDisplayTitle string         `json:"provider_display_title,omitempty"`
}

func newProviderCreateCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		output      string
		name        string
		typeName    string
		baseURL     string
		title       string
		credPairs   []string
		credJSON    string
		andValidate bool
	)
	cmd := &cobra.Command{
		Use:   "create",
		Short: "add a provider from a base URL and credentials",
		Long: `Add a provider Router can route to.

Four things are required: a name, the provider type, the base URL, and the
credentials that type asks for. Run "provider types <vendor>" first — it names
the credential fields, and they are the keys --credential takes.

Nothing is verified while writing. A base URL that does not resolve and a key
that was revoked both create a provider happily; "provider validate" is what
asks the upstream. Pass --validate to run that immediately after creating.

No models are attached. That is deliberate: a vendor catalog imported wholesale
the moment credentials are entered cannot be narrowed afterwards. Choose them
with "provider models add", or for an endpoint that publishes its own list, run
"provider sync-models".

Credentials given as --credential stay in your shell history and are visible
in the process list on a shared machine. Prefer --credentials-json -, which
reads the object from stdin.

An upstream that needs no key at all — a local OpenAI-compatible endpoint on a
trusted network — still has to say so: pass --credentials-json with an empty
object, or an empty value such as --credential api_key=.

Examples:
  olares-cli router provider create --name openai-prod --type openai \
    --base-url https://api.openai.com/v1 --credential openai_api_key=sk-...

  printf '{"api_key":"%s"}' "$ANTHROPIC_KEY" |
    olares-cli router provider create --name claude --type anthropic \
      --base-url https://api.anthropic.com --credentials-json - --validate
`,
		Args: cobra.NoArgs,
		RunE: func(c *cobra.Command, args []string) error {
			return runProviderCreate(c.Context(), f, createProviderRequest{
				Name:                 strings.TrimSpace(name),
				ProviderType:         strings.TrimSpace(typeName),
				BaseURL:              strings.TrimSpace(baseURL),
				ProviderDisplayTitle: strings.TrimSpace(title),
			}, credPairs, credJSON, andValidate, output)
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "unique name for the provider")
	cmd.Flags().StringVar(&typeName, "type", "", "provider type, as `provider types` lists it")
	cmd.Flags().StringVar(&baseURL, "base-url", "", "the upstream's API base URL")
	cmd.Flags().StringVar(&title, "title", "", "display label; defaults to the name")
	cmd.Flags().StringArrayVar(&credPairs, "credential", nil, credentialFlagUsage)
	cmd.Flags().StringVar(&credJSON, "credentials-json", "", credentialsJSONFlagUsage)
	cmd.Flags().BoolVar(&andValidate, "validate", false, "probe the upstream once the provider exists")
	addOutputFlag(cmd, &output)
	return cmd
}

func runProviderCreate(ctx context.Context, f *cmdutil.Factory, req createProviderRequest, credPairs []string, credJSON string, andValidate bool, outputRaw string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	format, err := parseFormat(outputRaw)
	if err != nil {
		return err
	}
	missing := make([]string, 0, 4)
	if req.Name == "" {
		missing = append(missing, "--name")
	}
	if req.ProviderType == "" {
		missing = append(missing, "--type")
	}
	if req.BaseURL == "" {
		missing = append(missing, "--base-url")
	}
	creds, err := collectCredentials(credPairs, credJSON)
	if err != nil {
		return err
	}
	if creds == nil {
		missing = append(missing, "--credential or --credentials-json")
	}
	if len(missing) > 0 {
		return fmt.Errorf("provider create needs %s; `olares-cli router provider types` lists the types and their credential fields",
			strings.Join(missing, ", "))
	}
	req.Credentials = creds

	pc, err := prepare(ctx, f)
	if err != nil {
		return err
	}
	var created providerRow
	if err := pc.router.doJSON(ctx, "POST", consoleAPI+"/providers", req, &created); err != nil {
		return err
	}

	if format == FormatJSON && !andValidate {
		return printJSON(os.Stdout, created)
	}
	if format == FormatTable {
		if err := renderProviderRow(os.Stdout, &created); err != nil {
			return err
		}
	}

	if !andValidate {
		_, err := fmt.Fprintf(os.Stdout,
			"\ncreated with no models attached. Next: `olares-cli router provider validate %s`, "+
				"then choose models with `olares-cli router provider models add %s ...`\n",
			created.Name, created.Name)
		return err
	}

	verdict, verr := validateProvider(ctx, pc, created.ID)
	if format == FormatJSON {
		return printJSON(os.Stdout, struct {
			Provider providerRow      `json:"provider"`
			Validate *validateVerdict `json:"validate,omitempty"`
			Error    string           `json:"validate_error,omitempty"`
		}{created, verdict, errString(verr)})
	}
	if verr != nil {
		// The provider exists either way; a probe that could not run is worth
		// reporting without unwinding the create.
		fmt.Fprintf(os.Stdout, "\ncreated, but the credential probe could not run: %v\n", verr)
		return nil
	}
	if _, err := fmt.Fprintln(os.Stdout); err != nil {
		return err
	}
	return renderValidateVerdict(os.Stdout, created.Name, verdict)
}

func errString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

// `provider register <app-name>` is the create route's other body shape: it
// resolves the provider Router already registered for an installed model
// application. Nothing is created — the background sync that watches for model
// apps owns those rows — so this is how you confirm an app has been picked up,
// and learn the provider id to inspect it by.
func newProviderRegisterCommand(f *cmdutil.Factory) *cobra.Command {
	var output string
	cmd := &cobra.Command{
		Use:   "register <app-name>",
		Short: "the provider Router registered for an installed model application",
		Long: `Resolve the provider that belongs to a model application.

Router discovers model applications on its own, on a poll of roughly half a
minute, and writes a provider row for each. This verb looks one up by
application name. It creates nothing: if the application was installed moments
ago, the answer is that Router has not seen it yet, and the fix is to wait for
the next poll rather than to add anything by hand.

The base URL, credentials and model list of such a provider are owned by the
application. They cannot be edited here, and the row disappears when the
application is uninstalled.

Example:
  olares-cli router provider register qwen3-6-27b
`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			return runProviderRegister(c.Context(), f, args[0], output)
		},
	}
	addOutputFlag(cmd, &output)
	return cmd
}

func runProviderRegister(ctx context.Context, f *cmdutil.Factory, appName, outputRaw string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	format, err := parseFormat(outputRaw)
	if err != nil {
		return err
	}
	appName = strings.TrimSpace(appName)
	if appName == "" {
		return fmt.Errorf("application name is required")
	}
	pc, err := prepare(ctx, f)
	if err != nil {
		return err
	}
	body := map[string]string{"source": "olares", "app_name": appName}
	var row providerRow
	if err := pc.router.doJSON(ctx, "POST", consoleAPI+"/providers", body, &row); err != nil {
		return err
	}
	if format == FormatJSON {
		return printJSON(os.Stdout, row)
	}
	return renderProviderRow(os.Stdout, &row)
}
