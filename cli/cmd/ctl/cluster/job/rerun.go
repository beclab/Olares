package job

import (
	"context"
	"fmt"
	"net/url"
	"os"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/cmd/ctl/cluster/internal/clusteropts"
	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// NewRerunCommand: `olares-cli cluster job rerun <ns/name | name>
// [-n NS] [--yes]`.
//
// Calls SPA's jobRerun
// (apps/packages/app/src/apps/controlPanelCommon/network/index.ts):
// `POST /kapis/operations.kubesphere.io/v1alpha2/namespaces/<ns>/
// jobs/<name>?action=rerun&resourceVersion=<rv>` with no body.
//
// Two-step (matches the SPA's JobsDetails.vue toolbar exactly):
//
//  1. GET the Job to read .metadata.resourceVersion. The operations
//     API rejects rerun without a current RV — passing a stale one
//     would lose the call.
//  2. POST with action=rerun&resourceVersion=<rv>.
//
// rerun spawns a new pod execution server-side; we wrap it in
// ConfirmDestructive (it's mutating + reversible only by deleting the
// new attempt's pods) so scripts must opt in via --yes.
func NewRerunCommand(f *cmdutil.Factory) *cobra.Command {
	o := clusteropts.NewClusterOptions(f)
	var (
		namespace string
		assumeYes bool
	)
	cmd := &cobra.Command{
		Use:   "rerun <ns/name | name>",
		Short: "rerun one Job (KubeSphere operations action; spawns a fresh attempt)",
		Long: `Rerun one Job by triggering KubeSphere's "action=rerun" operations
endpoint. The server creates a new Pod (and updates the Job status
accordingly).

Two-step flow (matches the SPA exactly):
  1. GET ` + "`/apis/batch/v1/namespaces/<ns>/jobs/<name>`" + ` to read the
     current resourceVersion (required by the operations API).
  2. POST ` + "`/kapis/operations.kubesphere.io/v1alpha2/namespaces/<ns>/jobs/<name>?action=rerun&resourceVersion=<rv>`" + ` with no body.

This is mutating: a new Pod is launched and the Job's status updates.
Pass --yes to skip the confirmation prompt for scripted use.
`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			ns, name, err := splitNsName(namespace, args[0])
			if err != nil {
				return err
			}
			return runRerun(c.Context(), o, ns, name, assumeYes)
		},
	}
	cmd.Flags().StringVarP(&namespace, "namespace", "n", "", "namespace (required when the positional argument is a bare name)")
	cmd.Flags().BoolVarP(&assumeYes, "yes", "y", false, "skip the confirmation prompt")
	o.AddOutputFlags(cmd)
	return cmd
}

// rerunResult is the JSON-mode shape emitted on success. The
// operations API itself returns an opaque body (KubeSphere doesn't
// document its shape and the SPA discards it), so we synthesize a
// stable summary callers can rely on.
type rerunResult struct {
	Operation       string `json:"operation"`
	Namespace       string `json:"namespace"`
	Job             string `json:"job"`
	ResourceVersion string `json:"resourceVersion"`
	Status          string `json:"status"`
}

func runRerun(ctx context.Context, o *clusteropts.ClusterOptions, namespace, name string, assumeYes bool) error {
	if ctx == nil {
		ctx = context.Background()
	}
	j, err := Get(ctx, o, namespace, name)
	if err != nil {
		return err
	}
	rv := j.Metadata.ResourceVersion
	if rv == "" {
		return fmt.Errorf("job %s/%s has no metadata.resourceVersion — cannot rerun without it", namespace, name)
	}

	if err := clusteropts.ConfirmDestructive(os.Stderr, os.Stdin,
		fmt.Sprintf("Rerun job %s/%s (resourceVersion=%s)? A new pod attempt will be launched", namespace, name, rv),
		assumeYes); err != nil {
		return err
	}

	client, err := o.Prepare()
	if err != nil {
		return err
	}
	q := url.Values{}
	q.Set("action", "rerun")
	q.Set("resourceVersion", rv)
	path := fmt.Sprintf("/kapis/operations.kubesphere.io/v1alpha2/namespaces/%s/jobs/%s?%s",
		url.PathEscape(namespace), url.PathEscape(name), q.Encode())

	// Body is intentionally nil — the SPA's jobRerun sends no body
	// (see network/index.ts). DoJSON encodes nil body as "no body"
	// and leaves Content-Type unset, which is what the operations
	// API expects for action triggers.
	if err := client.DoJSON(ctx, "POST", path, nil, nil); err != nil {
		return fmt.Errorf("rerun job %s/%s: %w", namespace, name, err)
	}

	result := rerunResult{
		Operation:       "rerun",
		Namespace:       namespace,
		Job:             name,
		ResourceVersion: rv,
		Status:          "accepted",
	}
	if o.IsJSON() {
		return o.PrintJSON(result)
	}
	if !o.Quiet {
		fmt.Fprintf(os.Stdout, "rerun accepted for job %s/%s (resourceVersion=%s)\n", namespace, name, rv)
	}
	return nil
}
