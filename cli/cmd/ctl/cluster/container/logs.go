package container

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/cmd/ctl/cluster/internal/clusteropts"
	"github.com/beclab/Olares/cli/cmd/ctl/cluster/pod"
	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// NewLogsCommand: `olares-cli cluster container logs
// <ns/pod/container | ns/pod | pod> [-n NS] [-c NAME] ...`.
//
// Thin alias over `cluster pod logs` — the only practical difference
// is the positional grammar. Container is mandatory here (it's the
// noun the verb is named after), so we accept a 3-segment positional
// `<ns>/<pod>/<container>` in addition to the 2-segment "<ns/pod>"
// + --container variant `cluster pod logs` already accepts.
//
// Same wire endpoint, same polling --follow, same option semantics —
// every flag delegates straight through to pod.RunLogs so behavior
// stays bit-exact between the two entry points.
func NewLogsCommand(f *cmdutil.Factory) *cobra.Command {
	o := clusteropts.NewClusterOptions(f)
	var (
		namespace string
		container string
		tail      int
		since     time.Duration
		limitB    int
		ts        bool
		follow    bool
		interval  time.Duration
		previous  bool
	)
	cmd := &cobra.Command{
		Use:   "logs <ns/pod/container | ns/pod | pod>",
		Short: "stream a container's log buffer (--follow polls, doesn't stream)",
		Long: `Print the log buffer of one container inside a pod.

Identity grammar:
  <ns>/<pod>/<container>     three-segment positional, no other flags
                             needed
  <ns>/<pod> --container N   two-segment positional (mirrors
                             ` + "`cluster container env`" + ` and
                             ` + "`cluster pod logs`" + `)
  <pod>      -n NS -c N      bare pod name with -n / --container

The container name is mandatory for this verb. (` + "`cluster pod logs`" + `
allows it to be omitted when the target pod is single-container; the
container alias asks you to be explicit either way.)

Everything else — --follow polling, --tail / --since on the initial
fetch, --previous, --timestamps — is identical to ` + "`cluster pod logs`" + `;
this verb just delegates.
`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			ns, podName, ctr, err := splitNsPodContainer(namespace, container, args[0])
			if err != nil {
				return err
			}
			if previous && follow {
				return fmt.Errorf("--follow and --previous are mutually exclusive")
			}
			return pod.RunLogs(c.Context(), o, ns, podName, pod.LogsOptions{
				Container:    ctr,
				TailLines:    tail,
				SinceSeconds: int(since / time.Second),
				LimitBytes:   limitB,
				Timestamps:   ts,
				Follow:       follow,
				Interval:     interval,
				Previous:     previous,
			})
		},
	}
	cmd.Flags().StringVarP(&namespace, "namespace", "n", "", "namespace (required when the positional argument doesn't include one)")
	cmd.Flags().StringVarP(&container, "container", "c", "", "container name (required when the positional argument doesn't include one)")
	cmd.Flags().IntVar(&tail, "tail", 200, "show the last N lines on the initial fetch (0 = unlimited; --follow always advances by sinceTime after the first fetch)")
	cmd.Flags().DurationVar(&since, "since", 0, "show logs newer than this duration ago on the initial fetch (e.g. 5m, 1h); 0 = unlimited")
	cmd.Flags().IntVar(&limitB, "limit-bytes", 0, "cap the response body size in bytes (0 = unlimited)")
	cmd.Flags().BoolVar(&ts, "timestamps", true, "ask the server to prefix every line with an RFC3339 timestamp")
	cmd.Flags().BoolVarP(&follow, "follow", "f", false, "keep polling for new lines until interrupted (Ctrl-C to stop)")
	cmd.Flags().DurationVar(&interval, "interval", 2*time.Second, "polling interval when --follow is set")
	cmd.Flags().BoolVar(&previous, "previous", false, "fetch the previous container instance's logs (after a crash); incompatible with --follow")
	return cmd
}

// splitNsPodContainer accepts the alias-specific 3-segment positional
// `<ns>/<pod>/<container>` in addition to the 2-segment `<ns>/<pod>`
// or bare `<pod>` grammars `cluster pod logs` accepts. When the
// container slot isn't filled by the positional, --container must
// supply it (this verb is named after the container, after all).
//
// nsFlag / containerFlag conflict checks mirror clusteropts.SplitNsName
// so users get a single, predictable error grammar across both entry
// points.
func splitNsPodContainer(nsFlag, containerFlag, arg string) (string, string, string, error) {
	parts := strings.Split(arg, "/")
	switch len(parts) {
	case 1:
		// Bare pod name; ns and container both come from flags.
		if nsFlag == "" {
			return "", "", "", fmt.Errorf("namespace required: pass --namespace or use <namespace>/<pod>[/<container>]")
		}
		if containerFlag == "" {
			return "", "", "", fmt.Errorf("container required: pass --container or use <namespace>/<pod>/<container>")
		}
		return nsFlag, arg, containerFlag, nil
	case 2:
		// "<ns>/<pod>"; container from --container.
		if parts[0] == "" || parts[1] == "" {
			return "", "", "", fmt.Errorf("invalid <namespace>/<pod>: %q", arg)
		}
		if nsFlag != "" && nsFlag != parts[0] {
			return "", "", "", fmt.Errorf("argument namespace %q conflicts with --namespace %q", parts[0], nsFlag)
		}
		if containerFlag == "" {
			return "", "", "", fmt.Errorf("container required: pass --container or use <namespace>/<pod>/<container>")
		}
		return parts[0], parts[1], containerFlag, nil
	case 3:
		// "<ns>/<pod>/<container>"; nothing else needed.
		if parts[0] == "" || parts[1] == "" || parts[2] == "" {
			return "", "", "", fmt.Errorf("invalid <namespace>/<pod>/<container>: %q", arg)
		}
		if nsFlag != "" && nsFlag != parts[0] {
			return "", "", "", fmt.Errorf("argument namespace %q conflicts with --namespace %q", parts[0], nsFlag)
		}
		if containerFlag != "" && containerFlag != parts[2] {
			return "", "", "", fmt.Errorf("argument container %q conflicts with --container %q", parts[2], containerFlag)
		}
		return parts[0], parts[1], parts[2], nil
	default:
		return "", "", "", fmt.Errorf("invalid identity %q: expected one of <pod>, <ns>/<pod>, <ns>/<pod>/<container>", arg)
	}
}
