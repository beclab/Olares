package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/beclab/Olares/framework/app-service/pkg/security"
	"sigs.k8s.io/yaml"
)

var exports = []struct {
	filename string
	ns       string
	peerNS   string
}{
	{filename: "app-gateway-mesh-np-linkerd.yaml", ns: "linkerd", peerNS: "os-gateway"},
	{filename: "app-gateway-mesh-np-os-gateway.yaml", ns: "os-gateway", peerNS: "linkerd"},
}

func main() {
	outDir := flag.String("out-dir", "", "directory to write NetworkPolicy YAML files (required)")
	flag.Parse()
	if *outDir == "" {
		fmt.Fprintln(os.Stderr, "error: --out-dir is required")
		os.Exit(1)
	}
	if err := run(*outDir); err != nil {
		fmt.Fprintf(os.Stderr, "export-mesh-np: %v\n", err)
		os.Exit(1)
	}
}

func run(outDir string) error {
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return fmt.Errorf("mkdir out-dir: %w", err)
	}
	for _, spec := range exports {
		np := security.NewAppGatewayMeshNetworkPolicy(spec.ns, spec.peerNS)
		data, err := yaml.Marshal(np)
		if err != nil {
			return fmt.Errorf("marshal %s: %w", spec.filename, err)
		}
		path := filepath.Join(outDir, spec.filename)
		if err := os.WriteFile(path, data, 0o644); err != nil {
			return fmt.Errorf("write %s: %w", path, err)
		}
	}
	return nil
}
