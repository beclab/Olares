package upgrade

import (
	"bufio"
	"context"
	"encoding/base64"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/beclab/Olares/cli/pkg/core/connector"
)

const dhcpStaged = "/opt/cni/.fixed-mac-v1/dhcp"

type dhcpCommands struct {
	verify func() error
	run    func(string) (string, error)
	ready  func(context.Context, string) (string, error)
}

func runtimeDHCPCommands(runtime connector.Runtime) dhcpCommands {
	run := func(cmd string) (string, error) {
		return runtime.GetRunner().SudoCmd(dhcpPrivilegedCommand(cmd), false, false)
	}
	return dhcpCommands{run: run, verify: func() error {
		command, err := secureDHCPVerifyCommand(runtime.GetSystemInfo().GetOsArch())
		if err != nil {
			return err
		}
		_, err = run(command)
		return err
	}, ready: func(ctx context.Context, binary string) (string, error) {
		return waitDHCPReady(ctx, func() (string, error) {
			// Verify the running executable, not just the pathname replaced on disk.
			if _, err := run("pid=$(systemctl show cni-dhcp -p MainPID --value) && test \"$pid\" -gt 0 && cmp -s " + shellWord(binary) + " /proc/\"$pid\"/exe"); err != nil {
				return "", err
			}
			before, err := daemonGeneration(runtime)
			if err != nil {
				return "", err
			}
			if err = probeDHCPRPC(ctx, "/run/cni/dhcp.sock"); err != nil {
				return "", err
			}
			after, err := daemonGeneration(runtime)
			if err != nil {
				return "", err
			}
			if before != after {
				return "", fmt.Errorf("DHCP restarted during readiness check")
			}
			return after, nil
		}, time.Second, 5*time.Second)
	}}
}

// Negotiate the daemon's HTTP RPC endpoint without allocating or releasing a lease.
// This migration is a LocalTask and uses the existing local service socket.
func probeDHCPRPC(ctx context.Context, socket string) error {
	conn, err := (&net.Dialer{Timeout: 2 * time.Second}).DialContext(ctx, "unix", socket)
	if err != nil {
		return err
	}
	defer conn.Close()
	deadline := time.Now().Add(2 * time.Second)
	if d, ok := ctx.Deadline(); ok && d.Before(deadline) {
		deadline = d
	}
	if err = conn.SetDeadline(deadline); err != nil {
		return err
	}
	if _, err = fmt.Fprint(conn, "CONNECT /_goRPC_ HTTP/1.0\n\n"); err != nil {
		return err
	}
	response, err := http.ReadResponse(bufio.NewReader(conn), &http.Request{Method: "CONNECT"})
	if err != nil {
		return err
	}
	if response.Status != "200 Connected to Go RPC" {
		return fmt.Errorf("unexpected DHCP RPC response: %s", response.Status)
	}
	return nil
}

func waitDHCPReady(ctx context.Context, probe func() (string, error), interval, stable time.Duration) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	var generation string
	var since time.Time
	var last error
	for {
		current, err := probe()
		if err != nil || current == "" {
			generation = ""
			last = err
		} else {
			if current != generation {
				generation, since = current, time.Now()
			}
			if time.Since(since) >= stable {
				return generation, nil
			}
		}
		select {
		case <-ctx.Done():
			return "", fmt.Errorf("DHCP did not remain ready: %w (last probe: %v)", ctx.Err(), last)
		case <-time.After(interval):
		}
	}
}

