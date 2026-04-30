package workload

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/cmd/ctl/cluster/internal/clusteropts"
	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// NewDeleteCommand: `olares-cli cluster workload delete
// <ns/name | name> --kind X [-n NS] [--yes] [--propagation P]`.
//
// CLI-original verb (the SPA has no equivalent helper — workload
// deletion goes through the SPA's app-uninstall path which is a
// different abstraction). Here we issue the canonical
// `DELETE /apis/apps/v1/namespaces/<ns>/<kind>/<name>` directly so
// operators have a sharp tool when they really need to remove a
// workload object.
//
// --propagation maps to the apiserver's propagationPolicy query
// parameter:
//
//   - foreground (default): blocks until the cascade completes —
//     pods, replicaSets, etc. all gone before the call returns.
//   - background:           returns immediately; cascade happens out
//     of band.
//   - orphan:               leaves the dependents alone (rarely what
//     you want; provided for parity).
//
// Wrapped in ConfirmDestructive — workload deletion is permanent.
func NewDeleteCommand(f *cmdutil.Factory) *cobra.Command {
	o := clusteropts.NewClusterOptions(f)
	var (
		namespace      string
		kindRaw        string
		assumeYes      bool
		propagationRaw string
	)
	cmd := &cobra.Command{
		Use:   "delete <ns/name | name>",
		Short: "delete one Deployment / StatefulSet / DaemonSet (cascade by default)",
		Long: `Delete one workload object.

Issues ` + "`DELETE /apis/apps/v1/namespaces/<ns>/<kind>/<name>?propagationPolicy=<P>`" + `.

--propagation values (case-insensitive):
  foreground  (default) wait for the cascade to finish before returning;
              dependent pods / replicaSets are gone when the call returns.
  background  return immediately; apiserver runs the cascade out of band.
  orphan      delete the workload but leave the dependents behind
              (rarely useful; provided for parity with kubectl).

This is a CLI-original verb — the SPA has no direct workload-delete
button (it goes through app uninstall). Use this when you really
want to remove a single K8s object.

Pass --yes to skip the confirmation prompt for scripted use.
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
			policy, err := normalizePropagation(propagationRaw)
			if err != nil {
				return err
			}
			ns, name, err := clusteropts.SplitNsName(namespace, args[0])
			if err != nil {
				return err
			}
			return runDelete(c.Context(), o, ns, name, plural, policy, assumeYes)
		},
	}
	cmd.Flags().StringVarP(&namespace, "namespace", "n", "", "namespace (required when the positional argument is a bare name)")
	cmd.Flags().StringVar(&kindRaw, "kind", "", "workload kind: deployment | statefulset | daemonset (REQUIRED)")
	cmd.Flags().BoolVarP(&assumeYes, "yes", "y", false, "skip the confirmation prompt")
	cmd.Flags().StringVar(&propagationRaw, "propagation", "foreground", "deletion propagation policy: foreground | background | orphan")
	o.AddOutputFlags(cmd)
	return cmd
}

// normalizePropagation maps user-facing --propagation values to the
// apiserver's enum. We accept lowercase / mixed-case for ergonomics
// but always send the canonical apiserver-recognized form.
func normalizePropagation(s string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "foreground":
		return "Foreground", nil
	case "background":
		return "Background", nil
	case "orphan":
		return "Orphan", nil
	case "":
		return "Foreground", nil
	default:
		return "", fmt.Errorf("unsupported --propagation value %q (want one of: foreground, background, orphan)", s)
	}
}

// deleteResult is the JSON-mode shape emitted on success. Synthesized
// rather than forwarding the apiserver's metav1.Status response.
type deleteResult struct {
	Operation         string `json:"operation"`
	Kind              string `json:"kind"`
	Namespace         string `json:"namespace"`
	Name              string `json:"name"`
	PropagationPolicy string `json:"propagationPolicy"`
}

func runDelete(ctx context.Context, o *clusteropts.ClusterOptions, namespace, name, kindPlural, policy string, assumeYes bool) error {
	if ctx == nil {
		ctx = context.Background()
	}

	if err := clusteropts.ConfirmDestructive(os.Stderr, os.Stdin,
		fmt.Sprintf("Delete %s %s/%s (propagationPolicy=%s)? This removes the workload and (with foreground/background) its pods",
			SingularKind(kindPlural), namespace, name, policy),
		assumeYes); err != nil {
		return err
	}

	client, err := o.Prepare()
	if err != nil {
		return err
	}
	q := url.Values{}
	q.Set("propagationPolicy", policy)
	path := buildGetPath(namespace, kindPlural, name) + "?" + q.Encode()
	if err := client.DoJSON(ctx, "DELETE", path, nil, nil); err != nil {
		return fmt.Errorf("delete %s %s/%s: %w", SingularKind(kindPlural), namespace, name, err)
	}

	result := deleteResult{
		Operation:         "delete",
		Kind:              SingularKind(kindPlural),
		Namespace:         namespace,
		Name:              name,
		PropagationPolicy: policy,
	}
	if o.IsJSON() {
		return o.PrintJSON(result)
	}
	if !o.Quiet {
		fmt.Fprintf(os.Stdout, "%s %s/%s deleted (propagationPolicy=%s)\n",
			SingularKind(kindPlural), namespace, name, policy)
	}
	return nil
}
