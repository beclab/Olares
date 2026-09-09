package router

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// `olares-cli router provider validate <name|id>`
// POST /console/api/providers/:id/validate
// POST /console/api/provider-validate       (--draft, nothing saved)
//
// Router decrypts the stored credentials and asks the upstream for its model
// list, from inside the cluster. That makes this the one check that covers the
// whole path at once: DNS, route, TLS chain, and whether the key is still
// accepted.
//
// Two outcomes are different in kind. A verdict means the probe ran and the
// upstream answered — including answering 401, which is a valid answer and
// reads as invalid credentials. A failure means the probe could not run at all,
// which is a network fact about the cluster rather than a judgement on the key.
//
// The draft form answers the same question before there is a row to ask about.
// It matters because a provider is cheap to create and expensive to have been
// wrong about: the credential history keeps every version, so finding out after
// saving leaves a row to delete and a version that was never good.

type validateVerdict struct {
	Valid           bool   `json:"valid"`
	Verdict         string `json:"verdict"`
	UpstreamStatus  int    `json:"upstream_status"`
	ProbeURL        string `json:"probe_url"`
	ModelCount      int    `json:"model_count,omitempty"`
	CheckedAt       string `json:"checked_at"`
	UpstreamMessage string `json:"upstream_message,omitempty"`
}

func newProviderValidateCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		output    string
		draft     bool
		typeName  string
		baseURL   string
		credPairs []string
		credJSON  string
	)
	cmd := &cobra.Command{
		Use:   "validate [name|id]",
		Short: "ask the upstream whether the stored credentials still work",
		Long: `Probe a provider's upstream with its stored credentials.

Router runs the probe itself, from inside the cluster, by asking the upstream
for its model list. So a pass means more than "the key is well formed": it
means this Olares can reach that upstream and the upstream accepted the key.

The verdict names what happened:

  upstream_ok             the upstream answered with a model list
  upstream_unauthorized   the key was rejected; rotate it with
                          "provider update --credential"
  upstream_4xx            the upstream refused for another reason, often a
                          base URL pointing at the wrong path
  upstream_5xx            the upstream is unwell; the key may still be fine
  upstream_non_json       something answered, but not this API — usually a
                          proxy or login page in front of the real endpoint

A probe that cannot run at all fails instead of returning a verdict, because
"unreachable from the cluster" is not a judgement about the credentials.

For a provider belonging to a model application, a failure right after install
is expected: the application is still starting.

--draft runs the same probe against credentials that have not been saved. It
takes a name for nothing and describes the upstream instead: --type, --base-url
and the credentials. Nothing is written, so a key that turns out to be wrong
leaves no provider behind to delete and no version in the credential history.
Some vendors resolve their own endpoint and need no --base-url; the refusal
says which when one is needed.

Examples:
  olares-cli router provider validate openai-prod
  printf '{"api_key":"%s"}' "$ANTHROPIC_KEY" |
    olares-cli router provider validate --draft --type anthropic \
      --base-url https://api.anthropic.com --credentials-json -
`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			if draft {
				if len(args) > 0 {
					return fmt.Errorf("--draft probes an upstream that has no provider yet, " +
						"so there is no name to give; describe it with --type and --base-url")
				}
				return runProviderValidateDraft(c.Context(), f, draftProbe{
					ProviderType: strings.TrimSpace(typeName),
					BaseURL:      strings.TrimSpace(baseURL),
				}, credPairs, credJSON, output)
			}
			if len(args) == 0 {
				return fmt.Errorf("name the provider to validate, or describe an unsaved one " +
					"with --draft --type <type>")
			}
			if typeName != "" || baseURL != "" || len(credPairs) > 0 || credJSON != "" {
				return fmt.Errorf("a saved provider is probed with what it stores; " +
					"--type, --base-url and the credential flags belong to --draft")
			}
			return runProviderValidate(c.Context(), f, args[0], output)
		},
	}
	cmd.Flags().BoolVar(&draft, "draft", false,
		"probe credentials that have not been saved, instead of a provider")
	cmd.Flags().StringVar(&typeName, "type", "", "provider type, as `provider types` lists it (--draft only)")
	cmd.Flags().StringVar(&baseURL, "base-url", "", "the upstream's API base URL (--draft only)")
	cmd.Flags().StringArrayVar(&credPairs, "credential", nil, credentialFlagUsage)
	cmd.Flags().StringVar(&credJSON, "credentials-json", "", credentialsJSONFlagUsage)
	addOutputFlag(cmd, &output)
	return cmd
}

