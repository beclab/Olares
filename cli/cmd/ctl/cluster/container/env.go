package container

import (
	"context"
	"fmt"
	"os"
	"sort"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/cmd/ctl/cluster/internal/clusteropts"
	"github.com/beclab/Olares/cli/cmd/ctl/cluster/pod"
	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// NewEnvCommand: `olares-cli cluster container env <ns/pod | pod>
// [-n NS] [--container NAME] [-o table|json]`.
//
// Renders explicit env vars from spec.containers[*].env. Source-of-
// truth is the same pod object `cluster pod get` returns; we just
// project containers[*].env into a flat table (or per-container
// sections when --container is omitted).
//
// IMPORTANT: this verb does NOT resolve valueFrom references. A var
// declared as `valueFrom: { secretKeyRef: { name: 'foo', key: 'k' } }`
// shows up with VALUE blank and FROM = "secretKey foo/k", so users
// can see WHERE the value would come from at pod-startup time
// without leaking the secret material itself. Resolving the ref
// requires extra GETs against ConfigMap / Secret and a per-call
// scope decision; that lives behind a future `--resolve` flag.
//
// Also intentionally NOT covered: envFrom (the implicit
// configMapRef / secretRef sets that import every key from a CM /
// Secret). Add when a verb actually needs the implicit set.
func NewEnvCommand(f *cmdutil.Factory) *cobra.Command {
	o := clusteropts.NewClusterOptions(f)
	var (
		namespace string
		container string
	)
	cmd := &cobra.Command{
		Use:   "env <ns/pod | pod>",
		Short: "list explicit env vars on one (or every) container in a pod",
		Long: `List explicit env vars declared on the target pod's containers.

Identity follows the same "<namespace>/<pod>" or "-n <ns> <pod>"
convention as ` + "`cluster pod get`" + `. Pass --container <name> to scope to
a single container; without it, each container is rendered as a
separate section in table mode.

valueFrom references (configMapKeyRef / secretKeyRef / fieldRef /
resourceFieldRef) are surfaced via a FROM column rather than
resolved — the value is left blank and the column shows where the
value would come from at pod startup. Resolving refs requires extra
API calls against ConfigMaps / Secrets; that's a future --resolve
flag.

envFrom (implicit configMapRef / secretRef sets) is intentionally
NOT enumerated.
`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			ns, name, err := clusteropts.SplitNsName(namespace, args[0])
			if err != nil {
				return err
			}
			return runEnv(c.Context(), o, ns, name, container)
		},
	}
	cmd.Flags().StringVarP(&namespace, "namespace", "n", "", "namespace (required when the positional argument is a bare pod name)")
	cmd.Flags().StringVar(&container, "container", "", "scope to a single container (default: every container in the pod)")
	o.AddOutputFlags(cmd)
	return cmd
}

// containerEnv is the per-container env projection emitted in
// `-o json` and consumed by renderEnvSections.
type containerEnv struct {
	Container string          `json:"container"`
	Env       []pod.PodEnvVar `json:"env"`
}

func runEnv(ctx context.Context, o *clusteropts.ClusterOptions, namespace, name, containerFilter string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	p, err := pod.Get(ctx, o, namespace, name)
	if err != nil {
		return err
	}

	var sections []containerEnv
	for _, c := range p.Spec.Containers {
		if containerFilter != "" && c.Name != containerFilter {
			continue
		}
		sections = append(sections, containerEnv{Container: c.Name, Env: c.Env})
	}
	if len(sections) == 0 {
		if containerFilter != "" {
			return fmt.Errorf("container %q not found in pod %s/%s", containerFilter, namespace, name)
		}
		// Pod has no containers (extremely unusual but technically
		// possible for a malformed object); print and return.
		fmt.Fprintf(os.Stderr, "no containers in pod %s/%s\n", namespace, name)
		return nil
	}

	if o.IsJSON() {
		return o.PrintJSON(struct {
			Pod        string         `json:"pod"`
			Namespace  string         `json:"namespace"`
			Containers []containerEnv `json:"containers"`
		}{Pod: name, Namespace: namespace, Containers: sections})
	}
	return renderEnvSections(sections, o.NoHeaders)
}

func renderEnvSections(sections []containerEnv, noHeaders bool) error {
	for i, s := range sections {
		// Section header — keep the format minimal so grep / awk over
		// the output stays cheap. Skip the header when there's only
		// one section (a single --container request).
		if len(sections) > 1 {
			if i > 0 {
				fmt.Fprintln(os.Stdout)
			}
			fmt.Fprintf(os.Stdout, "Container: %s\n", s.Container)
		}
		if len(s.Env) == 0 {
			fmt.Fprintln(os.Stdout, "(no explicit env vars)")
			continue
		}
		tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		if !noHeaders {
			fmt.Fprintln(tw, "NAME\tVALUE\tFROM")
		}
		// Stable order so repeated runs diff cleanly. Pods don't
		// guarantee env declaration order in K8s API responses.
		envs := append([]pod.PodEnvVar(nil), s.Env...)
		sort.SliceStable(envs, func(i, j int) bool { return envs[i].Name < envs[j].Name })
		for _, e := range envs {
			value := e.Value
			from := "-"
			if e.ValueFrom != nil {
				from = describeEnvFrom(e.ValueFrom)
			}
			if value == "" {
				value = "-"
			}
			fmt.Fprintf(tw, "%s\t%s\t%s\n", e.Name, value, from)
		}
		tw.Flush()
	}
	return nil
}

// describeEnvFrom turns a corev1.EnvVarSource into a short FROM
// label suitable for a single column. Mirrors how `kubectl describe
// pod` prints env sources but compressed.
func describeEnvFrom(src *pod.PodEnvVarFrom) string {
	switch {
	case src == nil:
		return "-"
	case src.ConfigMapKeyRef != nil:
		return fmt.Sprintf("configMapKey %s/%s", src.ConfigMapKeyRef.Name, src.ConfigMapKeyRef.Key)
	case src.SecretKeyRef != nil:
		return fmt.Sprintf("secretKey %s/%s", src.SecretKeyRef.Name, src.SecretKeyRef.Key)
	case src.FieldRef != nil:
		return fmt.Sprintf("fieldRef %s", src.FieldRef.FieldPath)
	case src.ResourceFieldRef != nil:
		c := src.ResourceFieldRef.ContainerName
		if c == "" {
			c = "(self)"
		}
		return fmt.Sprintf("resourceFieldRef %s/%s", c, src.ResourceFieldRef.Resource)
	}
	return "-"
}
