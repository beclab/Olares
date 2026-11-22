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
	if !strings.Contains(rb, "policyCompilationEnabled: \"false\"") {
		t.Fatal("policy compilation must default to false")
	}
}
