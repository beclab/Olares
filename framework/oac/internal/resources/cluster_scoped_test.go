package resources

import (
	"strings"
	"testing"

	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"helm.sh/helm/v3/pkg/kube"
	cliresource "k8s.io/cli-runtime/pkg/resource"
)

func TestCheckClusterScopedFixedNames_NoClusterScoped(t *testing.T) {
	dep := newDeployment("web")
	dep.Namespace = AppNamespace
	if err := CheckClusterScopedFixedNames(kube.ResourceList{dep}, kube.ResourceList{dep}, ClusterScopedFixedNameOpts{}); err != nil {
		t.Fatalf("namespaced-only lists must pass: %v", err)
	}
}

func TestCheckClusterScopedFixedNames_DynamicName(t *testing.T) {
	listA := kube.ResourceList{newClusterRole(clusterScopedProbeA + "-role")}
	listB := kube.ResourceList{newClusterRole(clusterScopedProbeB + "-role")}
	if err := CheckClusterScopedFixedNames(listA, listB, ClusterScopedFixedNameOpts{AllowMultipleInstall: true}); err != nil {
		t.Fatalf("release-varying cluster-scoped names must pass: %v", err)
	}
}

func TestCheckClusterScopedFixedNames_FixedName(t *testing.T) {
	const fixed = "shared-cluster-role"
	listA := kube.ResourceList{newClusterRole(fixed)}
	listB := kube.ResourceList{newClusterRole(fixed)}
	err := CheckClusterScopedFixedNames(listA, listB, ClusterScopedFixedNameOpts{AllowMultipleInstall: true})
	if err == nil {
		t.Fatal("expected error for identical cluster-scoped name across probes")
	}
	if !strings.Contains(err.Error(), fixed) {
		t.Fatalf("error should mention %q, got: %v", fixed, err)
	}
	// The user-facing remediation hint replaces the older "probe release"
	// wording: the synthetic probe names (clusterScopedProbeA/B) are an
	// implementation detail of the two-render comparison, so the error now
	// points authors directly at the fix -- name resources after
	// {{ .Release.Namespace }}.
	if !strings.Contains(err.Error(), "{{ .Release.Namespace }}") {
		t.Fatalf("error should mention the {{ .Release.Namespace }} remediation hint, got: %v", err)
	}
}

func TestCheckClusterScopedFixedNames_Aggregates(t *testing.T) {
	listA := kube.ResourceList{
		newClusterRole("fixed-a"),
		newClusterRoleBinding("fixed-b"),
	}
	listB := kube.ResourceList{
		newClusterRole("fixed-a"),
		newClusterRoleBinding("fixed-b"),
	}
	err := CheckClusterScopedFixedNames(listA, listB, ClusterScopedFixedNameOpts{AllowMultipleInstall: true})
	if err == nil {
		t.Fatal("expected aggregated errors")
	}
	if !strings.Contains(err.Error(), "fixed-a") || !strings.Contains(err.Error(), "fixed-b") {
		t.Fatalf("expected both offenders in error, got: %v", err)
	}
}

func TestCheckClusterScopedFixedNames_UsernameScopedPassesWhenSingleInstall(t *testing.T) {
	const ownerA = clusterScopedUserProbeA + "-role"
	const ownerB = clusterScopedUserProbeB + "-role"
	listA := kube.ResourceList{newClusterRole(ownerA)}
	listB := kube.ResourceList{newClusterRole(ownerA)}
	listUserA := kube.ResourceList{newClusterRole(ownerA)}
	listUserB := kube.ResourceList{newClusterRole(ownerB)}
	err := CheckClusterScopedFixedNames(listA, listB, ClusterScopedFixedNameOpts{
		AllowMultipleInstall: false,
		UsernameProbeA:     listUserA,
		UsernameProbeB:     listUserB,
	})
	if err != nil {
		t.Fatalf("username-scoped cluster-scoped names must pass when allowMultipleInstall=false: %v", err)
	}
}

func TestCheckClusterScopedFixedNames_FixedNameStillFailsWhenSingleInstall(t *testing.T) {
	const fixed = "shared-cluster-role"
	listA := kube.ResourceList{newClusterRole(fixed)}
	listB := kube.ResourceList{newClusterRole(fixed)}
	listUserA := kube.ResourceList{newClusterRole(fixed)}
	listUserB := kube.ResourceList{newClusterRole(fixed)}
	err := CheckClusterScopedFixedNames(listA, listB, ClusterScopedFixedNameOpts{
		AllowMultipleInstall: false,
		UsernameProbeA:     listUserA,
		UsernameProbeB:     listUserB,
	})
	if err == nil {
		t.Fatal("expected error for truly fixed cluster-scoped name even when allowMultipleInstall=false")
	}
}

func newClusterRoleBinding(name string) *cliresource.Info {
	crb := &rbacv1.ClusterRoleBinding{
		TypeMeta:   metav1.TypeMeta{Kind: KindClusterRoleBinding, APIVersion: "rbac.authorization.k8s.io/v1"},
		ObjectMeta: metav1.ObjectMeta{Name: name},
	}
	return &cliresource.Info{Name: name, Object: crb}
}
