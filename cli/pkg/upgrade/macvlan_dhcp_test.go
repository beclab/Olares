package upgrade

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os/exec"
	"strings"
	"testing"
	"time"
)

func TestDHCPTransactionCrashAfterRestartDoesNotRestartAgain(t *testing.T) {
	p := &fixedMACProgress{DHCPPhase: "prepared"}
	commands := 0
	cmd := dhcpCommands{verify: func() error { return nil }, run: func(string) (string, error) { commands++; return "", nil }, ready: func(context.Context, string) (string, error) { return "new", nil }}
	err := upgradeDHCPBinary(t.Context(), p, cmd, func(*fixedMACProgress) error { return nil })
	if err != nil || commands != 0 || p.DHCPPhase != "ready" || p.DaemonPID != "new" {
		t.Fatalf("err=%v commands=%d progress=%+v", err, commands, p)
	}
}
func TestDHCPTransactionSaveFailurePreventsReplacement(t *testing.T) {
	cmd := dhcpCommands{verify: func() error { return nil }, run: func(string) (string, error) { t.Fatal("mutation before durable intent"); return "", nil }}
	err := upgradeDHCPBinary(t.Context(), &fixedMACProgress{}, cmd, func(*fixedMACProgress) error { return errors.New("disk full") })
	if err == nil {
		t.Fatal("expected save failure")
	}
}

func TestDHCPFailuresResumeForward(t *testing.T) {
	for _, failure := range []string{"install", "restart", "readiness"} {
		t.Run(failure, func(t *testing.T) {
			p := &fixedMACProgress{}
			durablePhase := ""
			fail := true
			running := false
			cmd := dhcpCommands{verify: func() error { return nil },
				run: func(s string) (string, error) {
					if strings.Contains(s, "previous") || strings.Contains(s, "restore") {
						t.Fatalf("unexpected rollback: %s", s)
					}
					if durablePhase != "prepared" {
						t.Fatal("mutation without durable intent")
					}
					if strings.Contains(s, "install -m") && fail && failure == "install" {
						return "", errors.New("install failed")
					}
					if strings.Contains(s, "systemctl restart") {
						if fail && failure == "restart" {
							return "", errors.New("restart failed")
						}
						running = true
					}
					return "", nil
				},
				ready: func(context.Context, string) (string, error) {
					if !running || (fail && failure == "readiness") {
						return "", errors.New("not ready")
					}
					return "new", nil
				},
			}
			save := func(p *fixedMACProgress) error { durablePhase = p.DHCPPhase; return nil }
			if err := upgradeDHCPBinary(t.Context(), p, cmd, save); err == nil || p.DHCPPhase != "prepared" || p.Complete {
				t.Fatalf("err=%v progress=%+v", err, p)
			}
			fail = false
			if err := upgradeDHCPBinary(t.Context(), p, cmd, save); err != nil || p.DHCPPhase != "ready" || p.DaemonPID != "new" || p.Complete {
				t.Fatalf("err=%v progress=%+v", err, p)
			}
		})
	}
}

func TestDHCPReadinessRejectsChangingInvocation(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Millisecond)
	defer cancel()
	n := 0
	_, err := waitDHCPReady(ctx, func() (string, error) { n++; return fmt.Sprint(n), nil }, time.Millisecond, 10*time.Millisecond)
	if err == nil {
		t.Fatal("restarting service accepted")
	}
}
func TestDHCPRPCProbe(t *testing.T) {
	for _, respond := range []bool{false, true} {
		t.Run(fmt.Sprint(respond), func(t *testing.T) {
			path := t.TempDir() + "/dhcp.sock"
			listener, err := net.Listen("unix", path)
			if err != nil {
				t.Fatal(err)
			}
			defer listener.Close()
			if respond {
				go func() {
					conn, e := listener.Accept()
					if e != nil {
						return
					}
					defer conn.Close()
					_ = conn.SetDeadline(time.Now().Add(time.Second))
					var b [128]byte
					_, _ = conn.Read(b[:])
					_, _ = fmt.Fprint(conn, "HTTP/1.0 200 Connected to Go RPC\n\n")
				}()
			}
			ctx, cancel := context.WithTimeout(t.Context(), 100*time.Millisecond)
			defer cancel()
			err = probeDHCPRPC(ctx, path)
			if (err == nil) != respond {
				t.Fatalf("response=%v err=%v", respond, err)
			}
		})
	}
}

