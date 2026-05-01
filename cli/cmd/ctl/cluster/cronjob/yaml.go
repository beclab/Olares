package cronjob

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

// NewYAMLCommand: `olares-cli cluster cronjob yaml <ns/name | name>
// [-n ns]`. Same endpoint as `cronjob get`; difference is we forward
// the raw response bytes through sigs.k8s.io/yaml so every field is
// preserved (including ones the typed CronJob doesn't model).
func NewYAMLCommand(f *cmdutil.Factory) *cobra.Command {
	o := clusteropts.NewClusterOptions(f)
	var namespace string
	cmd := &cobra.Command{
		Use:   "yaml <ns/name | name>",
		Short: "print one CronJob's full K8s-native YAML",
		Long: `Print one CronJob's full K8s-native YAML.

Identity follows the same "<namespace>/<name>" or "-n <ns> <name>"
convention as ` + "`cluster cronjob get`" + `. Output is the JSON response
from ` + "`/apis/batch/v1beta1/namespaces/<ns>/cronjobs/<name>`" + `
converted to YAML — every field the server returned is preserved.

This verb deliberately does NOT honor --output (yaml is the whole
point). For JSON, use ` + "`cluster cronjob get -o json`" + `.
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
		return fmt.Errorf("get cronjob %s/%s: %w", namespace, name, err)
	}
	out, err := clusteropts.JSONToYAML(body)
	if err != nil {
		return fmt.Errorf("convert cronjob %s/%s response to YAML: %w", namespace, name, err)
	}
	if err := o.WriteStdout(out); err != nil {
		return err
	}
	if !o.Quiet && !strings.HasSuffix(string(out), "\n") {
		fmt.Fprintln(os.Stdout)
	}
	return nil
}
