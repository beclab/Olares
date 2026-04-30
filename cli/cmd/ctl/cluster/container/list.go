package container

import (
	"context"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/cmd/ctl/cluster/internal/clusteropts"
	"github.com/beclab/Olares/cli/cmd/ctl/cluster/pod"
	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// NewListCommand: `olares-cli cluster container list <ns/pod | name>
// [-n NS] [-o table|json]`.
//
// Renders one row per spec.containers[*] in the target pod. Re-uses
// `pod.Get` so the HTTP surface and error-handling story stay
// identical to `cluster pod get`.
func NewListCommand(f *cmdutil.Factory) *cobra.Command {
	o := clusteropts.NewClusterOptions(f)
	var namespace string
	cmd := &cobra.Command{
		Use:   "list <ns/pod | pod>",
		Short: "list containers inside one pod (image / state / ports)",
		Long: `List containers inside one pod.

Identity follows the same "<namespace>/<pod>" or "-n <ns> <pod>"
convention as ` + "`cluster pod get`" + `. The output table fuses
spec.containers[*] (the desired set) with status.containerStatuses[*]
(the runtime overlay):

  CONTAINER  IMAGE  READY  RESTARTS  STATE  PORTS

In ` + "`-o json`" + ` mode the per-container view is emitted verbatim as
{spec, status} so scripts can keep both.
`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			ns, name, err := clusteropts.SplitNsName(namespace, args[0])
			if err != nil {
				return err
			}
			return runList(c.Context(), o, ns, name)
		},
	}
	cmd.Flags().StringVarP(&namespace, "namespace", "n", "", "namespace (required when the positional argument is a bare pod name)")
	o.AddOutputFlags(cmd)
	return cmd
}

// containerView is the per-container projection emitted in `-o json`.
// We expose both the spec and status shapes so callers can decode
// either side without re-hitting the API.
type containerView struct {
	Name   string                `json:"name"`
	Spec   pod.PodContainer      `json:"spec"`
	Status *pod.PodContainerStatus `json:"status,omitempty"`
}

func runList(ctx context.Context, o *clusteropts.ClusterOptions, namespace, name string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	p, err := pod.Get(ctx, o, namespace, name)
	if err != nil {
		return err
	}

	statusByName := map[string]pod.PodContainerStatus{}
	for _, cs := range p.Status.ContainerStatuses {
		statusByName[cs.Name] = cs
	}
	views := make([]containerView, 0, len(p.Spec.Containers))
	for _, c := range p.Spec.Containers {
		v := containerView{Name: c.Name, Spec: c}
		if cs, ok := statusByName[c.Name]; ok {
			cs := cs
			v.Status = &cs
		}
		views = append(views, v)
	}

	if o.IsJSON() {
		return o.PrintJSON(struct {
			Pod        string          `json:"pod"`
			Namespace  string          `json:"namespace"`
			Containers []containerView `json:"containers"`
		}{Pod: name, Namespace: namespace, Containers: views})
	}
	return renderTable(views, o.NoHeaders)
}

func renderTable(views []containerView, noHeaders bool) error {
	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	defer tw.Flush()
	if !noHeaders {
		fmt.Fprintln(tw, "CONTAINER\tIMAGE\tREADY\tRESTARTS\tSTATE\tPORTS")
	}
	for _, v := range views {
		ready := "-"
		restarts := "-"
		state := "-"
		if v.Status != nil {
			if v.Status.Ready {
				ready = "true"
			} else {
				ready = "false"
			}
			restarts = fmt.Sprintf("%d", v.Status.RestartCount)
			state = describeContainerState(v.Status.State)
		}
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%s\n",
			v.Name, clusteropts.DashIfEmpty(v.Spec.Image), ready, restarts, state, formatPorts(v.Spec.Ports))
	}
	return nil
}

// formatPorts reduces []PodContainerPort to a comma-joined "name:
// port/proto" string suitable for a single column. Empty list ->
// "-".
func formatPorts(ports []pod.PodContainerPort) string {
	if len(ports) == 0 {
		return "-"
	}
	parts := make([]string, 0, len(ports))
	for _, p := range ports {
		proto := p.Protocol
		if proto == "" {
			proto = "TCP"
		}
		s := fmt.Sprintf("%d/%s", p.ContainerPort, proto)
		if p.Name != "" {
			s = p.Name + ":" + s
		}
		parts = append(parts, s)
	}
	return strings.Join(parts, ",")
}

// describeContainerState mirrors cluster/pod/get.go::describeContainerState
// — kept private per package so the leaf packages stay independent
// of each other.
func describeContainerState(state map[string]interface{}) string {
	if state == nil {
		return "-"
	}
	if _, ok := state["running"]; ok {
		return "Running"
	}
	if w, ok := state["waiting"].(map[string]interface{}); ok {
		if reason, ok := w["reason"].(string); ok && reason != "" {
			return "Waiting (" + reason + ")"
		}
		return "Waiting"
	}
	if t, ok := state["terminated"].(map[string]interface{}); ok {
		reason, _ := t["reason"].(string)
		ec, _ := t["exitCode"].(float64)
		if reason != "" {
			return fmt.Sprintf("Terminated (%s, exit %d)", reason, int(ec))
		}
		return fmt.Sprintf("Terminated (exit %d)", int(ec))
	}
	return "-"
}

