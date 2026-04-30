package namespace

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

// NewGetCommand: `olares-cli cluster namespace get <name>
// [-o table|json]`. Calls `/api/v1/namespaces/<ns>` (K8s native).
//
// In `-o table` we print the kubectl-shaped vertical summary plus a
// labels block; in `-o json` the K8s native response is forwarded
// verbatim.
//
// Compared to ` + "`cluster application get`" + ` (which renders the same
// underlying object): this verb leans into the K8s framing —
// status.phase, all labels including bytetrade/kubesphere ones —
// rather than the SPA's workspace-first framing. They are
// complementary; pick whichever matches the user's mental model.
func NewGetCommand(f *cmdutil.Factory) *cobra.Command {
	o := clusteropts.NewClusterOptions(f)
	cmd := &cobra.Command{
		Use:   "get <name>",
		Short: "show one K8s namespace's detail (K8s native shape)",
		Args:  cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			name := strings.TrimSpace(args[0])
			if name == "" {
				return fmt.Errorf("namespace name must be non-empty")
			}
			return runGet(c.Context(), o, name)
		},
	}
	o.AddOutputFlags(cmd)
	return cmd
}

// nsDetail is the per-namespace K8s native shape we render.
type nsDetail struct {
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

func runGet(ctx context.Context, o *clusteropts.ClusterOptions, name string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	client, err := o.Prepare()
	if err != nil {
		return err
	}
	path := "/api/v1/namespaces/" + url.PathEscape(name)
	var ns nsDetail
	if err := clusterclient.GetK8sObject(ctx, client, path, &ns); err != nil {
		return fmt.Errorf("get namespace %q: %w", name, err)
	}
	if o.IsJSON() {
		return o.PrintJSON(ns)
	}
	return renderGetTable(ns)
}

func renderGetTable(ns nsDetail) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	defer w.Flush()
	fmt.Fprintf(w, "Name:\t%s\n", ns.Metadata.Name)
	fmt.Fprintf(w, "Kind:\t%s\n", clusteropts.DashIfEmpty(ns.Kind))
	fmt.Fprintf(w, "Phase:\t%s\n", clusteropts.DashIfEmpty(ns.Status.Phase))
	fmt.Fprintf(w, "Created:\t%s\n", clusteropts.DashIfEmpty(ns.Metadata.CreationTimestamp))
	fmt.Fprintf(w, "Age:\t%s\n", clusteropts.Age(ns.Metadata.CreationTimestamp, time.Now()))
	if ns.Metadata.UID != "" {
		fmt.Fprintf(w, "UID:\t%s\n", ns.Metadata.UID)
	}
	if len(ns.Metadata.Labels) > 0 {
		fmt.Fprintln(w, "Labels:")
		keys := make([]string, 0, len(ns.Metadata.Labels))
		for k := range ns.Metadata.Labels {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			fmt.Fprintf(w, "  %s\t%s\n", k, ns.Metadata.Labels[k])
		}
	}
	if len(ns.Metadata.Annotations) > 0 {
		fmt.Fprintln(w, "Annotations:")
		keys := make([]string, 0, len(ns.Metadata.Annotations))
		for k := range ns.Metadata.Annotations {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			fmt.Fprintf(w, "  %s\t%s\n", k, ns.Metadata.Annotations[k])
		}
	}
	return nil
}
