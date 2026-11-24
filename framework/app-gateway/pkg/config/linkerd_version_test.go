package config

import (
	"bytes"
	"gopkg.in/yaml.v3"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// pinnedLinkerdEdge is the control-plane/proxy tag shipped in cluster deploy manifests.
const pinnedLinkerdEdge = "edge-26.9.1"

func TestLinkerdControlPlanePinsEdge2691(t *testing.T) {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "../.."))
	manifest := filepath.Join(root, ".olares/config/cluster/deploy/linkerd-control-plane.yaml")
	b, err := os.ReadFile(manifest)
	if err != nil {
		t.Fatalf("read linkerd-control-plane.yaml: %v", err)
	}
	body := string(b)
	if strings.Contains(body, "edge-26.5.1") {
		t.Fatal("linkerd-control-plane.yaml still references vulnerable edge-26.5.1")
	}
	if !strings.Contains(body, "linkerdVersion: "+pinnedLinkerdEdge) {
		t.Fatalf("expected linkerdVersion %s in control-plane manifest", pinnedLinkerdEdge)
	}
	if !strings.Contains(body, "cr.l5d.io/linkerd/proxy:"+pinnedLinkerdEdge) {
		t.Fatalf("expected proxy image tag %s", pinnedLinkerdEdge)
	}
}

func TestLinkerdNativeControlPlaneBootstrap(t *testing.T) {
	_, file, _, _ := runtime.Caller(0)
	data, err := os.ReadFile(filepath.Join(filepath.Dir(file), "../../.olares/config/cluster/deploy/linkerd-control-plane.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	type container struct {
		Name          string
		RestartPolicy string         `yaml:"restartPolicy"`
		StartupProbe  map[string]any `yaml:"startupProbe"`
	}
	want := map[string]bool{"linkerd-identity": true, "linkerd-destination": true, "linkerd-proxy-injector": true}
	for {
		var resource struct {
			Kind     string
			Metadata struct{ Name string }
			Data     map[string]string
			Spec     struct {
				Template struct {
					Spec struct {
						Containers     []container
						InitContainers []container `yaml:"initContainers"`
					}
				}
			}
		}
		err := decoder.Decode(&resource)
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		if resource.Kind == "ConfigMap" && resource.Metadata.Name == "linkerd-config" {
			var values struct {
				Proxy struct {
					NativeSidecar bool `yaml:"nativeSidecar"`
				}
			}
			if err := yaml.Unmarshal([]byte(resource.Data["values"]), &values); err != nil || !values.Proxy.NativeSidecar {
				t.Fatalf("injector must default to native sidecars, got %v (%v)", values.Proxy, err)
			}
		}
		if resource.Kind != "Deployment" || !want[resource.Metadata.Name] {
			continue
		}
		pod := resource.Spec.Template.Spec
		if len(pod.InitContainers) != 2 || pod.InitContainers[0].Name != "linkerd-init" ||
			pod.InitContainers[1].Name != "linkerd-proxy" || pod.InitContainers[1].RestartPolicy != "Always" ||
			pod.InitContainers[1].StartupProbe != nil {
			t.Fatalf("%s must start its native proxy after interception, without waiting for identity readiness", resource.Metadata.Name)
		}
		for _, c := range pod.Containers {
			if c.Name == "linkerd-proxy" {
				t.Fatalf("%s still contains a traditional proxy", resource.Metadata.Name)
			}
		}
		delete(want, resource.Metadata.Name)
	}
	if len(want) != 0 {
		t.Fatalf("control-plane Deployments not found: %v", want)
	}
}
