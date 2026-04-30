package workload

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

// NewGetCommand: `olares-cli cluster workload get <ns/name | name>
// [-n NS] --kind <kind> [-o table|json]`.
//
// Calls SPA's getWorkloadsControler
// (apps/packages/app/src/apps/controlPanelCommon/network/index.ts:384):
// `/apis/apps/v1/namespaces/<ns>/<kind>/<name>` — K8s native.
//
// --kind is REQUIRED here (cannot be "all"); identity is fully
// qualified by namespace + kind + name. Accept the same input shapes
// as `cluster pod get`: positional "<ns>/<name>" or "-n <ns>" + bare
// "<name>".
func NewGetCommand(f *cmdutil.Factory) *cobra.Command {
	o := clusteropts.NewClusterOptions(f)
	var (
		namespace string
		kindRaw   string
	)
	cmd := &cobra.Command{
		Use:   "get <ns/name | name>",
		Short: "show one workload's details (K8s native shape)",
		Long: `Show one workload's full detail.

--kind is required (one of: deployment, statefulset, daemonset).
Identity may be passed as a single "<namespace>/<name>" positional
or as a bare "<name>" with -n <namespace>.

Output:
  table   vertical key/value summary plus a CONDITIONS line and an
          UpdateStrategy line; pivot from here to "cluster pod list -l
          <selector>" using the SELECTOR row to find the controlled pods.
  json    the K8s native response forwarded verbatim.
`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			plural, err := NormalizeKind(kindRaw)
			if err != nil {
				return err
			}
			if plural == KindAll {
				return fmt.Errorf("--kind must be one of: deployment, statefulset, daemonset (not %q)", kindRaw)
			}
			ns, name, err := splitNsName(namespace, args[0])
			if err != nil {
				return err
			}
			return runGet(c.Context(), o, ns, name, plural)
		},
	}
	cmd.Flags().StringVarP(&namespace, "namespace", "n", "", "namespace (required when the positional argument is a bare name)")
	cmd.Flags().StringVar(&kindRaw, "kind", "", "workload kind: deployment | statefulset | daemonset (REQUIRED)")
	o.AddOutputFlags(cmd)
	return cmd
}

// splitNsName accepts either "ns/name" (no -n) or "name" (with -n).
// Mirrors cluster/pod/get.go::splitNsName so the two `get` verbs
// share the same argument grammar.
func splitNsName(nsFlag, arg string) (string, string, error) {
	if strings.Contains(arg, "/") {
		parts := strings.SplitN(arg, "/", 2)
		if parts[0] == "" || parts[1] == "" {
			return "", "", fmt.Errorf("invalid <namespace>/<name>: %q", arg)
		}
		if nsFlag != "" && nsFlag != parts[0] {
			return "", "", fmt.Errorf("argument namespace %q conflicts with --namespace %q", parts[0], nsFlag)
		}
		return parts[0], parts[1], nil
	}
	if nsFlag == "" {
		return "", "", fmt.Errorf("namespace required: pass --namespace or use <namespace>/<name>")
	}
	return nsFlag, arg, nil
}

func runGet(ctx context.Context, o *clusteropts.ClusterOptions, namespace, name, kindPlural string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	client, err := o.Prepare()
	if err != nil {
		return err
	}
	path := buildGetPath(namespace, kindPlural, name)
	var w Workload
	if err := clusterclient.GetK8sObject(ctx, client, path, &w); err != nil {
		return fmt.Errorf("get %s %s/%s: %w", SingularKind(kindPlural), namespace, name, err)
	}
	if w.Kind == "" {
		w.Kind = SingularKind(kindPlural)
	}
	if o.IsJSON() {
		return o.PrintJSON(w)
	}
	return renderGetTable(w, kindPlural)
}

// buildGetPath is exported for `cluster workload yaml` to share.
func buildGetPath(namespace, kindPlural, name string) string {
	return fmt.Sprintf("/apis/apps/v1/namespaces/%s/%s/%s",
		url.PathEscape(namespace), kindPlural, url.PathEscape(name))
}

func renderGetTable(w Workload, kindPlural string) error {
	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	defer tw.Flush()
	fmt.Fprintf(tw, "Name:\t%s\n", w.Metadata.Name)
	fmt.Fprintf(tw, "Namespace:\t%s\n", dashIfEmpty(w.Metadata.Namespace))
	fmt.Fprintf(tw, "Kind:\t%s\n", dashIfEmpty(w.Kind))
	fmt.Fprintf(tw, "API Version:\t%s\n", dashIfEmpty(w.APIVersion))
	fmt.Fprintf(tw, "Ready:\t%s\n", w.Ready(kindPlural))
	avail := w.Available(kindPlural)
	if w.RolloutInProgress() {
		avail += " (rollout in progress)"
	}
	fmt.Fprintf(tw, "Availability:\t%s\n", avail)
	fmt.Fprintf(tw, "Update Strategy:\t%s\n", w.UpdateStrategyLabel(kindPlural))
	if kindPlural == "statefulsets" && w.Spec.ServiceName != "" {
		fmt.Fprintf(tw, "Service Name:\t%s\n", w.Spec.ServiceName)
	}
	if len(w.Spec.Selector.MatchLabels) > 0 {
		var pairs []string
		for k, v := range w.Spec.Selector.MatchLabels {
			pairs = append(pairs, fmt.Sprintf("%s=%s", k, v))
		}
		// stable order so repeated runs diff cleanly
		sortStrings(pairs)
		fmt.Fprintf(tw, "Selector:\t%s\n", strings.Join(pairs, ","))
	}
	fmt.Fprintf(tw, "Created:\t%s\n", dashIfEmpty(w.Metadata.CreationTimestamp))
	fmt.Fprintf(tw, "Age:\t%s\n", w.Age(time.Now()))
	if w.Metadata.Generation > 0 {
		fmt.Fprintf(tw, "Generation:\t%d (observed: %d)\n",
			w.Metadata.Generation, w.Status.ObservedGeneration)
	}
	return nil
}

// sortStrings is a tiny helper to keep stdlib sort import private to
// this file and its callers (avoids adding sort to types.go).
func sortStrings(ss []string) {
	for i := 1; i < len(ss); i++ {
		for j := i; j > 0 && ss[j] < ss[j-1]; j-- {
			ss[j], ss[j-1] = ss[j-1], ss[j]
		}
	}
}
