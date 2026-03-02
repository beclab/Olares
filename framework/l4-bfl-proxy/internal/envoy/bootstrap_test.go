package envoy

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultBootstrapConfig(t *testing.T) {
	cfg := DefaultBootstrapConfig(8794)

	assert.Equal(t, "l4-bfl-proxy", cfg.NodeID)
	assert.Equal(t, "l4-bfl-proxy", cfg.Cluster)
	assert.Equal(t, "127.0.0.1", cfg.XdsAddress)
	assert.Equal(t, 8794, cfg.XdsPort)
	assert.Equal(t, 19000, cfg.AdminPort)
}

func TestWriteBootstrapConfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "envoy.yaml")

	cfg := &BootstrapConfig{
		NodeID:     "test-node",
		Cluster:    "test-cluster",
		XdsAddress: "127.0.0.1",
		XdsPort:    9999,
		AdminPort:  19000,
	}

	err := WriteBootstrapConfig(path, cfg)
	require.NoError(t, err)

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	content := string(data)

	assert.Contains(t, content, "id: test-node")
	assert.Contains(t, content, "cluster: test-cluster")
	assert.Contains(t, content, "port_value: 9999")
	assert.Contains(t, content, "port_value: 19000")
	assert.Contains(t, content, "address: 127.0.0.1")
	assert.Contains(t, content, "xds_cluster")
	assert.Contains(t, content, "http2_protocol_options")
	assert.Contains(t, content, "api_type: DELTA_GRPC")
	assert.NotContains(t, content, "api_type: GRPC\n")
}

func TestWriteBootstrapConfig_ValidYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "envoy.yaml")

	cfg := DefaultBootstrapConfig(8794)
	err := WriteBootstrapConfig(path, cfg)
	require.NoError(t, err)

	data, err := os.ReadFile(path)
	require.NoError(t, err)

	// Verify it starts with a valid YAML structure (node:).
	content := strings.TrimSpace(string(data))
	assert.True(t, strings.HasPrefix(content, "node:"), "expected YAML to start with 'node:'")
}

func TestWriteBootstrapConfig_BadPath(t *testing.T) {
	err := WriteBootstrapConfig("/nonexistent/dir/envoy.yaml", DefaultBootstrapConfig(8794))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "create bootstrap file")
}

func TestDefaultEnvoyConfig(t *testing.T) {
	// Without ENVOY_BINARY env var, should default to /usr/local/bin/envoy.
	os.Unsetenv("ENVOY_BINARY")
	cfg := DefaultEnvoyConfig()
	assert.Equal(t, "/usr/local/bin/envoy", cfg.BinaryPath)
	assert.Equal(t, "/etc/envoy/envoy.yaml", cfg.BootstrapPath)
}

func TestDefaultEnvoyConfig_CustomBinary(t *testing.T) {
	t.Setenv("ENVOY_BINARY", "/custom/envoy")
	cfg := DefaultEnvoyConfig()
	assert.Equal(t, "/custom/envoy", cfg.BinaryPath)
}
