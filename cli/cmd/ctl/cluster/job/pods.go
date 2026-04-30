package job

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/cmd/ctl/cluster/internal/clusteropts"
	"github.com/beclab/Olares/cli/cmd/ctl/cluster/pod"
	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// NewPodsCommand: `olares-cli cluster job pods <ns/name | name>
// [-n NS] [-l ...] [--field-selector ...] [--limit N]`.
//
// Two-step: GET the Job to read its .metadata.uid, then list pods
// scoped server-side via labelSelector=controller-uid=<uid>. Mirrors
// the SPA's "tree → pods under a Job" lazy-load
// (apps/.../controlHub/pages/Jobs/IndexPage.vue) which uses the same
// `controller-uid` selector trick.
//
// We delegate to pod.RunList for the actual fetch + render so output
// stays bit-identical to `cluster pod list -l controller-uid=<uid>
// -n <ns>`. --label / --field-selector are appended to the
// controller-uid clause so users can further filter on top.
func NewPodsCommand(f *cmdutil.Factory) *cobra.Command {
	o := clusteropts.NewClusterOptions(f)
	var (
		namespace     string
		labelSelector string
		fieldSelector string
		limit         int
	)
	cmd := &cobra.Command{
		Use:   "pods <ns/name | name>",
		Short: "list pods controlled by one Job (alias for `cluster pod list -l controller-uid=<uid>`)",
		Long: `List pods controlled by one Job.

Two-step: GET the Job to read its UID, then list pods server-side via
labelSelector=controller-uid=<uid>. Mirrors the SPA's "Jobs tree →
pods under a Job" lazy-load.

--label and --field-selector are appended to the controller-uid clause
so additional filtering happens server-side as well; the CLI never
filters or scopes pods locally.
`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			ns, name, err := splitNsName(namespace, args[0])
			if err != nil {
				return err
			}
			return runPods(c.Context(), o, ns, name, labelSelector, fieldSelector, limit)
		},
	}
	cmd.Flags().StringVarP(&namespace, "namespace", "n", "", "namespace (required when the positional argument is a bare name)")
	cmd.Flags().StringVarP(&labelSelector, "label", "l", "", "additional label selector to filter pods (K8s syntax; ANDed with controller-uid=<uid>)")
	cmd.Flags().StringVar(&fieldSelector, "field-selector", "", "field selector to filter pods (K8s syntax)")
	cmd.Flags().IntVar(&limit, "limit", 100, "max items to fetch in one request (server-side cap)")
	o.AddOutputFlags(cmd)
	return cmd
}

func runPods(ctx context.Context, o *clusteropts.ClusterOptions, namespace, name, extraLabel, fieldSelector string, limit int) error {
	if ctx == nil {
		ctx = context.Background()
	}
	j, err := Get(ctx, o, namespace, name)
	if err != nil {
		return err
	}
	if j.Metadata.UID == "" {
		return fmt.Errorf("job %s/%s has no metadata.uid — server response missing the field", namespace, name)
	}

	// Compose `controller-uid=<uid>` with any user-supplied --label.
	// K8s label selectors AND comma-separated clauses, so `,` is the
	// safe join character — same convention kubectl uses internally
	// when it merges multiple --selector arguments.
	selector := "controller-uid=" + j.Metadata.UID
	if extraLabel != "" {
		selector += "," + extraLabel
	}
	return pod.RunList(ctx, o, namespace, selector, fieldSelector, limit)
}
