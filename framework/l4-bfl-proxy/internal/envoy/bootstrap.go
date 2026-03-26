package envoy

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync/atomic"
	"syscall"
	"text/template"
	"time"

	"k8s.io/klog/v2"
)

const bootstrapTemplate = `node:
  id: {{ .NodeID }}
  cluster: {{ .Cluster }}
dynamic_resources:
  ads_config:
    api_type: DELTA_GRPC
    transport_api_version: V3
    grpc_services:
    - envoy_grpc:
        cluster_name: xds_cluster
  lds_config:
    resource_api_version: V3
    ads: {}
  cds_config:
    resource_api_version: V3
    ads: {}
static_resources:
  clusters:
  - name: xds_cluster
    connect_timeout: 5s
    type: STATIC
    load_assignment:
      cluster_name: xds_cluster
      endpoints:
      - lb_endpoints:
        - endpoint:
            address:
              socket_address:
                address: {{ .XdsAddress }}
                port_value: {{ .XdsPort }}
    typed_extension_protocol_options:
      envoy.extensions.upstreams.http.v3.HttpProtocolOptions:
        "@type": type.googleapis.com/envoy.extensions.upstreams.http.v3.HttpProtocolOptions
        explicit_http_config:
          http2_protocol_options: {}
admin:
  address:
    socket_address:
      address: 127.0.0.1
      port_value: {{ .AdminPort }}
`

type BootstrapConfig struct {
	NodeID     string
	Cluster    string
	XdsAddress string
	XdsPort    int
	AdminPort  int
}

func DefaultBootstrapConfig(xdsPort int) *BootstrapConfig {
	return &BootstrapConfig{
		NodeID:     "l4-bfl-proxy",
		Cluster:    "l4-bfl-proxy",
		XdsAddress: "127.0.0.1",
		XdsPort:    xdsPort,
		AdminPort:  19000,
	}
}

func WriteBootstrapConfig(path string, cfg *BootstrapConfig) error {
	tmpl, err := template.New("bootstrap").Parse(bootstrapTemplate)
	if err != nil {
		return fmt.Errorf("parse bootstrap template: %w", err)
	}
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create bootstrap file: %w", err)
	}
	defer f.Close()
	return tmpl.Execute(f, cfg)
}

type EnvoyConfig struct {
	BinaryPath    string
	BootstrapPath string
	// AdminAddr is the address of the Envoy admin API, e.g. "127.0.0.1:19000".
	// It is used to initiate graceful drain before sending SIGTERM.
	AdminAddr string
}

func DefaultEnvoyConfig() *EnvoyConfig {
	binary := os.Getenv("ENVOY_BINARY")
	if binary == "" {
		binary = "/usr/local/bin/envoy"
	}
	return &EnvoyConfig{
		BinaryPath:    binary,
		BootstrapPath: "/etc/envoy/envoy.yaml",
		AdminAddr:     "127.0.0.1:19000",
	}
}

// adminHTTPClient is used exclusively for Envoy admin API requests during shutdown.
var adminHTTPClient = &http.Client{Timeout: 3 * time.Second}

const (
	// envoyPreStopDelay is the time to wait after the context is cancelled
	// before initiating any drain. During this window Kubernetes propagates
	// the endpoint removal through kube-proxy/IPVS, ensuring no new connections
	// are routed to this pod while Envoy is still accepting them.
	envoyPreStopDelay = 5 * time.Second

	// envoyDrainTimeout is the maximum time to wait for active connections to
	// close after /drain_listeners?graceful is sent to the Envoy admin API.
	// HTTP/1.1 connections drain on the next response; WebSocket connections
	// stay alive until this timeout expires.
	// Kubernetes terminationGracePeriodSeconds should be >
	// envoyPreStopDelay + envoyDrainTimeout + envoyKillTimeout (recommend >= 40s).
	envoyDrainTimeout = 25 * time.Second

	// envoyKillTimeout is the additional time to wait for the Envoy process to
	// exit after SIGTERM is sent (post-drain). If it does not exit within this
	// window a SIGKILL is sent.
	envoyKillTimeout = 5 * time.Second
)