// draftProbe is what the probe needs and nothing else. A name, a title and a
// source describe a row, and --draft writes no row — carrying them would
// invite a caller to believe the probe honoured them.
type draftProbe struct {
	ProviderType string         `json:"provider_type"`
	BaseURL      string         `json:"base_url,omitempty"`
	Credentials  map[string]any `json:"credentials"`
}

func runProviderValidateDraft(ctx context.Context, f *cmdutil.Factory, probe draftProbe,
	credPairs []string, credJSON, outputRaw string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	format, err := parseFormat(outputRaw)
	if err != nil {
		return err
	}
	if probe.ProviderType == "" {
		return fmt.Errorf("--draft needs --type; `olares-cli router provider types` lists them " +
			"with the credential fields each one takes")
	}
	creds, err := collectCredentials(credPairs, credJSON)
	if err != nil {
		return err
	}
	// An absent credentials object is not an error. A local Ollama, an Olares
	// application and a self-hosted SearXNG have nothing to send, and Router
	// decides per provider type whether that is a problem.
	probe.Credentials = creds
	if probe.Credentials == nil {
		probe.Credentials = map[string]any{}
	}
	pc, err := prepare(ctx, f)
	if err != nil {
		return err
	}
	var verdict validateVerdict
	if err := pc.router.doJSON(ctx, "POST", epProviderValidateDraft, probe, &verdict); err != nil {
		return err
	}
	if format == FormatJSON {
		return printJSON(os.Stdout, verdict)
	}
	if err := renderValidateVerdict(os.Stdout, "(not saved: "+probe.ProviderType+")", &verdict); err != nil {
		return err
	}
	if !verdict.Valid {
		return nil
	}
	_, err = fmt.Fprintf(os.Stdout, "\nNothing was saved. `olares-cli router provider create "+
		"--name <name> --type %s` with the same credentials keeps it.\n", probe.ProviderType)
	return err
}

func runProviderValidate(ctx context.Context, f *cmdutil.Factory, ref, outputRaw string) error {
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
	verdict, err := validateProvider(ctx, pc, found.ID)
	if err != nil {
		return err
	}
	if format == FormatJSON {
		return printJSON(os.Stdout, verdict)
	}
	return renderValidateVerdict(os.Stdout, found.Name, verdict)
}

func validateProvider(ctx context.Context, pc *preparedClient, id string) (*validateVerdict, error) {
	var v validateVerdict
	path := epProviderValidate(id)
	if err := pc.router.doJSON(ctx, "POST", path, nil, &v); err != nil {
		return nil, err
	}
	return &v, nil
}

func renderValidateVerdict(w io.Writer, providerName string, v *validateVerdict) error {
	t := newTable(w)
	t.row("PROVIDER", nonEmpty(providerName))
	t.row("CREDENTIALS WORK", boolStr(v.Valid))
	t.row("VERDICT", nonEmpty(v.Verdict))
	t.row("UPSTREAM STATUS", intOrDash(v.UpstreamStatus))
	t.row("PROBED", nonEmpty(v.ProbeURL))
	t.row("MODELS OFFERED", intOrDash(v.ModelCount))
	t.row("CHECKED AT", nonEmpty(v.CheckedAt))
	if v.UpstreamMessage != "" {
		t.row("UPSTREAM SAID", truncate(v.UpstreamMessage, 200))
	}
	return t.flush()
}
