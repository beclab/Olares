package resources

import (
	"strings"
	"testing"

	"helm.sh/helm/v3/pkg/kube"
	appsv1 "k8s.io/api/apps/v1"
	cliresource "k8s.io/cli-runtime/pkg/resource"
)

func TestCollectWorkloadNames(t *testing.T) {
	list := kube.ResourceList{
		newDeployment("web"),
		newStatefulSet("db"),
		newDaemonSet("node-agent"), // must be ignored
	}
	names := CollectWorkloadNames(list)
	if _, ok := names["web"]; !ok {
		t.Fatalf("expected deployment 'web' to be collected")
	}
	if _, ok := names["db"]; !ok {
		t.Fatalf("expected statefulset 'db' to be collected")
	}
	if _, ok := names["node-agent"]; ok {
		t.Fatalf("daemonset must not be collected as a replica-managed workload")
	}
	if len(names) != 2 {
		t.Fatalf("expected exactly 2 workloads, got %d: %v", len(names), names)
	}
}

func TestCheckWorkloadReplicas_ExactMatch(t *testing.T) {
	list := kube.ResourceList{
		newDeployment("web"),
		newStatefulSet("db"),
	}
	if err := CheckWorkloadReplicas(list, map[string]int32{"web": 1, "db": 2}); err != nil {
		t.Fatalf("exact coverage must pass: %v", err)
	}
}

func TestCheckWorkloadReplicas_MissingEntry(t *testing.T) {
	list := kube.ResourceList{
		newDeployment("web"),
		newStatefulSet("db"),
	}
	err := CheckWorkloadReplicas(list, map[string]int32{"web": 1})
	if err == nil {
		t.Fatal("expected error when a workload has no workloadReplicas entry")
	}
	if !strings.Contains(err.Error(), `missing an entry for workload "db"`) {
		t.Fatalf("error should flag missing 'db', got: %v", err)
	}
}

func TestCheckWorkloadReplicas_UnknownEntry(t *testing.T) {
	list := kube.ResourceList{
		newDeployment("web"),
	}
	err := CheckWorkloadReplicas(list, map[string]int32{"web": 1, "ghost": 1})
	if err == nil {
		t.Fatal("expected error when workloadReplicas names a non-existent workload")
	}
	if !strings.Contains(err.Error(), `entry "ghost" does not match`) {
		t.Fatalf("error should flag unknown 'ghost', got: %v", err)
	}
}

// deploymentWithReplicas builds a Deployment whose spec.replicas is set, which
// the shared newDeployment helper deliberately leaves nil.
func deploymentWithReplicas(name string, replicas int32) *cliresource.Info {
	info := newDeployment(name)
	info.Object.(*appsv1.Deployment).Spec.Replicas = &replicas
	return info
}

func statefulSetWithReplicas(name string, replicas int32) *cliresource.Info {
	info := newStatefulSet(name)
	info.Object.(*appsv1.StatefulSet).Spec.Replicas = &replicas
	return info
}

func TestCheckWorkloadReplicasRendered_Follows(t *testing.T) {
	list := kube.ResourceList{
		deploymentWithReplicas("web", 0),
		statefulSetWithReplicas("db", 0),
		newDaemonSet("node-agent"), // no replicas to track
	}
	if err := CheckWorkloadReplicasRendered(list, 0); err != nil {
		t.Fatalf("workloads that track the injected value must pass: %v", err)
	}
}

func TestCheckWorkloadReplicasRendered_Pinned(t *testing.T) {
	// What `| default 1` produces: the injected 0 is discarded and the
	// workload renders at 1, so the app cannot be stopped.
	list := kube.ResourceList{deploymentWithReplicas("web", 1)}
	err := CheckWorkloadReplicasRendered(list, 0)
	if err == nil {
		t.Fatal("expected error when spec.replicas ignores the injected 0")
	}
	if !strings.Contains(err.Error(), "rendered as 1") {
		t.Fatalf("error should report the rendered value, got: %v", err)
	}
}

func TestCheckWorkloadReplicasRendered_Unset(t *testing.T) {
	// A template with no spec.replicas at all. Kubernetes defaults it to 1,
	// so it is a failure and not something to skip.
	list := kube.ResourceList{newDeployment("web")}
	err := CheckWorkloadReplicasRendered(list, 0)
	if err == nil {
		t.Fatal("expected error when spec.replicas is absent")
	}
	if !strings.Contains(err.Error(), "unset") {
		t.Fatalf("error should say the field is unset, got: %v", err)
	}
}

