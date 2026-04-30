package job

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/cmd/ctl/cluster/internal/clusteropts"
	"github.com/beclab/Olares/cli/pkg/clusterclient"
	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// NewGetCommand: `olares-cli cluster job get <ns/name | name> [-n NS]
// [-o table|json]`.
//
// Calls SPA's getCornJobsDetail
// (apps/packages/app/src/apps/controlPanelCommon/network/index.ts):
// `/apis/batch/v1/namespaces/<ns>/jobs/<name>` — K8s native.
//
// In table mode renders a vertical key/value summary plus active /
// succeeded / failed counts and Conditions block. JSON forwards the
// typed view; for byte-perfect round-trip use `cluster job yaml`.
func NewGetCommand(f *cmdutil.Factory) *cobra.Command {
	o := clusteropts.NewClusterOptions(f)
	var namespace string
	cmd := &cobra.Command{
		Use:   "get <ns/name | name>",
		Short: "show one Job's details (K8s native shape)",
		Long: `Show one Job's full detail.

Identity may be passed as a single "<namespace>/<name>" positional or
as a bare "<name>" with -n <namespace>. Without -n the positional
form is required so we don't guess a namespace.

In table mode, the output is a vertical key/value summary including
COMPLETIONS / STATUS / DURATION / Conditions; in JSON mode the typed
view is forwarded. Use ` + "`cluster job yaml`" + ` for the byte-perfect
K8s native YAML.
`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			ns, name, err := clusteropts.SplitNsName(namespace, args[0])
			if err != nil {
				return err
			}
			return runGet(c.Context(), o, ns, name)
		},
	}
	cmd.Flags().StringVarP(&namespace, "namespace", "n", "", "namespace (required when the positional argument is a bare name)")
	o.AddOutputFlags(cmd)
	return cmd
}

// Get is the exported single-Job fetcher used by sibling verbs
// (job pods needs the .uid for the controller-uid label selector;
// job rerun needs the .resourceVersion the operations API requires).
// Returns the typed Job; renderers stay in this package.
func Get(ctx context.Context, o *clusteropts.ClusterOptions, namespace, name string) (*Job, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	client, err := o.Prepare()
	if err != nil {
		return nil, err
	}
	path := buildGetPath(namespace, name)
	var j Job
	if err := clusterclient.GetK8sObject(ctx, client, path, &j); err != nil {
		return nil, fmt.Errorf("get job %s/%s: %w", namespace, name, err)
	}
	if j.Kind == "" {
		j.Kind = "Job"
	}
	return &j, nil
}

// buildGetPath is exported (lowercase) so yaml.go can share without
// reaching into this file — the path template is the same across get
// and yaml since both verbs hit the same endpoint.
func buildGetPath(namespace, name string) string {
	return fmt.Sprintf("/apis/batch/v1/namespaces/%s/jobs/%s",
		url.PathEscape(namespace), url.PathEscape(name))
}

func runGet(ctx context.Context, o *clusteropts.ClusterOptions, namespace, name string) error {
	j, err := Get(ctx, o, namespace, name)
	if err != nil {
		return err
	}
	if o.IsJSON() {
		return o.PrintJSON(*j)
	}
	return renderGetTable(*j)
}

func renderGetTable(j Job) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	defer w.Flush()
	now := time.Now()

	fmt.Fprintf(w, "Name:\t%s\n", j.Metadata.Name)
	fmt.Fprintf(w, "Namespace:\t%s\n", clusteropts.DashIfEmpty(j.Metadata.Namespace))
	fmt.Fprintf(w, "Status:\t%s\n", j.status())
	fmt.Fprintf(w, "Completions:\t%s\n", j.completionsLabel())
	if j.Spec.Parallelism != nil {
		fmt.Fprintf(w, "Parallelism:\t%d\n", *j.Spec.Parallelism)
	}
	if j.Spec.BackoffLimit != nil {
		fmt.Fprintf(w, "Backoff Limit:\t%d\n", *j.Spec.BackoffLimit)
	}
	fmt.Fprintf(w, "Active:\t%d\n", j.Status.Active)
	fmt.Fprintf(w, "Succeeded:\t%d\n", j.Status.Succeeded)
	fmt.Fprintf(w, "Failed:\t%d\n", j.Status.Failed)
	if j.Status.StartTime != "" {
		fmt.Fprintf(w, "Start Time:\t%s\n", j.Status.StartTime)
	}
	if j.Status.CompletionTime != "" {
		fmt.Fprintf(w, "Completion Time:\t%s\n", j.Status.CompletionTime)
	}
	fmt.Fprintf(w, "Duration:\t%s\n", j.duration(now))
	fmt.Fprintf(w, "Created:\t%s\n", clusteropts.DashIfEmpty(j.Metadata.CreationTimestamp))
	fmt.Fprintf(w, "Age:\t%s\n", j.age(now))
	if parent := j.parentCronJob(); parent != "" {
		fmt.Fprintf(w, "Controlled By:\tCronJob/%s\n", parent)
	}

	if len(j.Status.Conditions) > 0 {
		var lines []string
		for _, c := range j.Status.Conditions {
			s := c.Type + "=" + c.Status
			if c.Reason != "" {
				s += " (" + c.Reason + ")"
			}
			lines = append(lines, s)
		}
		fmt.Fprintf(w, "Conditions:\t%s\n", strings.Join(lines, ", "))
	}
	return nil
}
