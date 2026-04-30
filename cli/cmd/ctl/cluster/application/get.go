package application

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

// NewGetCommand: `olares-cli cluster application get <namespace>
// [-o table|json]`.
//
// Hits `/api/v1/namespaces/<ns>` (the same K8s native namespace
// detail SPA's getNamespacesDetail uses, see
// apps/packages/app/src/apps/controlPanelCommon/network/index.ts:283).
//
// In `-o table` mode we print a vertical key/value summary
// (workspace via labels, status, age, label list); in `-o json`
// the response is forwarded verbatim. Pivot from here to
// `cluster application workloads <ns>` / `cluster application pods
// <ns>` for the contents.
func NewGetCommand(f *cmdutil.Factory) *cobra.Command {
	o := clusteropts.NewClusterOptions(f)
	cmd := &cobra.Command{
		Use:   "get <namespace>",
		Short: "show one application space's K8s Namespace detail",
		Long: `Show the K8s Namespace detail for one application space.

Identity is a single positional <namespace>. Server-side scope
applies — a namespace exists but isn't visible to your token will
return HTTP 404 (the standard kube-apiserver "not found" path).
`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			ns := strings.TrimSpace(args[0])
			if ns == "" {
				return fmt.Errorf("namespace must be non-empty")
			}
			return runGet(c.Context(), o, ns)
		},
	}
	o.AddOutputFlags(cmd)
	return cmd
}

// NamespaceDetail is the K8s native Namespace shape (subset of
// corev1.Namespace) we render in `cluster application get`. status.phase
// covers the Active / Terminating distinction; labels carry KubeSphere's
// workspace + workload-type metadata.
type NamespaceDetail struct {
	Kind       string `json:"kind,omitempty"`
	APIVersion string `json:"apiVersion,omitempty"`
	Metadata   struct {
		Name              string            `json:"name"`
		UID               string            `json:"uid,omitempty"`
		CreationTimestamp string            `json:"creationTimestamp,omitempty"`
		Labels            map[string]string `json:"labels,omitempty"`
		Annotations       map[string]string `json:"annotations,omitempty"`
	} `json:"metadata"`
	Status struct {
		Phase string `json:"phase,omitempty"`
	} `json:"status,omitempty"`
}

func runGet(ctx context.Context, o *clusteropts.ClusterOptions, namespace string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	client, err := o.Prepare()
	if err != nil {
		return err
	}
	path := "/api/v1/namespaces/" + url.PathEscape(namespace)
	var ns NamespaceDetail
	if err := clusterclient.GetK8sObject(ctx, client, path, &ns); err != nil {
		return fmt.Errorf("get application space %q: %w", namespace, err)
	}
	if o.IsJSON() {
		return o.PrintJSON(ns)
	}
	return renderGetTable(ns)
}

func renderGetTable(ns NamespaceDetail) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	defer w.Flush()
	fmt.Fprintf(w, "Name:\t%s\n", ns.Metadata.Name)
	fmt.Fprintf(w, "Kind:\t%s\n", clusteropts.DashIfEmpty(ns.Kind))
	fmt.Fprintf(w, "Phase:\t%s\n", clusteropts.DashIfEmpty(ns.Status.Phase))
	if ws, ok := ns.Metadata.Labels["kubesphere.io/workspace"]; ok && ws != "" {
		fmt.Fprintf(w, "Workspace:\t%s\n", ws)
	}
	if alias, ok := ns.Metadata.Annotations["kubesphere.io/alias-name"]; ok && alias != "" {
		fmt.Fprintf(w, "Alias:\t%s\n", alias)
	}
	if creator, ok := ns.Metadata.Annotations["kubesphere.io/creator"]; ok && creator != "" {
		fmt.Fprintf(w, "Creator:\t%s\n", creator)
	}
	fmt.Fprintf(w, "Created:\t%s\n", clusteropts.DashIfEmpty(ns.Metadata.CreationTimestamp))
	fmt.Fprintf(w, "Age:\t%s\n", clusteropts.Age(ns.Metadata.CreationTimestamp, time.Now()))
	if ns.Metadata.UID != "" {
		fmt.Fprintf(w, "UID:\t%s\n", ns.Metadata.UID)
	}

	// Surface a flat list of labels to match `kubectl describe namespace`
	// behavior — useful for spotting bytetrade.io/* / kubesphere.io/*
	// annotations that drive UI-side classification. Skip `kubesphere.io/
	// workspace` since we already printed it above.
	if len(ns.Metadata.Labels) > 0 {
		var keys []string
		for k := range ns.Metadata.Labels {
			if k == "kubesphere.io/workspace" {
				continue
			}
			keys = append(keys, k)
		}
		sort.Strings(keys)
		if len(keys) > 0 {
			fmt.Fprintln(w, "Labels:")
			for _, k := range keys {
				fmt.Fprintf(w, "  %s\t%s\n", k, ns.Metadata.Labels[k])
			}
		}
	}
	return nil
}
