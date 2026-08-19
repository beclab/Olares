package router

import (
	"context"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// `olares-cli router provider validate <name|id>`
// POST /console/api/providers/:id/validate
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
	var output string
	cmd := &cobra.Command{
		Use:   "validate <name|id>",
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

Example:
  olares-cli router provider validate openai-prod
`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			return runProviderValidate(c.Context(), f, args[0], output)
		},
	}
	addOutputFlag(cmd, &output)
	return cmd
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
