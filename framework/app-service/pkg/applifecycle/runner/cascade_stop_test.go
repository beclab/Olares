package runner

import (
	"context"
	"testing"
	"time"

	appv1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/beclab/Olares/framework/app-service/pkg/apiserver/api"
	"github.com/beclab/Olares/framework/app-service/pkg/appcfg"
	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/adapters/fake"
	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/domain"
	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/ports"
)

func TestCascadeStopAllPeers_PersistsStopWithoutWritingStatus(t *testing.T) {
	store := fake.NewAMStore()
	store.Add(&appv1.ApplicationManager{
		ObjectMeta: metav1.ObjectMeta{
			Name: "parent",
			Annotations: map[string]string{
				api.AppStopAllKey: "true",
			},
		},
		Spec: appv1.ApplicationManagerSpec{AppName: "app", AppOwner: "alice", Type: appv1.App},
		Status: appv1.ApplicationManagerStatus{
			State:   appv1.Stopping,
			Reason:  "Evicted",
			Message: "evicted pod",
		},
	})
	store.Add(&appv1.ApplicationManager{
		ObjectMeta: metav1.ObjectMeta{Name: "peer"},
		Spec:       appv1.ApplicationManagerSpec{AppName: "app", AppOwner: "bob", Type: appv1.App, OpType: appv1.InstallOp},
		Status:     appv1.ApplicationManagerStatus{State: appv1.Running, OpID: "old-op"},
	})

	pending := NewPendingActionQueue()
	requeuer := &RecordingRequeuer{}
	r, err := New(Config{
		Engine:      domain.NewEngine(domain.DefaultTable),
		AMStore:     store,
		WorkloadOps: &fake.WorkloadOps{},
		Validator:   &fake.InstallValidator{},
		ImagePuller: fake.NewImagePuller(),
		Recorder:    &fake.EventRecorder{},
		Clock:       fake.NewClock(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)),
		Pending:     pending,
		Requeuer:    requeuer,
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(r.Stop)

	r.cascadeStopAllPeers(context.Background(), ports.WorkloadRequest{
		AM: &appv1.ApplicationManager{
			ObjectMeta: metav1.ObjectMeta{
				Name:        "parent",
				Annotations: map[string]string{api.AppStopAllKey: "true"},
			},
			Spec:   appv1.ApplicationManagerSpec{AppName: "app", AppOwner: "alice"},
			Status: appv1.ApplicationManagerStatus{Reason: "Evicted", Message: "evicted pod"},
		},
		AppCfg: &appcfg.ApplicationConfig{APIVersion: appcfg.V2},
	})

	got, err := store.Get(context.Background(), "peer")
	if err != nil {
		t.Fatal(err)
	}
	if got.Status.State != appv1.Running || got.Status.OpID != "old-op" {
		t.Fatalf("peer status was written: state=%s opID=%s", got.Status.State, got.Status.OpID)
	}
	if got.Spec.OpType != appv1.StopOp {
		t.Fatalf("Spec.OpType = %q, want StopOp", got.Spec.OpType)
	}
	if got.Annotations[api.AppStopReasonKey] != "Evicted" || got.Annotations[api.AppStopMessageKey] != "evicted pod" {
		t.Fatalf("stop annotations = %v", got.Annotations)
	}
	if got.Annotations[api.AppStopAllKey] != "" {
		t.Fatalf("peer stop-all = %q, want empty", got.Annotations[api.AppStopAllKey])
	}
	if got.Annotations[domain.AnnotationOperationRequestID] == "" || got.Annotations[domain.AnnotationOperationRequestID] == "old-op" {
		t.Fatalf("operation request id = %q", got.Annotations[domain.AnnotationOperationRequestID])
	}
	if action := pending.Pop("peer"); action != nil {
		t.Fatalf("queued action = %#v, want none", action)
	}
	if calls := requeuer.Calls(); len(calls) != 0 {
		t.Fatalf("requeue = %v, want none", calls)
	}
}

func TestCascadeStopAllPeers_PrefersStopAnnotationsOverStaleStatus(t *testing.T) {
	store := fake.NewAMStore()
	store.Add(&appv1.ApplicationManager{
		ObjectMeta: metav1.ObjectMeta{Name: "peer"},
		Spec:       appv1.ApplicationManagerSpec{AppName: "app", AppOwner: "bob", Type: appv1.App},
		Status:     appv1.ApplicationManagerStatus{State: appv1.Running},
	})

	r, err := New(Config{
		Engine:      domain.NewEngine(domain.DefaultTable),
		AMStore:     store,
		WorkloadOps: &fake.WorkloadOps{},
		Validator:   &fake.InstallValidator{},
		ImagePuller: fake.NewImagePuller(),
		Recorder:    &fake.EventRecorder{},
		Clock:       fake.NewClock(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)),
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(r.Stop)

	r.cascadeStopAllPeers(context.Background(), ports.WorkloadRequest{
		AM: &appv1.ApplicationManager{
			ObjectMeta: metav1.ObjectMeta{
				Name: "parent",
				Annotations: map[string]string{
					api.AppStopAllKey:     "true",
					api.AppStopReasonKey:  "Evicted",
					api.AppStopMessageKey: "evicted pod",
				},
			},
			Spec: appv1.ApplicationManagerSpec{AppName: "app", AppOwner: "alice"},
			Status: appv1.ApplicationManagerStatus{
				State:   appv1.Running,
				Reason:  string(appv1.Running),
				Message: "application is running",
			},
		},
		AppCfg: &appcfg.ApplicationConfig{APIVersion: appcfg.V2},
	})

	got, err := store.Get(context.Background(), "peer")
	if err != nil {
		t.Fatal(err)
	}
	if got.Annotations[api.AppStopReasonKey] != "Evicted" || got.Annotations[api.AppStopMessageKey] != "evicted pod" {
		t.Fatalf("stop annotations = %v", got.Annotations)
	}
}

func TestCascadeStopAllPeers_SkipsPeerWithUnconsumedStop(t *testing.T) {
	store := fake.NewAMStore()
	store.Add(&appv1.ApplicationManager{
		ObjectMeta: metav1.ObjectMeta{
			Name: "peer",
			Annotations: map[string]string{
				domain.AnnotationOperationRequestID: "pending-stop",
			},
		},
		Spec:   appv1.ApplicationManagerSpec{AppName: "app", AppOwner: "bob", Type: appv1.App, OpType: appv1.StopOp},
		Status: appv1.ApplicationManagerStatus{State: appv1.Running, OpID: "old-op"},
	})

	r, err := New(Config{
		Engine:      domain.NewEngine(domain.DefaultTable),
		AMStore:     store,
		WorkloadOps: &fake.WorkloadOps{},
		Validator:   &fake.InstallValidator{},
		ImagePuller: fake.NewImagePuller(),
		Recorder:    &fake.EventRecorder{},
		Clock:       fake.NewClock(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)),
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(r.Stop)

	r.cascadeStopAllPeers(context.Background(), ports.WorkloadRequest{
		AM: &appv1.ApplicationManager{
			ObjectMeta: metav1.ObjectMeta{
				Name:        "parent",
				Annotations: map[string]string{api.AppStopAllKey: "true"},
			},
			Spec:   appv1.ApplicationManagerSpec{AppName: "app", AppOwner: "alice"},
			Status: appv1.ApplicationManagerStatus{State: appv1.Stopping, Reason: "Evicted"},
		},
		AppCfg: &appcfg.ApplicationConfig{APIVersion: appcfg.V2},
	})

	got, err := store.Get(context.Background(), "peer")
	if err != nil {
		t.Fatal(err)
	}
	if got.Annotations[domain.AnnotationOperationRequestID] != "pending-stop" {
		t.Fatalf("request id = %q, want pending-stop", got.Annotations[domain.AnnotationOperationRequestID])
	}
	if got.Status.State != appv1.Running {
		t.Fatalf("peer state = %s, want Running", got.Status.State)
	}
}

func TestCascadeStopAllPeers_UnconsumedRequest(t *testing.T) {
	cases := []struct {
		op   appv1.OpType
		stop bool
	}{
		{op: appv1.UpgradeOp, stop: false},
		{op: appv1.UninstallOp, stop: false},
		{op: appv1.ApplyEnvOp, stop: true},
	}
	for _, tc := range cases {
		t.Run(string(tc.op), func(t *testing.T) {
			store := fake.NewAMStore()
			store.Add(&appv1.ApplicationManager{
				ObjectMeta: metav1.ObjectMeta{
					Name: "peer",
					Annotations: map[string]string{
						domain.AnnotationOperationRequestID: "pending-other",
					},
				},
				Spec:   appv1.ApplicationManagerSpec{AppName: "app", AppOwner: "bob", Type: appv1.App, OpType: tc.op},
				Status: appv1.ApplicationManagerStatus{State: appv1.Running, OpID: "old-op"},
			})
			r, err := New(Config{
				Engine:      domain.NewEngine(domain.DefaultTable),
				AMStore:     store,
				WorkloadOps: &fake.WorkloadOps{},
				Validator:   &fake.InstallValidator{},
				ImagePuller: fake.NewImagePuller(),
				Recorder:    &fake.EventRecorder{},
				Clock:       fake.NewClock(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)),
			})
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(r.Stop)

			r.cascadeStopAllPeers(context.Background(), ports.WorkloadRequest{
				AM: &appv1.ApplicationManager{
					ObjectMeta: metav1.ObjectMeta{
						Name:        "parent",
						Annotations: map[string]string{api.AppStopAllKey: "true"},
					},
					Spec:   appv1.ApplicationManagerSpec{AppName: "app", AppOwner: "alice"},
					Status: appv1.ApplicationManagerStatus{State: appv1.Stopping},
				},
				AppCfg: &appcfg.ApplicationConfig{APIVersion: appcfg.V2},
			})

			got, err := store.Get(context.Background(), "peer")
			if err != nil {
				t.Fatal(err)
			}
			if tc.stop {
				if got.Spec.OpType != appv1.StopOp {
					t.Fatalf("Spec.OpType = %q, want StopOp", got.Spec.OpType)
				}
				return
			}
			if got.Spec.OpType != tc.op {
				t.Fatalf("Spec.OpType = %q, want %q", got.Spec.OpType, tc.op)
			}
			if got.Annotations[domain.AnnotationOperationRequestID] != "pending-other" {
				t.Fatalf("request id = %q, want pending-other", got.Annotations[domain.AnnotationOperationRequestID])
			}
		})
	}
}
