package node

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"sort"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/cmd/ctl/cluster/internal/clusteropts"
	"github.com/beclab/Olares/cli/pkg/clusterclient"
	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// NewGetCommand: `olares-cli cluster node get <name> [-o table|json]`.
//
// Calls SPA's getNodeDetail
// (apps/packages/app/src/apps/controlPanelCommon/network/index.ts:164):
// `/kapis/resources.kubesphere.io/v1alpha3/nodes/<node>` — returns
// the K8s native Node object (no envelope; KubeSphere's per-resource
// detail just forwards the upstream shape).
func NewGetCommand(f *cmdutil.Factory) *cobra.Command {
	o := clusteropts.NewClusterOptions(f)
	cmd := &cobra.Command{
		Use:   "get <name>",
		Short: "show one node's detail (K8s native shape)",
		Args:  cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			name := strings.TrimSpace(args[0])
			if name == "" {
				return fmt.Errorf("node name must be non-empty")
			}
			return runGet(c.Context(), o, name)
		},
	}
	o.AddOutputFlags(cmd)
	return cmd
}

func runGet(ctx context.Context, o *clusteropts.ClusterOptions, name string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	client, err := o.Prepare()
	if err != nil {
		return err
	}
	path := "/kapis/resources.kubesphere.io/v1alpha3/nodes/" + url.PathEscape(name)
	var n Node
	if err := clusterclient.GetK8sObject(ctx, client, path, &n); err != nil {
		return fmt.Errorf("get node %q: %w", name, err)
	}
	if o.IsJSON() {
		return o.PrintJSON(n)
	}
	return renderGetTable(n)
}

func renderGetTable(n Node) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	defer w.Flush()
	fmt.Fprintf(w, "Name:\t%s\n", n.Metadata.Name)
	fmt.Fprintf(w, "Status:\t%s\n", n.StatusLabel())
	fmt.Fprintf(w, "Roles:\t%s\n", n.Roles())
	fmt.Fprintf(w, "Internal IP:\t%s\n", n.InternalIP())
	fmt.Fprintf(w, "Kubelet:\t%s\n", dashIfEmpty(n.KubeletVersion()))
	fmt.Fprintf(w, "Container Runtime:\t%s\n", dashIfEmpty(n.Status.NodeInfo.ContainerRuntimeVersion))
	fmt.Fprintf(w, "OS Image:\t%s\n", dashIfEmpty(n.Status.NodeInfo.OSImage))
	fmt.Fprintf(w, "Kernel:\t%s\n", dashIfEmpty(n.Status.NodeInfo.KernelVersion))
	fmt.Fprintf(w, "Architecture:\t%s\n", dashIfEmpty(n.Status.NodeInfo.Architecture))
	fmt.Fprintf(w, "Created:\t%s\n", dashIfEmpty(n.Metadata.CreationTimestamp))
	fmt.Fprintf(w, "Age:\t%s\n", n.Age(time.Now()))
	if n.Spec.Unschedulable {
		fmt.Fprintln(w, "Schedulable:\tfalse (cordoned)")
	} else {
		fmt.Fprintln(w, "Schedulable:\ttrue")
	}

	// Capacity / Allocatable: render only the standard K8s keys to
	// keep the table dense (cpu / memory / pods / ephemeral-storage).
	wellKnown := []string{"cpu", "memory", "pods", "ephemeral-storage"}
	if len(n.Status.Capacity) > 0 {
		fmt.Fprintln(w, "Capacity:")
		for _, k := range wellKnown {
			if v, ok := n.Status.Capacity[k]; ok {
				fmt.Fprintf(w, "  %s\t%s\n", k, v)
			}
		}
	}
	if len(n.Status.Allocatable) > 0 {
		fmt.Fprintln(w, "Allocatable:")
		for _, k := range wellKnown {
			if v, ok := n.Status.Allocatable[k]; ok {
				fmt.Fprintf(w, "  %s\t%s\n", k, v)
			}
		}
	}

	if len(n.Spec.Taints) > 0 {
		fmt.Fprintln(w, "Taints:")
		for _, t := range n.Spec.Taints {
			val := t.Value
			if val == "" {
				val = "<none>"
			}
			fmt.Fprintf(w, "  %s\t%s:%s\n", t.Key, val, t.Effect)
		}
	}

	if len(n.Status.Conditions) > 0 {
		fmt.Fprintln(w, "Conditions:")
		conds := append([]NodeCondition(nil), n.Status.Conditions...)
		sort.SliceStable(conds, func(i, j int) bool { return conds[i].Type < conds[j].Type })
		for _, c := range conds {
			extra := ""
			if c.Reason != "" {
				extra = " (" + c.Reason + ")"
			}
			fmt.Fprintf(w, "  %s\t%s%s\n", c.Type, c.Status, extra)
		}
	}

	if len(n.Status.Addresses) > 0 {
		fmt.Fprintln(w, "Addresses:")
		for _, a := range n.Status.Addresses {
			fmt.Fprintf(w, "  %s\t%s\n", a.Type, a.Address)
		}
	}

	// Surface a flat label list at the end. Helpful for spotting
	// node-role.kubernetes.io/* labels that drove the ROLES column,
	// plus any zone / topology labels.
	if len(n.Metadata.Labels) > 0 {
		fmt.Fprintln(w, "Labels:")
		keys := make([]string, 0, len(n.Metadata.Labels))
		for k := range n.Metadata.Labels {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			fmt.Fprintf(w, "  %s\t%s\n", k, n.Metadata.Labels[k])
		}
	}

	return nil
}
