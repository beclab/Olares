package access

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestAccessCRDsAndRBACPresent(t *testing.T) {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	deploy := filepath.Clean(filepath.Join(filepath.Dir(file), "../../.olares/config/cluster/deploy"))
	crds, err := os.ReadFile(filepath.Join(deploy, "access-crds.yaml"))
	if err != nil {
		t.Fatalf("access-crds.yaml: %v", err)
	}
	body := string(crds)
	for _, name := range []string{
		"appcapabilitysets.access.olares.io",
		"appaccessrequests.access.olares.io",
		"appaccesspolicies.access.olares.io",
		"AppCapabilitySet",
		"AppAccessRequest",
		"AppAccessPolicy",
	} {
		if !strings.Contains(body, name) {
			t.Fatalf("missing %s in access-crds.yaml", name)
		}
	}
	rbac, err := os.ReadFile(filepath.Join(deploy, "access-rbac.yaml"))
	if err != nil {
		t.Fatalf("access-rbac.yaml: %v", err)
	}
	rb := string(rbac)
	if !strings.Contains(rb, "name: os-app-access") {
		t.Fatal("missing os-app-access namespace")
	}
	if strings.Contains(rb, "endState:") {
		t.Fatal("app-access-runtime must not expose endState kill switch")
	}
	if !strings.Contains(rb, `httpMeshPolicy: "cluster-authenticated"`) {
		t.Fatal("runtime CM must project cluster-authenticated")
	}
}

func TestAccessAdmissionPoliciesRemainFailClosed(t *testing.T) {
	_, file, _, _ := runtime.Caller(0)
	deploy := filepath.Join(filepath.Dir(file), "../../.olares/config/cluster/deploy")
	if _, err := os.Stat(filepath.Join(deploy, "access-admission.yaml")); !os.IsNotExist(err) {
		t.Fatalf("standalone admission workload is still shipped: %v", err)
	}
	raw, err := os.ReadFile(filepath.Join(deploy, "access-admission-policies.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	body := string(raw)
	for _, required := range []string{"olares-access-posture", "olares-access-grant-lifecycle", "olares-access-projection-writer", "authentication.kubernetes.io/pod-uid", "parameterNotFoundAction: Deny", "validationActions: [Deny, Audit]", "failurePolicy: Fail", "helm.sh/resource-policy: keep"} {
		if !strings.Contains(body, required) {
			t.Errorf("missing access guard %s", required)
		}
	}
	if strings.Contains(body, "failurePolicy: Ignore") {
		t.Fatal("access guards must fail closed even on bootstrap")
	}
}

func TestHTTPMeshWaveConfigNoKillSwitch(t *testing.T) {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	deploy := filepath.Clean(filepath.Join(filepath.Dir(file), "../../.olares/config/cluster/deploy"))
	raw, err := os.ReadFile(filepath.Join(deploy, "http-mesh-wave-config.yaml"))
	if err != nil {
		t.Fatalf("http-mesh-wave-config.yaml: %v", err)
	}
	body := string(raw)
	if strings.Contains(body, "endState:") {
		t.Fatal("wave CM must not expose endState")
	}
	if !strings.Contains(body, `httpMeshWave: "4"`) {
		t.Fatal("wave must project 4")
	}
	if !strings.Contains(body, `httpMeshPolicy: "cluster-authenticated"`) {
		t.Fatal("must project cluster-authenticated")
	}
}

func TestDeployOmitsTrafficGroupDefaultDenyNP(t *testing.T) {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	deploy := filepath.Clean(filepath.Join(filepath.Dir(file), "../../.olares/config/cluster/deploy"))
	entries, err := os.ReadDir(deploy)
	if err != nil {
		t.Fatalf("read deploy: %v", err)
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".yaml") {
			continue
		}
		raw, err := os.ReadFile(filepath.Join(deploy, e.Name()))
		if err != nil {
			t.Fatalf("read %s: %v", e.Name(), err)
		}
		body := string(raw)
		if strings.Contains(body, "olares-traffic-group-default-deny") {
			t.Fatalf("%s must not ship traffic-group default-deny NetworkPolicy (AC-5)", e.Name())
		}
		if e.Name() == "traffic-group-labels.yaml" && strings.Contains(body, "kind: NetworkPolicy") {
			t.Fatal("traffic-group-labels.yaml must be labels-only; NetworkPolicy is opt-in outside deploy/")
		}
	}
}
