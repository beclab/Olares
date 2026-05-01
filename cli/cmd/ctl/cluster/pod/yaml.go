package pod

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/cmd/ctl/cluster/internal/clusteropts"
	"github.com/beclab/Olares/cli/pkg/clusterclient"
	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// NewYAMLCommand: `olares-cli cluster pod yaml <ns/name | name> [-n ns]`.
//
// Fetches the same K8s-native pod object as `pod get`
// (`/api/v1/namespaces/<ns>/pods/<name>`) and converts the JSON
// response body to YAML for the user. Mirrors `kubectl get pod -o yaml`
// in feel: faithful round-trip of every field the server exposed
// (we don't decode through our minimal Pod struct so unknown fields
// stay in the output).
//
// We lean on sigs.k8s.io/yaml (already a transitive dep via
// k8s.io/client-go in the rest of the CLI) so the K8s-native field
// ordering convention (kind, apiVersion, metadata, ...) is preserved.
func NewYAMLCommand(f *cmdutil.Factory) *cobra.Command {
	o := clusteropts.NewClusterOptions(f)
	var namespace string

	cmd := &cobra.Command{
		Use:   "yaml <ns/name | name>",
		Short: "print one pod's full K8s-native YAML",
		Long: `Print one pod's full K8s-native YAML.

Identity follows the same "<namespace>/<name>" or "-n <ns> <name>"
convention as ` + "`cluster pod get`" + `. Output is the JSON response from
` + "`/api/v1/namespaces/<ns>/pods/<name>`" + ` converted to YAML — every
field the server returned (including ones the CLI's typed struct
doesn't know about) is preserved.

This verb deliberately does NOT honor --output (yaml is the whole
point of the verb). For JSON, use ` + "`cluster pod get -o json`" + `.
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

	path := fmt.Sprintf("/api/v1/namespaces/%s/pods/%s",
		url.PathEscape(namespace), url.PathEscape(name))
	body, err := clusterclient.GetRaw(ctx, client, path)
	if err != nil {
		return fmt.Errorf("get pod %s/%s: %w", namespace, name, err)
	}

	out, err := clusteropts.JSONToYAML(body)
	if err != nil {
		return fmt.Errorf("convert pod %s/%s response to YAML: %w", namespace, name, err)
	}
	if err := o.WriteStdout(out); err != nil {
		return err
	}
	if !o.Quiet && !strings.HasSuffix(string(out), "\n") {
		fmt.Fprintln(os.Stdout)
	}
	return nil
}
