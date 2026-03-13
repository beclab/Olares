package envoy

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"text/template"

	"k8s.io/klog/v2"
)

const bootstrapTemplate = `node:
  id: {{ .NodeID }}
  cluster: {{ .Cluster }}
dynamic_resources:
  lds_config:
    resource_api_version: V3
    api_config_source:
      api_type: DELTA_GRPC
      transport_api_version: V3
      grpc_services:
      - envoy_grpc:
          cluster_name: xds_cluster
  cds_config:
    resource_api_version: V3
    api_config_source:
      api_type: DELTA_GRPC
      transport_api_version: V3
      grpc_services:
      - envoy_grpc:
          cluster_name: xds_cluster
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
}

func DefaultEnvoyConfig() *EnvoyConfig {
	binary := os.Getenv("ENVOY_BINARY")
	if binary == "" {
		binary = "/usr/local/bin/envoy"
	}
	return &EnvoyConfig{
		BinaryPath:    binary,
		BootstrapPath: "/etc/envoy/envoy.yaml",
	}
}

func StartEnvoy(ctx context.Context, cancel context.CancelFunc, cfg *EnvoyConfig) error {
	args := []string{
		"-c", cfg.BootstrapPath,
		"--service-cluster", "l4-bfl-proxy",
		"--service-node", "l4-bfl-proxy",
		"--log-level", "info",
	}

	cmd := exec.CommandContext(ctx, cfg.BinaryPath, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	klog.Infof("envoy: starting %s %v", cfg.BinaryPath, args)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start envoy: %w", err)
	}

	go func() {
		if err := cmd.Wait(); err != nil {
			klog.Errorf("envoy: process exited with error: %v", err)
		} else {
			klog.Warning("envoy: process exited unexpectedly")
		}
		cancel()
	}()
	return nil
}