// StartEnvoy starts the Envoy process and returns a channel that is closed when
// the process has fully exited. The caller should wait on this channel before
// returning from main() so that the graceful drain sequence (SIGTERM →
// envoyDrainTimeout → SIGKILL) is allowed to complete before the Go process
// itself exits.
func StartEnvoy(ctx context.Context, cancel context.CancelFunc, cfg *EnvoyConfig) (<-chan struct{}, error) {
	args := []string{
		"-c", cfg.BootstrapPath,
		"--service-cluster", "l4-bfl-proxy",
		"--service-node", "l4-bfl-proxy",
		"--log-level", "info",
	}

	// Use plain exec.Command so we control the shutdown signal ourselves.
	// exec.CommandContext would send SIGKILL immediately on context cancellation,
	// preventing Envoy from draining in-flight connections gracefully.
	cmd := exec.Command(cfg.BinaryPath, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	klog.Infof("envoy: starting %s %v", cfg.BinaryPath, args)
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("start envoy: %w", err)
	}

	// exitCh is closed once cmd.Wait() returns, i.e. the Envoy process has
	// fully exited. It is returned to the caller so that main() can block on
	// it and avoid exiting before the drain completes.
	exitCh := make(chan struct{})

	// sigTermSent tracks whether we initiated the shutdown so that the Wait
	// goroutine can emit the appropriate log message.
	var sigTermSent atomic.Bool

	// Monitor the process: if it exits on its own, propagate the cancellation
	// upward so the whole control plane shuts down.
	go func() {
		defer close(exitCh)
		if err := cmd.Wait(); err != nil {
			klog.Errorf("envoy: process exited with error: %v", err)
		} else if sigTermSent.Load() {
			klog.Info("envoy: process exited gracefully after drain")
		} else {
			klog.Warning("envoy: process exited unexpectedly")
		}
		cancel()
	}()

	// Graceful shutdown goroutine:
	//   1. Wait for context cancellation (SIGTERM from Kubernetes).
	//   2. Sleep envoyPreStopDelay to allow Kubernetes endpoint removal to
	//      propagate so no new connections are routed to this pod.
	//   3. POST /drain_listeners?graceful to the Envoy admin API. This stops
	//      new connection acceptance while keeping existing connections alive:
	//      HTTP/1.1 gets "Connection: close" on the next response; WebSocket /
	//      upgraded connections remain open until they close naturally or the
	//      drain timeout is reached. This is fundamentally different from
	//      SIGTERM, which immediately closes all connections.
	//   4. Poll active connection count until 0 or envoyDrainTimeout elapses.
	//   5. Send SIGTERM so the process exits cleanly (connections are already
	//      drained at this point, so exit is near-instant).
	//   6. If the process does not exit within envoyKillTimeout, send SIGKILL.
	go func() {
		select {
		case <-ctx.Done():
		case <-exitCh:
			return
		}

		klog.Infof("envoy: context cancelled, waiting %s for endpoint propagation before drain", envoyPreStopDelay)
		select {
		case <-time.After(envoyPreStopDelay):
		case <-exitCh:
			return
		}

		klog.Info("envoy: initiating graceful drain via admin API")
		if err := envoyAdminDrain(cfg.AdminAddr); err != nil {
			klog.Errorf("envoy: admin drain failed: %v; falling back to immediate SIGTERM", err)
		} else {
			klog.Infof("envoy: waiting up to %s for connections to drain", envoyDrainTimeout)
			deadline := time.Now().Add(envoyDrainTimeout)
			for time.Now().Before(deadline) {
				select {
				case <-exitCh:
					return
				case <-time.After(500 * time.Millisecond):
				}
				n, err := envoyActiveConnections(cfg.AdminAddr)
				if err != nil {
					klog.V(4).Infof("envoy: could not read active connections: %v", err)
					continue
				}
				klog.V(4).Infof("envoy: %d active connection(s)", n)
				if n == 0 {
					klog.Info("envoy: all connections drained")
					break
				}
			}
		}

		klog.Info("envoy: sending SIGTERM to terminate process")
		sigTermSent.Store(true)
		if err := cmd.Process.Signal(syscall.SIGTERM); err != nil {
			klog.Errorf("envoy: send SIGTERM: %v", err)
			return
		}
		select {
		case <-exitCh:
			// logged by the Wait goroutine
		case <-time.After(envoyKillTimeout):
			klog.Warningf("envoy: process still alive %s after SIGTERM, sending SIGKILL", envoyKillTimeout)
			if err := cmd.Process.Kill(); err != nil {
				klog.Errorf("envoy: send SIGKILL: %v", err)
			}
		}
	}()

	return exitCh, nil
}

// envoyAdminDrain calls the Envoy admin API to initiate a graceful listener
// drain. Unlike SIGTERM, this keeps existing connections alive while stopping
// new connection acceptance.
func envoyAdminDrain(adminAddr string) error {
	resp, err := adminHTTPClient.Post(
		fmt.Sprintf("http://%s/drain_listeners?graceful", adminAddr),
		"application/json", nil)
	if err != nil {
		return fmt.Errorf("drain_listeners: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("drain_listeners returned HTTP %d", resp.StatusCode)
	}
	return nil
}

// envoyActiveConnections queries the Envoy admin stats API and returns the
// total number of currently active downstream connections across all listeners.
func envoyActiveConnections(adminAddr string) (int, error) {
	resp, err := adminHTTPClient.Get(
		fmt.Sprintf("http://%s/stats?filter=downstream_cx_active&usedonly", adminAddr))
	if err != nil {
		return 0, fmt.Errorf("stats: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("read stats body: %w", err)
	}
	total := 0
	for _, line := range strings.Split(string(body), "\n") {
		if !strings.Contains(line, "downstream_cx_active") {
			continue
		}
		parts := strings.SplitN(line, ": ", 2)
		if len(parts) != 2 {
			continue
		}
		n, err := strconv.Atoi(strings.TrimSpace(parts[1]))
		if err == nil {
			total += n
		}
	}
	return total, nil
}
