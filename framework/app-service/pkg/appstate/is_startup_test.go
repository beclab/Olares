package appstate

import (
	"testing"

	"github.com/beclab/Olares/framework/app-service/pkg/testutil"
)

func TestIsStartUpReadyPod(t *testing.T) {
	dep := testutil.NewDeployment("nginx", "nginx-alice", 1)
	pod := testutil.NewReadyPod("nginx-abc", "nginx-alice", map[string]string{"app": "nginx"})
	am := testutil.NewAppManager("nginx", testutil.WithNamespace("nginx-alice"))
	c := testutil.NewFakeClient(dep, pod, am)

	ok, err := isStartUp(am, c)
	if err != nil {
		t.Fatalf("isStartUp: %v", err)
	}
	if !ok {
		t.Error("expected started up = true for a ready pod")
	}
}

func TestIsStartUpNoPods(t *testing.T) {
	dep := testutil.NewDeployment("nginx", "nginx-alice", 1)
	am := testutil.NewAppManager("nginx", testutil.WithNamespace("nginx-alice"))
	c := testutil.NewFakeClient(dep, am)

	ok, err := isStartUp(am, c)
	if err == nil {
		t.Error("expected error when no pods are found")
	}
	if ok {
		t.Error("expected started up = false when no pods are found")
	}
}

func TestIsStartUpNotStarted(t *testing.T) {
	dep := testutil.NewDeployment("nginx", "nginx-alice", 1)
	pod := testutil.NewReadyPod("nginx-abc", "nginx-alice", map[string]string{"app": "nginx"})
	notStarted := false
	pod.Status.ContainerStatuses[0].Started = &notStarted
	pod.Status.ContainerStatuses[0].Ready = false
	am := testutil.NewAppManager("nginx", testutil.WithNamespace("nginx-alice"))
	c := testutil.NewFakeClient(dep, pod, am)

	ok, err := isStartUp(am, c)
	if err != nil {
		t.Fatalf("isStartUp: %v", err)
	}
	if ok {
		t.Error("expected started up = false for a not-started container")
	}
}
