package job

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

// NewYAMLCommand: `olares-cli cluster job yaml <ns/name | name>
// [-n ns]`. Same endpoint as `job get`; difference is we forward the
// raw response bytes through sigs.k8s.io/yaml so the K8s native
// JSON-shaped object converts to YAML preserving every field
// (including ones not modeled in this package's typed Job).
func NewYAMLCommand(f *cmdutil.Factory) *cobra.Command {
	o := clusteropts.NewClusterOptions(f)
	var namespace string
	cmd := &cobra.Command{
		Use:   "yaml <ns/name | name>",
		Short: "print one Job's full K8s-native YAML",
		Long: `Print one Job's full K8s-native YAML.

Identity follows the same "<namespace>/<name>" or "-n <ns> <name>"
convention as ` + "`cluster job get`" + `. Output is the JSON response from
` + "`/apis/batch/v1/namespaces/<ns>/jobs/<name>`" + ` converted to YAML —
every field the server returned (including ones the CLI's typed
struct doesn't know about) is preserved.

This verb deliberately does NOT honor --output (yaml is the whole
point). For JSON, use ` + "`cluster job get -o json`" + `.
`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			ns, name, err := clusteropts.SplitNsName(namespace, args[0])
			if err != nil {
				return err
			}
			return runYAML(c.Context(), o, ns, name)
		},
	}
	cmd.Flags().StringVarP(&namespace, "namespace", "n", "", "namespace (required when the positional argument is a bare name)")
	o.AddQuietFlag(cmd)
	return cmd
}

func runYAML(ctx context.Context, o *clusteropts.ClusterOptions, namespace, name string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	client, err := o.Prepare()
	if err != nil {
		return err
	}
	body, err := clusterclient.GetRaw(ctx, client, buildGetPath(namespace, name))
	if err != nil {
		return fmt.Errorf("get job %s/%s: %w", namespace, name, err)
	}
	out, err := clusteropts.JSONToYAML(body)
	if err != nil {
		return fmt.Errorf("convert job %s/%s response to YAML: %w", namespace, name, err)
	}
	if err := o.WriteStdout(out); err != nil {
		return err
	}
	if !o.Quiet && !strings.HasSuffix(string(out), "\n") {
		fmt.Fprintln(os.Stdout)
	}
	return nil
}