func TestCheckWorkloadReplicasRendered_StuckAtZero(t *testing.T) {
	// The opposite mistake: a chart hardcoded to 0 renders correctly for a
	// stop and can never be started.
	list := kube.ResourceList{deploymentWithReplicas("web", 0)}
	err := CheckWorkloadReplicasRendered(list, 1)
	if err == nil {
		t.Fatal("expected error when spec.replicas cannot be raised above 0")
	}
	if !strings.Contains(err.Error(), "rendered as 0") {
		t.Fatalf("error should report the rendered value, got: %v", err)
	}
}

func TestCheckReleaseNameWorkload_Deployment(t *testing.T) {
	probeA, probeB := "probe-a", "probe-b"
	listA := kube.ResourceList{newDeployment(probeA), newDeployment("static")}
	listB := kube.ResourceList{newDeployment(probeB), newDeployment("static")}
	if err := CheckReleaseNameWorkload(listA, listB, probeA, probeB); err != nil {
		t.Fatalf("a {{ .Release.Name }} Deployment must satisfy the check: %v", err)
	}
}

func TestCheckReleaseNameWorkload_StatefulSet(t *testing.T) {
	probeA, probeB := "probe-a", "probe-b"
	listA := kube.ResourceList{newStatefulSet(probeA)}
	listB := kube.ResourceList{newStatefulSet(probeB)}
	if err := CheckReleaseNameWorkload(listA, listB, probeA, probeB); err != nil {
		t.Fatalf("a {{ .Release.Name }} StatefulSet must satisfy the check: %v", err)
	}
}

func TestCheckReleaseNameWorkload_NoMatch(t *testing.T) {
	probeA, probeB := "probe-a", "probe-b"
	// Both lists contain only fixed-name workloads -- no metadata.name is
	// templated on the release name, so neither rendered list contains a
	// workload whose name equals the probe.
	listA := kube.ResourceList{newDeployment("fixed")}
	listB := kube.ResourceList{newDeployment("fixed")}
	err := CheckReleaseNameWorkload(listA, listB, probeA, probeB)
	if err == nil {
		t.Fatal("expected error when no Deployment/StatefulSet uses {{ .Release.Name }}")
	}
	if !strings.Contains(err.Error(), "{{ .Release.Name }}") {
		t.Fatalf("error should reference the {{ .Release.Name }} remediation, got: %v", err)
	}
}

func TestCheckReleaseNameWorkload_DaemonSetIgnored(t *testing.T) {
	probeA, probeB := "probe-a", "probe-b"
	// A DaemonSet whose name happens to render to the release name does
	// not satisfy the rule -- the check is about Deployment/StatefulSet
	// primary workloads only, which mirrors CollectWorkloadNames.
	listA := kube.ResourceList{newDaemonSet(probeA)}
	listB := kube.ResourceList{newDaemonSet(probeB)}
	if err := CheckReleaseNameWorkload(listA, listB, probeA, probeB); err == nil {
		t.Fatal("DaemonSet alone must not satisfy the release-name workload rule")
	}
}

func TestCheckReleaseNameWorkload_PartialMatch(t *testing.T) {
	probeA, probeB := "probe-a", "probe-b"
	// listA carries a release-named Deployment but listB does not -- this
	// only happens when a chart conditionally emits a release-named
	// workload. The check must fail closed (require both) because we
	// cannot prove a single template uses {{ .Release.Name }} otherwise.
	listA := kube.ResourceList{newDeployment(probeA)}
	listB := kube.ResourceList{newDeployment("fixed")}
	if err := CheckReleaseNameWorkload(listA, listB, probeA, probeB); err == nil {
		t.Fatal("expected error when only one render contains a release-named workload")
	}
}

func TestCheckOverlayGatewayWorkloads(t *testing.T) {
	list := kube.ResourceList{
		newDeployment("jellyfin"),
		newStatefulSet("db"),
	}
	if err := CheckOverlayGatewayWorkloads(list, []string{"jellyfin"}); err != nil {
		t.Fatalf("existing workload reference must pass: %v", err)
	}

	err := CheckOverlayGatewayWorkloads(list, []string{"jellyfin", "missing"})
	if err == nil {
		t.Fatal("expected error for overlayGateway workload that does not exist")
	}
	if !strings.Contains(err.Error(), `workload "missing"`) {
		t.Fatalf("error should flag missing workload, got: %v", err)
	}

	// An empty workload reference is skipped (the manifest validator already
	// enforces required-ness; this avoids a duplicate, less specific error).
	if err := CheckOverlayGatewayWorkloads(list, []string{""}); err != nil {
		t.Fatalf("empty workload reference must be skipped here: %v", err)
	}
}
