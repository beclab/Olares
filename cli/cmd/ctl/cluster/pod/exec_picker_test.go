package pod

import "testing"

func makePod(ns, name, phase string, containers ...string) Pod {
	var p Pod
	p.Metadata.Namespace = ns
	p.Metadata.Name = name
	p.Status.Phase = phase
	for _, c := range containers {
		p.Spec.Containers = append(p.Spec.Containers, PodContainer{Name: c})
	}
	return p
}

func TestPodsToEntries_FlattensContainers(t *testing.T) {
	pods := []Pod{
		makePod("ns-a", "web-1", "Running", "nginx", "sidecar"),
		makePod("ns-a", "db-1", "Pending", "postgres"),
	}
	got := podsToEntries(pods, "")
	if len(got) != 3 {
		t.Fatalf("want 3 entries (2+1 containers), got %d", len(got))
	}

	// Running flag tracks pod phase.
	for _, e := range got {
		wantRunning := e.Pod == "web-1"
		if e.Running != wantRunning {
			t.Errorf("%s/%s: Running=%v, want %v", e.Pod, e.Container, e.Running, wantRunning)
		}
	}
}

func TestPodsToEntries_NamespaceFallback(t *testing.T) {
	// Namespace-scoped list may leave metadata.namespace blank.
	pods := []Pod{makePod("", "web-1", "Running", "nginx")}
	got := podsToEntries(pods, "user-space-x")
	if len(got) != 1 {
		t.Fatalf("want 1 entry, got %d", len(got))
	}
	if got[0].Namespace != "user-space-x" {
		t.Fatalf("namespace fallback not applied: got %q", got[0].Namespace)
	}
}

func TestPodsToEntries_StatusFallbackDash(t *testing.T) {
	pods := []Pod{makePod("ns", "p", "", "c")} // no phase, no statuses
	got := podsToEntries(pods, "ns")
	if got[0].Status != "-" {
		t.Fatalf("empty status should render as '-', got %q", got[0].Status)
	}
	if got[0].Running {
		t.Fatalf("empty phase must not be Running")
	}
}

func TestPodsToEntries_SkipsContainerlessPods(t *testing.T) {
	pods := []Pod{makePod("ns", "p", "Running")} // no containers
	got := podsToEntries(pods, "ns")
	if len(got) != 0 {
		t.Fatalf("pod with no containers yields no entries, got %d", len(got))
	}
}
