package devdeploy

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/beclab/Olares/cli/pkg/cliconfig"
	"github.com/beclab/Olares/cli/pkg/core/connector"
)

// sshDialTimeout matches the masterSSHTimeout used by
// pkg/pipelines/join_credentials.go so a wrong address fails at the
// same speed everywhere in the CLI.
const sshDialTimeout = 10 * time.Second

// sshTransport streams the image over SSH into the node's containerd.
//
// It uses connector.Connection.PExec, whose stdin parameter lets the
// local `docker save` pipe straight into the remote `ctr images import
// -` with no temp file on either side. Scp-then-import would need
// (image size) of free disk on the node and a cleanup path for the
// half-written file when the transfer is interrupted.
type sshTransport struct {
	tool string
	node cliconfig.DevNodeConfig
}

func newSSHTransport(opts Options) (Transport, string) {
	if opts.Node == nil || strings.TrimSpace(opts.Node.Address) == "" {
		return nil, "no node configured for this profile (run `olares-cli dev node set`)"
	}
	if strings.TrimSpace(opts.Node.User) == "" {
		return nil, "the configured node has no SSH user (re-run `olares-cli dev node set --user <user>`)"
	}
	tool, err := resolveImageTool(opts.ImageTool)
	if err != nil {
		return nil, err.Error()
	}
	return &sshTransport{tool: tool, node: *opts.Node}, ""
}

func (t *sshTransport) Name() string { return TransportSSH }

func (t *sshTransport) Describe() string {
	port := t.node.Port
	if port == 0 {
		port = 22
	}
	return fmt.Sprintf("%s save | ssh %s@%s:%d 'ctr -n %s images import -'",
		t.tool, t.node.User, t.node.Address, port, ContainerdNamespace)
}

// connect builds the standalone connection + host pair, mirroring
// verifyMasterSSH in pkg/pipelines/join_credentials.go. Password auth is
// intentionally unsupported (see cliconfig.DevNodeConfig).
func (t *sshTransport) connect() (connector.Connection, connector.Host, error) {
	host := connector.NewHost()
	host.SetAddress(t.node.Address)
	host.SetInternalAddress(t.node.Address)
	host.SetPort(t.portOrDefault())
	host.SetUser(t.node.User)
	host.SetPrivateKeyPath(t.node.PrivateKeyPath)

	cfg := connector.Cfg{
		Username: t.node.User,
		Address:  t.node.Address,
		Port:     t.portOrDefault(),
		KeyFile:  t.node.PrivateKeyPath,
		Timeout:  sshDialTimeout,
	}
	// With no explicit key, fall back to the agent rather than failing
	// validateOptions' "must specify at least one auth method" check.
	if cfg.KeyFile == "" {
		if os.Getenv("SSH_AUTH_SOCK") == "" {
			return nil, nil, fmt.Errorf(
				"no SSH key configured for the node and no agent running (SSH_AUTH_SOCK unset); " +
					"re-run `olares-cli dev node set --private-key <path>`")
		}
		cfg.AgentSocket = "env:SSH_AUTH_SOCK"
	}

	conn, err := connector.NewConnection(cfg)
	if err != nil {
		return nil, nil, fmt.Errorf("ssh to %s@%s: %w", t.node.User, t.node.Address, err)
	}
	return conn, host, nil
}

func (t *sshTransport) portOrDefault() int {
	if t.node.Port == 0 {
		return 22
	}
	return t.node.Port
}

func (t *sshTransport) Import(ctx context.Context, image string, progress io.Writer) error {
	conn, host, err := t.connect()
	if err != nil {
		return err
	}
	defer conn.Close()

	remoteCtr, err := t.probeCtr(conn, host)
	if err != nil {
		return err
	}

	save := saveCommand(ctx, t.tool, image)
	saveOut, err := save.StdoutPipe()
	if err != nil {
		return fmt.Errorf("pipe %s save: %w", t.tool, err)
	}
	var saveErr strings.Builder
	save.Stderr = &saveErr
	if err := save.Start(); err != nil {
		return fmt.Errorf("start %s save %s: %w", t.tool, image, err)
	}

	cmd := strings.Join(importArgs(remoteCtr), " ")
	cmd = host.SudoPrefixIfNecessary(cmd)
	progressf(progress, "streaming %s to %s@%s (%s)\n", image, t.node.User, t.node.Address, cmd)

	var remoteErr strings.Builder
	code, execErr := conn.PExec(cmd, saveOut, progress, &remoteErr, host)

	// Reap the producer regardless of how the remote side ended, so a
	// failed import doesn't leave an orphaned `docker save`.
	saveWaitErr := save.Wait()

	// Local producer errors are reported first: a failed `docker save`
	// (usually "no such image") closes the pipe having written nothing,
	// and the remote ctr then fails with "unrecognized image format",
	// which describes the symptom rather than the mistake.
	if saveWaitErr != nil {
		msg := strings.TrimSpace(saveErr.String())
		if msg == "" {
			msg = saveWaitErr.Error()
		}
		return fmt.Errorf("%s save %s failed: %s", t.tool, image, msg)
	}
	if execErr != nil {
		return fmt.Errorf("remote import failed: %w (%s)", execErr, strings.TrimSpace(remoteErr.String()))
	}
	if code != 0 {
		msg := strings.TrimSpace(remoteErr.String())
		if msg == "" {
			msg = fmt.Sprintf("exit status %d", code)
		}
		return fmt.Errorf("remote `%s` failed: %s", cmd, msg)
	}
	return nil
}

// probeCtr asks the node which ctr it has rather than assuming k3s.
// KUBE_TYPE can be k8s, in which case there is no `k3s` binary and the
// standalone `ctr` is the right entry point.
func (t *sshTransport) probeCtr(conn connector.Connection, host connector.Host) ([]string, error) {
	out, _, err := conn.Exec("command -v k3s >/dev/null 2>&1 && echo k3s || (command -v ctr >/dev/null 2>&1 && echo ctr || echo none)", host)
	if err != nil {
		return nil, fmt.Errorf("probe for ctr on %s: %w", t.node.Address, err)
	}
	switch strings.TrimSpace(out) {
	case "k3s":
		return []string{"k3s", "ctr"}, nil
	case "ctr":
		return []string{"ctr"}, nil
	default:
		return nil, fmt.Errorf("neither `k3s` nor `ctr` found on %s — is this an Olares node?", t.node.Address)
	}
}
