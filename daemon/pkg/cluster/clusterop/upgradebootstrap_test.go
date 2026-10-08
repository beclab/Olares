package clusterop

import (
	"context"
	"net/http"
	"testing"

	"github.com/beclab/Olares/daemon/pkg/cluster/inventory"
	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	ktesting "k8s.io/client-go/testing"
)

func TestLegacyWorkerBootstrapJobIsNodeBoundAndResumable(t *testing.T) {
	client := fake.NewSimpleClientset()
	client.PrependReactor("create", "jobs", func(action ktesting.Action) (bool, runtime.Object, error) {
		job := action.(ktesting.CreateAction).GetObject().(*batchv1.Job)
		job.Status.Succeeded = 1
		return false, nil, nil
	})
	node := inventory.Node{NodeName: "worker-1", Role: inventory.RoleWorker}
	where := UpgradeLocations{WizardURL: "https://mirror.example/wizard.tar.gz"}
	if err := bootstrapLegacyUpgradeWorker(context.Background(), client, node, "op-1", "1.12.8", where); err != nil {
		t.Fatal(err)
	}
	jobs, err := client.BatchV1().Jobs(upgradeSecretNamespace).List(context.Background(), metav1.ListOptions{})
	if err != nil || len(jobs.Items) != 1 {
		t.Fatalf("jobs = %d, error = %v", len(jobs.Items), err)
	}
	job := &jobs.Items[0]
	if job.Spec.Template.Spec.NodeName != node.NodeName || !bootstrapJobVersionIs(job, "1.12.8") {
		t.Fatalf("bootstrap job does not bind node and version: %+v", job.Spec.Template.Spec)
	}
	container := job.Spec.Template.Spec.Containers[0]
	if container.SecurityContext == nil || container.SecurityContext.Privileged == nil || !*container.SecurityContext.Privileged {
		t.Fatal("bootstrap cannot enter the host filesystem")
	}
	if err := bootstrapLegacyUpgradeWorker(context.Background(), client, node, "op-1", "1.12.8", where); err != nil {
		t.Fatalf("resuming existing job: %v", err)
	}
}

func TestLegacyWorkerBootstrapOnlyOnMissingReadinessRoute(t *testing.T) {
	for _, tc := range []struct {
		name          string
		status        int
		wantBootstrap bool
	}{
		{"legacy 404", http.StatusNotFound, true},
		{"auth refusal", http.StatusForbidden, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			h := newUpgradeHarness(t, samplePlan(), twoNodeCluster())
			h.deps.Store = newTestStore(t)
			bootstrapped := false
			h.deps.Upgrade.Readiness = func(_ context.Context, _ inventory.Node, _, _ string) (UpgradeReadiness, error) {
				if bootstrapped {
					return readyOn("1.12.8"), nil
				}
				return UpgradeReadiness{}, &UpgradeHTTPStatusError{StatusCode: tc.status, Status: http.StatusText(tc.status)}
			}
			h.deps.Upgrade.BootstrapLegacy = func(_ context.Context, node inventory.Node, _, version string, _ UpgradeLocations) error {
				if node.NodeName != "worker-1" || version != "1.12.8" {
					t.Fatalf("wrong bootstrap target: %s %s", node.NodeName, version)
				}
				bootstrapped = true
				return nil
			}
			m, err := NewManager(h.deps)
			if err != nil {
				t.Fatal(err)
			}
			op, err := m.Create(context.Background(), CreateRequest{
				Type: TypeUpgrade, RequestID: "olares-upgrade-1.12.8-test",
				Scope: ScopeCluster, ClusterID: "cluster-1", Owner: "alice@olares.com",
				UpgradeVersion: "1.12.8",
			})
			if err != nil {
				t.Fatal(err)
			}
			got := waitForTerminal(t, m, op.ID)
			if bootstrapped != tc.wantBootstrap {
				t.Fatalf("bootstrap = %t, want %t", bootstrapped, tc.wantBootstrap)
			}
			if tc.wantBootstrap && got.Status != StatusSucceeded {
				t.Fatalf("upgrade = %s/%s", got.Status, got.Code)
			}
		})
	}
}