func TestDHCPTransactionSuccessfulReplacement(t *testing.T) {
	p := &fixedMACProgress{}
	restarted := false
	cmd := dhcpCommands{verify: func() error { return nil },
		run: func(s string) (string, error) {
			if strings.Contains(s, "systemctl restart") {
				restarted = true
			}
			return "", nil
		},
		ready: func(context.Context, string) (string, error) {
			if restarted {
				return "new", nil
			}
			return "", errors.New("old binary")
		},
	}
	var phases []string
	err := upgradeDHCPBinary(t.Context(), p, cmd, func(p *fixedMACProgress) error { phases = append(phases, p.DHCPPhase); return nil })
	if err != nil || !restarted || strings.Join(phases, ",") != "prepared,ready" || p.DaemonPID != "new" {
		t.Fatalf("err=%v phases=%v progress=%+v", err, phases, p)
	}
}

func TestDHCPIntegrityFailurePreventsExecution(t *testing.T) {
	cmd := dhcpCommands{verify: func() error { return errors.New("tampered binary") }, run: func(string) (string, error) { t.Fatal("executed before integrity verification"); return "", nil }, ready: func(context.Context, string) (string, error) {
		t.Fatal("probed before integrity verification")
		return "", nil
	}}
	if err := upgradeDHCPBinary(t.Context(), &fixedMACProgress{DHCPPhase: "prepared"}, cmd, func(*fixedMACProgress) error { return nil }); err == nil {
		t.Fatal("tampered artifact accepted")
	}
}

func TestDHCPUpgradeStartsOnlyConfirmedStoppedService(t *testing.T) {
	for _, state := range []string{"active", "inactive", "failed", "activating", "query-error"} {
		t.Run(state, func(t *testing.T) {
			started := false
			cmd := dhcpCommands{verify: func() error { return nil }, run: func(s string) (string, error) {
				if strings.Contains(s, "systemctl show") {
					if state == "query-error" {
						return "", errors.New("permission denied")
					}
					return state, nil
				}
				if !strings.Contains(s, "systemctl start") || strings.Contains(s, "restart") {
					t.Fatalf("unexpected command %s", s)
				}
				started = true
				return "", nil
			}, ready: func(context.Context, string) (string, error) { return "invocation", nil }}
			_, err := ensureUpgradeDHCPRunning(t.Context(), cmd)
			if started != (state == "inactive" || state == "failed") {
				t.Fatalf("state=%s started=%v", state, started)
			}
			if (err != nil) != (state == "activating" || state == "query-error") {
				t.Fatalf("state=%s err=%v", state, err)
			}
		})
	}
}

func TestDHCPPrivilegedCommandPreservesInnerExpansions(t *testing.T) {
	command := `pid=123; printf '%s' "$pid"`
	wrapped := dhcpPrivilegedCommand(command)
	// Emulate the runner's outer double-quoted shell without executing sudo.
	out, err := exec.Command("bash", "-c", `bash -c "`+wrapped+`"`).CombinedOutput()
	if err != nil || string(out) != "123" {
		t.Fatalf("output=%q error=%v", out, err)
	}
}

func TestDHCPTransactionAcceptsCompletedTopologyPhase(t *testing.T) {
	p := &fixedMACProgress{DHCPPhase: "topology"}
	var phases []string
	cmd := dhcpCommands{verify: func() error { return nil }, ready: func(context.Context, string) (string, error) { return "target", nil }, run: func(string) (string, error) { t.Fatal("already running target must not restart"); return "", nil }}
	if err := upgradeDHCPBinary(t.Context(), p, cmd, func(p *fixedMACProgress) error { phases = append(phases, p.DHCPPhase); return nil }); err != nil {
		t.Fatal(err)
	}
	if strings.Join(phases, ",") != "prepared,ready" || p.DaemonPID != "target" {
		t.Fatalf("invalid transition: %v %+v", phases, p)
	}
}
