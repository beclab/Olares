package pod

import (
	"context"
	"fmt"
	"net/url"
	"os"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/cmd/ctl/cluster/internal/clusteropts"
	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// NewDeleteCommand: `olares-cli cluster pod delete <ns/name | name>
// [-n NS] [--yes] [--grace-period N]`.
//
// Calls SPA's deletePod (apps/.../controlPanelCommon/network/index.ts):
// `DELETE /api/v1/namespaces/<ns>/pods/<name>` (optional
// `?gracePeriodSeconds=N`).
//
// Server-decides authority: a 403 means the operator's token can't
// delete in that namespace; we surface the error verbatim. CLI does
// NOT perform any local authorization check before issuing the
// request.
//
// Wrapped in ConfirmDestructive — pod deletion is mutating (the
// controller will recreate it for managed pods, but standalone /
// stuck pods don't come back). --yes opts out for scripted use.
func NewDeleteCommand(f *cmdutil.Factory) *cobra.Command {
	o := clusteropts.NewClusterOptions(f)
	var (
		namespace   string
		assumeYes   bool
		gracePeriod int
	)
	cmd := &cobra.Command{
		Use:   "delete <ns/name | name>",
		Short: "delete one Pod",
		Long: `Delete one Pod by name.

Issues ` + "`DELETE /api/v1/namespaces/<ns>/pods/<name>`" + `; the controller
that owns the pod (Deployment / StatefulSet / DaemonSet / Job /
ReplicaSet) decides whether to recreate it. For standalone pods,
this is final.

--grace-period passes ?gracePeriodSeconds=<N> verbatim. Default -1
means "let the apiserver use the pod's own terminationGracePeriodSeconds";
0 forces immediate kill (matches kubectl --grace-period=0).

Pass --yes to skip the confirmation prompt for scripted use.
`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			ns, name, err := clusteropts.SplitNsName(namespace, args[0])
			if err != nil {
				return err
			}
			return runDelete(c.Context(), o, ns, name, assumeYes, gracePeriod)
		},
	}
	cmd.Flags().StringVarP(&namespace, "namespace", "n", "", "namespace (required when the positional argument is a bare name)")
	cmd.Flags().BoolVarP(&assumeYes, "yes", "y", false, "skip the confirmation prompt")
	cmd.Flags().IntVar(&gracePeriod, "grace-period", -1, "graceful termination period in seconds (-1 = pod default; 0 = immediate)")
	o.AddOutputFlags(cmd)
	return cmd
}

// deleteResult is the JSON-mode shape emitted on success. Synthesized
// rather than forwarding the apiserver's metav1.Status response (the
// SPA discards it too — success is signaled by 2xx).
type deleteResult struct {
	Operation   string `json:"operation"`
	Namespace   string `json:"namespace"`
	Pod         string `json:"pod"`
	GracePeriod int    `json:"gracePeriodSeconds,omitempty"`
}

// RunDelete is the exported entry point used by `pod restart` (which
// is the same DELETE under a friendlier verb). Keeps the wire-call
// in one place.
func RunDelete(ctx context.Context, o *clusteropts.ClusterOptions, namespace, name string, assumeYes bool, gracePeriod int, opName string) error {
	return runDeleteOp(ctx, o, namespace, name, assumeYes, gracePeriod, opName)
}

func runDelete(ctx context.Context, o *clusteropts.ClusterOptions, namespace, name string, assumeYes bool, gracePeriod int) error {
	return runDeleteOp(ctx, o, namespace, name, assumeYes, gracePeriod, "delete")
}

func runDeleteOp(ctx context.Context, o *clusteropts.ClusterOptions, namespace, name string, assumeYes bool, gracePeriod int, opName string) error {
	if ctx == nil {
		ctx = context.Background()
	}

	if err := clusteropts.ConfirmDestructive(os.Stderr, os.Stdin,
		fmt.Sprintf("%s pod %s/%s?", capitalize(opName), namespace, name),
		assumeYes); err != nil {
		return err
	}

	client, err := o.Prepare()
	if err != nil {
		return err
	}
	path := fmt.Sprintf("/api/v1/namespaces/%s/pods/%s",
		url.PathEscape(namespace), url.PathEscape(name))
	if gracePeriod >= 0 {
		q := url.Values{}
		q.Set("gracePeriodSeconds", fmt.Sprintf("%d", gracePeriod))
		path += "?" + q.Encode()
	}
	if err := client.DoJSON(ctx, "DELETE", path, nil, nil); err != nil {
		return fmt.Errorf("%s pod %s/%s: %w", opName, namespace, name, err)
	}

	result := deleteResult{
		Operation:   opName,
		Namespace:   namespace,
		Pod:         name,
		GracePeriod: gracePeriod,
	}
	if o.IsJSON() {
		return o.PrintJSON(result)
	}
	if !o.Quiet {
		fmt.Fprintf(os.Stdout, "pod %s/%s %sed\n", namespace, name, opName)
	}
	return nil
}

// capitalize uppercases the first byte of an ASCII verb. We only
// pass "delete" / "restart" so the simple-byte version is fine; not
// safe for arbitrary unicode strings.
func capitalize(s string) string {
	if s == "" {
		return s
	}
	if s[0] >= 'a' && s[0] <= 'z' {
		return string(s[0]-32) + s[1:]
	}
	return s
}
