package workload

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/cmd/ctl/cluster/internal/clusteropts"
	"github.com/beclab/Olares/cli/pkg/clusterclient"
	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// NewYAMLCommand: `olares-cli cluster workload yaml <ns/name | name>
// [-n NS] --kind <kind>`. Same K8s native endpoint as `get`, but
// bytes are forwarded through sigs.k8s.io/yaml so unknown fields
// stay in the output (the typed Workload struct only models the
// fields verbs render).
func NewYAMLCommand(f *cmdutil.Factory) *cobra.Command {
	o := clusteropts.NewClusterOptions(f)
	var (
		namespace string
		kindRaw   string
	)
	cmd := &cobra.Command{
		Use:   "yaml <ns/name | name>",
		Short: "print one workload's full K8s-native YAML",
		Long: `Print one workload's full K8s-native YAML.

--kind is required (one of: deployment, statefulset, daemonset).
Identity follows the same "<namespace>/<name>" or "-n <ns> <name>"
convention as ` + "`cluster workload get`" + `. The output is the JSON response
from /apis/apps/v1/namespaces/<ns>/<kind>/<name> converted to YAML
— every field the server returned (including ones the CLI's typed
struct doesn't know about) is preserved.

This verb deliberately does NOT honor --output (yaml is the whole
point of the verb). For JSON, use ` + "`cluster workload get -o json`" + `.
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
			ns, name, err := clusteropts.SplitNsName(namespace, args[0])
			if err != nil {
				return err
			}
			return runYAML(c.Context(), o, ns, name, plural)
		},
	}
	cmd.Flags().StringVarP(&namespace, "namespace", "n", "", "namespace (required when the positional argument is a bare name)")
	cmd.Flags().StringVar(&kindRaw, "kind", "", "workload kind: deployment | statefulset | daemonset (REQUIRED)")
	return cmd
}

func runYAML(ctx context.Context, o *clusteropts.ClusterOptions, namespace, name, kindPlural string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	client, err := o.Prepare()
	if err != nil {
		return err
	}
	body, err := clusterclient.GetRaw(ctx, client, buildGetPath(namespace, kindPlural, name))
	if err != nil {
		return fmt.Errorf("get %s %s/%s: %w", SingularKind(kindPlural), namespace, name, err)
	}
	out, err := clusteropts.JSONToYAML(body)
	if err != nil {
		return fmt.Errorf("convert %s %s/%s response to YAML: %w", SingularKind(kindPlural), namespace, name, err)
	}
	if err := o.WriteStdout(out); err != nil {
		return err
	}
	if !o.Quiet && !strings.HasSuffix(string(out), "\n") {
		fmt.Fprintln(os.Stdout)
	}
	return nil
}
