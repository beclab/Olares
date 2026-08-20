package devdeploy

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

// containerdSockets are the socket paths probed to decide whether this
// process is running on an Olares node, in the order Olares itself
// prefers: k3s is the default KUBE_TYPE, so its socket is checked
// first, and a host running both should be driven through k3s.
//
// k3sOwned marks a socket that only exists on a k3s node and is
// therefore self-proving. The generic containerd paths are NOT: a
// plain Docker install puts its own containerd at exactly
// /run/containerd/containerd.sock (with /var/run symlinked to /run), so
// finding a socket there says nothing about whether this machine runs
// Kubernetes. Importing into that containerd would "succeed" and put
// the image somewhere no kubelet will ever look.
var containerdSockets = []struct {
	path     string
	ctr      []string
	k3sOwned bool
}{
	{"/run/k3s/containerd/containerd.sock", []string{"k3s", "ctr"}, true},
	{"/var/run/containerd/containerd.sock", []string{"ctr"}, false},
	{"/run/containerd/containerd.sock", []string{"ctr"}, false},
}

// kubeletMarkers are paths that exist on a machine running a kubelet
// and essentially nowhere else. One of them must be present before an
// ambiguous containerd socket is accepted as "this is the node".
var kubeletMarkers = []string{
	"/etc/rancher/k3s",
	"/var/lib/kubelet",
	"/etc/kubernetes",
}

func hasKubeletMarker() bool {
	for _, path := range kubeletMarkers {
		if _, err := os.Stat(path); err == nil {
			return true
		}
	}
	return false
}

// localTransport imports straight into this machine's containerd. This
// is the LocalTask half of the install engine's local/remote split
// (pkg/core/task/local_task.go): a plain child process, no SSH. The
// engine's Dialer has no local shortcut — it would SSH to localhost —
// so a host-mode push must not go through it.
type localTransport struct {
	tool   string   // docker | podman
	ctr    []string // ["k3s","ctr"] or ["ctr"]
	sock   string
	sudo   bool
	socket string
}

// newLocalTransport returns nil plus a human-readable reason when the
// local transport cannot be used here.
func newLocalTransport(opts Options) (Transport, string) {
	if RemoteOnly() {
		return nil, "this is the npm/npx build, which never runs on the Olares node itself"
	}
	var (
		sock         string
		ctr          []string
		sawAmbiguous bool
	)
	for _, candidate := range containerdSockets {
		fi, err := os.Stat(candidate.path)
		if err != nil || fi.Mode()&os.ModeSocket == 0 {
			continue
		}
		// A socket is only useful if the matching ctr binary is
		// actually installed; k3s ships its own.
		if _, err := exec.LookPath(candidate.ctr[0]); err != nil {
			continue
		}
		if !candidate.k3sOwned && !hasKubeletMarker() {
			// Almost certainly a developer machine's Docker daemon.
			sawAmbiguous = true
			continue
		}
		sock, ctr = candidate.path, candidate.ctr
		break
	}
	if sock == "" {
		if sawAmbiguous {
			return nil, "found a containerd socket but no kubelet on this machine — " +
				"that containerd belongs to a local Docker/containerd install, not an Olares node"
		}
		return nil, "no containerd socket found (not an Olares node, or containerd is not running)"
	}
	tool, err := resolveImageTool(opts.ImageTool)
	if err != nil {
		return nil, err.Error()
	}
	return &localTransport{
		tool:   tool,
		ctr:    ctr,
		sock:   sock,
		socket: sock,
		sudo:   os.Geteuid() != 0,
	}, ""
}

func (t *localTransport) Name() string { return TransportLocal }

func (t *localTransport) Describe() string {
	prefix := ""
	if t.sudo {
		prefix = "sudo "
	}
	return fmt.Sprintf("%s save | %s%s -n %s images import - (socket %s)",
		t.tool, prefix, strings.Join(t.ctr, " "), ContainerdNamespace, t.sock)
}

func (t *localTransport) Import(ctx context.Context, image string, progress io.Writer) error {
	save := saveCommand(ctx, t.tool, image)
	saveOut, err := save.StdoutPipe()
	if err != nil {
		return fmt.Errorf("pipe %s save: %w", t.tool, err)
	}
	var saveErr strings.Builder
	save.Stderr = &saveErr

	args := importArgs(t.ctr)
	name := args[0]
	rest := args[1:]
	if t.sudo {
		// -E preserves the environment (notably CONTAINERD_ADDRESS if
		// the operator set one); -n keeps sudo from silently blocking
		// on a password prompt inside a pipeline, where the prompt
		// would be invisible and look like a hang.
		rest = append([]string{"-n", "-E", name}, rest...)
		name = "sudo"
	}
	imp := exec.CommandContext(ctx, name, rest...)
	imp.Stdin = saveOut
	var impErr strings.Builder
	imp.Stderr = &impErr
	imp.Stdout = progress

	progressf(progress, "importing %s into containerd (%s)\n", image, t.sock)

	if err := save.Start(); err != nil {
		return fmt.Errorf("start %s save %s: %w", t.tool, image, err)
	}
	if err := imp.Start(); err != nil {
		_ = save.Process.Kill()
		_ = save.Wait()
		return fmt.Errorf("start ctr import: %w", err)
	}

	// Wait on the importer first: if it dies the writer gets EPIPE and
	// unblocks. Waiting on the producer first would deadlock whenever
	// the importer exits early (bad archive, permission denied).
	impWaitErr := imp.Wait()
	saveWaitErr := save.Wait()

	// Report the producer's failure first even though it is reaped
	// second. When `docker save` fails (typically "no such image") it
	// closes the pipe having written nothing, and ctr then fails with
	// "unrecognized image format" — a downstream symptom that says
	// nothing about the actual mistake.
	if saveWaitErr != nil {
		msg := strings.TrimSpace(saveErr.String())
		if msg == "" {
			msg = saveWaitErr.Error()
		}
		return fmt.Errorf("%s save %s failed: %s", t.tool, image, msg)
	}
	if impWaitErr != nil {
		msg := strings.TrimSpace(impErr.String())
		if strings.Contains(msg, "permission denied") || strings.Contains(msg, "sudo:") {
			return fmt.Errorf("ctr import failed with a permission error (%s); "+
				"containerd's socket is root-owned — re-run as root or configure passwordless sudo", msg)
		}
		if msg == "" {
			msg = impWaitErr.Error()
		}
		return fmt.Errorf("ctr import failed: %s", msg)
	}
	return nil
}