// Apply the target version and converge forward, like other component upgrades.
// Persist intent before replacing the executable. On failure, leave prepared
// progress for the task runner to retry; never restore the old binary.
func upgradeDHCPBinary(ctx context.Context, p *fixedMACProgress, cmd dhcpCommands, save func(*fixedMACProgress) error) error {
	if p.DHCPPhase == "" || p.DHCPPhase == "topology" {
		p.DHCPPhase = "prepared"
		if err := save(p); err != nil {
			return err
		}
	}
	if p.DHCPPhase != "prepared" {
		return fmt.Errorf("unexpected DHCP transaction phase %q", p.DHCPPhase)
	}
	if cmd.verify == nil {
		return fmt.Errorf("DHCP integrity verifier is required")
	}
	if err := cmd.verify(); err != nil {
		return fmt.Errorf("verify staged DHCP: %w", err)
	}
	// Handles a crash after restart but before recording its invocation ID.
	probeCtx, cancel := context.WithTimeout(ctx, 6*time.Second)
	generation, err := cmd.ready(probeCtx, dhcpStaged)
	cancel()
	if err != nil {
		// Copy into the destination directory so rename cannot cross filesystems.
		_, err = cmd.run("install -m 0755 " + dhcpStaged + " /opt/cni/bin/.dhcp.fixed-mac-v1.next && cmp -s " + dhcpStaged + " /opt/cni/bin/.dhcp.fixed-mac-v1.next && sync -f /opt/cni/bin/.dhcp.fixed-mac-v1.next && mv -fT /opt/cni/bin/.dhcp.fixed-mac-v1.next /opt/cni/bin/dhcp && sync -f /opt/cni/bin")
		if err == nil {
			_, err = cmd.run("timeout 60s systemctl restart cni-dhcp")
		}
		if err == nil {
			generation, err = cmd.ready(ctx, dhcpStaged)
		}
		if err != nil {
			return fmt.Errorf("apply target DHCP version: %w; upgrade remains incomplete and will resume forward on retry", err)
		}
	}
	p.DHCPPhase, p.DaemonPID = "ready", generation
	return save(p)
}

func secureDHCPExtractCommand(archive string) string {
	return "rm -f " + dhcpStaged + ".new && (umask 077; tar -xOzf " + shellWord(archive) + " ./dhcp > " + dhcpStaged + ".new) && chown 0:0 " + dhcpStaged + ".new && chmod 0500 " + dhcpStaged + ".new && mv -fT " + dhcpStaged + ".new " + dhcpStaged
}

func secureDHCPVerifyCommand(arch string) (string, error) {
	hashes := map[string]string{
		"amd64": "9662e410096d0019eb0a12c172ca5da7e0414a079115322843b61e0cd0395327",
		"arm64": "8e0c60d5158b4478744cc8bfd13ef6d5a34b26a34aee04c8c0e52aa2c0fc61e3",
	}
	hash, ok := hashes[arch]
	if !ok {
		return "", fmt.Errorf("unsupported DHCP architecture %q", arch)
	}
	return "test ! -L /opt/cni/.fixed-mac-v1 && test \"$(stat -c '%u:%g:%a' /opt/cni/.fixed-mac-v1)\" = '0:0:700' && test -f " + dhcpStaged + " && test ! -L " + dhcpStaged + " && test \"$(stat -c '%u:%g:%a' " + dhcpStaged + ")\" = '0:0:500' && printf '%s  %s\\n' " + shellWord(hash) + " " + shellWord(dhcpStaged) + " | sha256sum -c -", nil
}

// This is called only by an unfinished upgrade, not by a background watchdog.
// A query error is not evidence of a stopped daemon and must never cause a start.
func ensureUpgradeDHCPRunning(ctx context.Context, cmd dhcpCommands) (string, error) {
	state, err := cmd.run("systemctl show cni-dhcp --property=ActiveState --value")
	if err != nil {
		return "", fmt.Errorf("query DHCP service: %w", err)
	}
	switch strings.TrimSpace(state) {
	case "active":
		if cmd.verify == nil {
			return "", fmt.Errorf("DHCP integrity verifier is required")
		}
		if err := cmd.verify(); err != nil {
			return "", err
		}
		return cmd.ready(ctx, dhcpStaged)
	case "inactive", "failed":
		if cmd.verify == nil {
			return "", fmt.Errorf("DHCP integrity verifier is required")
		}
		if err := cmd.verify(); err != nil {
			return "", err
		}
		if _, err := cmd.run("cmp -s " + dhcpStaged + " /opt/cni/bin/dhcp && timeout 60s systemctl start cni-dhcp"); err != nil {
			return "", err
		}
		return cmd.ready(ctx, dhcpStaged)
	default:
		return "", fmt.Errorf("DHCP service is %q; retry after the transition completes", strings.TrimSpace(state))
	}
}

// Keep expansions inside the privileged shell even when the runner adds sudo's
// double-quoted bash wrapper for a non-root connection user.
func dhcpPrivilegedCommand(command string) string {
	return "printf %s " + shellWord(base64.StdEncoding.EncodeToString([]byte(command))) + " | base64 -d | /bin/bash"
}
