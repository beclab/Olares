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
	if err := CheckClusterScopedFixedNames(kube.ResourceList{dep}, kube.ResourceList{dep}); err != nil {
		t.Fatalf("namespaced-only lists must pass: %v", err)
	}
}

func TestCheckClusterScopedFixedNames_DynamicName(t *testing.T) {
	listA := kube.ResourceList{newClusterRole(clusterScopedProbeA + "-role")}
	listB := kube.ResourceList{newClusterRole(clusterScopedProbeB + "-role")}
	if err := CheckClusterScopedFixedNames(listA, listB); err != nil {
		t.Fatalf("release-varying cluster-scoped names must pass: %v", err)
	}
}

func TestCheckClusterScopedFixedNames_FixedName(t *testing.T) {
	const fixed = "shared-cluster-role"
	listA := kube.ResourceList{newClusterRole(fixed)}
	listB := kube.ResourceList{newClusterRole(fixed)}
	err := CheckClusterScopedFixedNames(listA, listB)
	if err == nil {
		t.Fatal("expected error for identical cluster-scoped name across probes")
	}
	if !strings.Contains(err.Error(), fixed) {
		t.Fatalf("error should mention %q, got: %v", fixed, err)
	}
	if !strings.Contains(err.Error(), clusterScopedProbeA) {
		t.Fatalf("error should mention probe release %q, got: %v", clusterScopedProbeA, err)
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
	err := CheckClusterScopedFixedNames(listA, listB)
	if err == nil {
		t.Fatal("expected aggregated errors")
	}
	if !strings.Contains(err.Error(), "fixed-a") || !strings.Contains(err.Error(), "fixed-b") {
		t.Fatalf("expected both offenders in error, got: %v", err)
	}
}

func newClusterRoleBinding(name string) *cliresource.Info {
	crb := &rbacv1.ClusterRoleBinding{
		TypeMeta:   metav1.TypeMeta{Kind: KindClusterRoleBinding, APIVersion: "rbac.authorization.k8s.io/v1"},
		ObjectMeta: metav1.ObjectMeta{Name: name},
	}
	return &cliresource.Info{Name: name, Object: crb}
}
