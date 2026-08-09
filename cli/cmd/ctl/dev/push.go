package dev

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/cmd/ctl/cluster/node"
	"github.com/beclab/Olares/cli/pkg/cmdutil"
	"github.com/beclab/Olares/cli/pkg/devdeploy"
)

// NewPushCommand: `olares-cli dev push <image> [--transport T]
// [--image-tool docker|podman] [--skip-arch-check]`.
func NewPushCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		transport     string
		imageTool     string
		skipArchCheck bool
		quiet         bool
	)
	cmd := &cobra.Command{
		Use:   "push <image>",
		Short: "side-load a locally built image into the Olares node's containerd",
		Long: `Move a locally built image into the image store the kubelet reads.

The image is exported with ` + "`docker save`" + ` (or ` + "`podman save`" + `) and
streamed into ` + "`ctr -n " + devdeploy.ContainerdNamespace + " images import -`" + ` on
the node. Nothing is written to disk on either side, and no registry is
involved.

The containerd namespace is not configurable: an image imported
anywhere other than ` + devdeploy.ContainerdNamespace + ` is invisible to
the kubelet, which is the most common way a hand-rolled ctr import
appears to succeed and then ImagePullBackOffs.

Before transferring, the image's architecture is compared against the
cluster's nodes and a mismatch is refused — a wrong-arch image imports
perfectly and then fails at runtime with an "exec format error" buried
in a container log. Pass --skip-arch-check to override (or when the
check cannot reach the cluster).

This only moves bytes. Point a workload at the image with
` + "`olares-cli dev deploy`" + `.
`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			image := strings.TrimSpace(args[0])
			if image == "" {
				return fmt.Errorf("image is required")
			}
			return runPush(c.Context(), f, pushParams{
				Image:         image,
				Transport:     transport,
				ImageTool:     imageTool,
				SkipArchCheck: skipArchCheck,
				Quiet:         quiet,
			})
		},
	}
	cmd.Flags().StringVar(&transport, "transport", devdeploy.TransportAuto,
		"how to move the image: auto | local | ssh | registry | api")
	cmd.Flags().StringVar(&imageTool, "image-tool", "",
		"local tool used to export the image: docker | podman (default: whichever is on PATH)")
	cmd.Flags().BoolVar(&skipArchCheck, "skip-arch-check", false,
		"do not compare the image's architecture against the cluster's nodes")
	cmd.Flags().BoolVarP(&quiet, "quiet", "q", false, "suppress progress output; exit code indicates success/failure")
	return cmd
}

type pushParams struct {
	Image         string
	Transport     string
	ImageTool     string
	SkipArchCheck bool
	Quiet         bool
}

func runPush(ctx context.Context, f *cmdutil.Factory, p pushParams) error {
	if ctx == nil {
		ctx = context.Background()
	}
	t, err := devdeploy.Select(p.Transport, devdeploy.Options{
		Node:      currentDevNode(),
		ImageTool: p.ImageTool,
	})
	if err != nil {
		return err
	}

	if !p.SkipArchCheck {
		if err := checkArch(ctx, f, p.Image, p.ImageTool, p.Quiet); err != nil {
			return err
		}
	}

	// io.Writer, not *os.File: a typed nil pointer in an interface is
	// not a nil interface, so keeping the static type as the interface
	// is what makes the `progress == nil` checks inside the transports
	// actually fire.
	var progress io.Writer
	if !p.Quiet {
		progress = os.Stderr
		fmt.Fprintf(progress, "transport: %s — %s\n", t.Name(), t.Describe())
	}
	if err := t.Import(ctx, p.Image, progress); err != nil {
		return err
	}

	if !p.Quiet {
		fmt.Fprintf(os.Stdout, "imported %s\n", p.Image)
	}
	return nil
}

// checkArch compares the local image's architecture with the cluster's.
//
// Failure to reach the cluster is a warning, not an error: the push
// itself may well be the thing that fixes a broken cluster, and
// refusing to move bytes because a read-only preflight could not run
// would be the wrong trade.
func checkArch(ctx context.Context, f *cmdutil.Factory, image, imageTool string, quiet bool) error {
	imageArch, err := localImageArch(ctx, imageTool, image)
	if err != nil {
		if !quiet {
			fmt.Fprintf(os.Stderr, "warning: could not determine the image's architecture (%v); skipping the check\n", err)
		}
		return nil
	}
	nodeArches, err := node.Architectures(ctx, f)
	if err != nil {
		if !quiet {
			fmt.Fprintf(os.Stderr, "warning: could not read node architectures (%v); skipping the check\n", err)
		}
		return nil
	}
	if len(nodeArches) == 0 {
		return nil
	}
	for _, arch := range nodeArches {
		if arch == imageArch {
			return nil
		}
	}
	return fmt.Errorf(
		"%s is built for %s but the cluster's nodes are %s\n"+
			"Rebuild for the node's architecture (docker build --platform linux/%s ...) "+
			"or pass --skip-arch-check if you know better",
		image, imageArch, strings.Join(nodeArches, ", "), nodeArches[0])
}

// localImageArch reads the image's architecture from the local image
// store. Both docker and podman accept this inspect format.
func localImageArch(ctx context.Context, imageTool, image string) (string, error) {
	tool := imageTool
	if tool == "" {
		for _, candidate := range []string{"docker", "podman"} {
			if _, err := exec.LookPath(candidate); err == nil {
				tool = candidate
				break
			}
		}
	}
	if tool == "" {
		return "", fmt.Errorf("neither docker nor podman on PATH")
	}
	out, err := exec.CommandContext(ctx, tool, "image", "inspect", "--format", "{{.Architecture}}", image).Output()
	if err != nil {
		return "", fmt.Errorf("%s image inspect %s: %w", tool, image, err)
	}
	arch := strings.TrimSpace(string(out))
	if arch == "" {
		return "", fmt.Errorf("%s reported an empty architecture for %s", tool, image)
	}
	return arch, nil
}
