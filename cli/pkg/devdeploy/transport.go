// Package devdeploy implements the transports that move a locally built
// container image into a running Olares node's image store, plus the
// helpers `olares-cli dev` uses to pick between them.
//
// The package deliberately does NOT contain the workload-patching logic:
// that lives in cmd/ctl/cluster/workload (RunSetImage) and is shared
// verbatim, so `dev deploy` and `cluster workload set-image` cannot
// drift. Transport is the only thing that varies with where the CLI is
// running; the patch always goes through the same authenticated
// ControlHub path regardless.
package devdeploy

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/beclab/Olares/cli/pkg/cliconfig"
)

// Transport names accepted by `dev push --transport`.
const (
	TransportAuto     = "auto"
	TransportLocal    = "local"
	TransportSSH      = "ssh"
	TransportRegistry = "registry"
	TransportAPI      = "api"
)

// ContainerdNamespace is the containerd namespace kubelet pulls from.
// An image imported into any other namespace is invisible to the
// kubelet, which is the single most common way a hand-rolled `ctr
// images import` appears to succeed and then ImagePullBackOffs.
const ContainerdNamespace = "k8s.io"

// Transport moves an image into a node's image store.
type Transport interface {
	// Name is the --transport value that selects this transport.
	Name() string
	// Import streams the image into the node's containerd. progress is
	// where human-readable step output goes; nil discards it.
	Import(ctx context.Context, image string, progress io.Writer) error
	// Describe is a one-line summary of what this transport will do,
	// shown before a push so the user can see where the bytes go.
	Describe() string
}

// Options carries everything the transports need. Zero values are
// meaningful: a nil Node simply makes the ssh transport unavailable.
type Options struct {
	// Node is the SSH coordinate for TransportSSH.
	Node *cliconfig.DevNodeConfig
	// ImageTool overrides the local image tool ("docker" / "podman").
	// Empty means auto-detect.
	ImageTool string
}

// Select resolves a --transport value to a concrete Transport.
//
// For TransportAuto it walks the ladder local → ssh → registry, which
// orders by iteration speed: importing straight into the local
// containerd is the fastest, streaming over SSH costs a network hop,
// and a registry round-trip costs two plus a public push.
func Select(name string, opts Options) (Transport, error) {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "", TransportAuto:
		return selectAuto(opts)
	case TransportLocal:
		t, why := newLocalTransport(opts)
		if t == nil {
			return nil, fmt.Errorf("--transport local is not usable here: %s", why)
		}
		return t, nil
	case TransportSSH:
		t, why := newSSHTransport(opts)
		if t == nil {
			return nil, fmt.Errorf("--transport ssh is not usable here: %s", why)
		}
		return t, nil
	case TransportRegistry:
		return nil, fmt.Errorf("--transport registry is not implemented yet; " +
			"push the image to a registry the node can pull from and use " +
			"`olares-cli cluster workload set-image` directly")
	case TransportAPI:
		return nil, fmt.Errorf("--transport api is not implemented yet; " +
			"it needs a backend that can import an uploaded image, which no released Olares ships. " +
			"Use --transport ssh (or local, on the node itself), or push to a registry the node " +
			"can pull from and use `olares-cli cluster workload set-image` directly")
	default:
		return nil, fmt.Errorf("unknown transport %q (want one of: auto, local, ssh, registry, api)", name)
	}
}

func selectAuto(opts Options) (Transport, error) {
	local, localWhy := newLocalTransport(opts)
	if local != nil {
		return local, nil
	}
	ssh, sshWhy := newSSHTransport(opts)
	if ssh != nil {
		return ssh, nil
	}
	return nil, fmt.Errorf(
		"no image transport is available\n"+
			"  local: %s\n"+
			"  ssh:   %s\n"+
			"Configure a node with `olares-cli dev node set --address <host> --user <user>`, "+
			"or push the image to a registry the node can pull from and use "+
			"`olares-cli cluster workload set-image` directly.",
		localWhy, sshWhy)
}

// RemoteOnly reports whether this process is the npm/npx distribution,
// which the Node shim marks with OLARES_CLI_REMOTE_ONLY=1. The local
// transport is meaningless there: the shim only ever runs on a machine
// that is not the Olares node.
func RemoteOnly() bool { return os.Getenv("OLARES_CLI_REMOTE_ONLY") == "1" }

// imageToolCandidates is the probe order for the tool that can produce
// a docker-archive stream from a local image.
var imageToolCandidates = []string{"docker", "podman"}

// resolveImageTool picks the local tool used for `<tool> save`.
func resolveImageTool(override string) (string, error) {
	if override != "" {
		if _, err := exec.LookPath(override); err != nil {
			return "", fmt.Errorf("image tool %q not found on PATH: %w", override, err)
		}
		return override, nil
	}
	for _, candidate := range imageToolCandidates {
		if _, err := exec.LookPath(candidate); err == nil {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("neither docker nor podman found on PATH; " +
		"one of them is needed to export the image (override with --image-tool)",
	)
}

// saveCommand builds the local `<tool> save <image>` whose stdout is a
// docker-archive tar stream — the format `ctr images import` reads.
func saveCommand(ctx context.Context, tool, image string) *exec.Cmd {
	return exec.CommandContext(ctx, tool, "save", image)
}

// importArgs is the remote/local side: unpack the streamed archive into
// the namespace the kubelet reads.
//
// --digests is deliberately omitted: a locally built image has no
// registry digest, and asking ctr to record one makes the import fail
// on some containerd versions rather than degrade.
func importArgs(ctrCmd []string) []string {
	return append(append([]string{}, ctrCmd...),
		"-n", ContainerdNamespace, "images", "import", "-")
}

func progressf(w io.Writer, format string, args ...interface{}) {
	if w == nil {
		return
	}
	fmt.Fprintf(w, format, args...)
}
