package imgpull_test

import (
	"context"
	"errors"
	"sync"
	"testing"

	appv1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/beclab/Olares/framework/app-service/pkg/appcfg"
	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/adapters/imgpull"
	"github.com/beclab/Olares/framework/app-service/pkg/images"
	"github.com/beclab/Olares/framework/app-service/pkg/testutil"
)

// fakeIM is a mock images.ImageManager that records every call.
// Keeps tests independent of testutil.NewFakeClient's Status
// subresource wiring (which does not register ImageManager).
type fakeIM struct {
	mu          sync.Mutex
	createCalls []fakeIMCreate
	updateCalls []fakeIMUpdate
	createErr   error
	updateErr   error
	pollCalls   int
	pollErr     error
}

type fakeIMCreate struct {
	Name string
	Refs []appv1.Ref
}

type fakeIMUpdate struct {
	Name    string
	State   string
	Message string
}

func (f *fakeIM) Create(_ context.Context, am *appv1.ApplicationManager, refs []appv1.Ref) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.createCalls = append(f.createCalls, fakeIMCreate{Name: am.Name, Refs: refs})
	return f.createErr
}

func (f *fakeIM) UpdateStatus(_ context.Context, name, state, message string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.updateCalls = append(f.updateCalls, fakeIMUpdate{Name: name, State: state, Message: message})
	return f.updateErr
}

func (f *fakeIM) PollDownloadProgress(_ context.Context, _ *appv1.ApplicationManager) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.pollCalls++
	return f.pollErr
}

// Compile-time guard so the fake stays in lockstep with the
// legacy interface — the whole point of composition is that a
// signature drift breaks HERE, not in the adapter's port method
// signatures.
var _ images.ImageManager = (*fakeIM)(nil)

// -----------------------------------------------------------------
// New
// -----------------------------------------------------------------

func TestNew_NilClientPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("New(nil, ...): expected panic")
		}
	}()
	_ = imgpull.New(nil, func(context.Context, *appv1.ApplicationManager, *appcfg.ApplicationConfig) ([]appv1.Ref, error) {
		return nil, nil
	})
}

func TestNew_NilResolverPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("New(client, nil): expected panic")
		}
	}()
	_ = imgpull.New(testutil.NewFakeClient(), nil)
}

// -----------------------------------------------------------------
// ResolveRefs
// -----------------------------------------------------------------

func TestResolveRefs_NilAM(t *testing.T) {
	p := imgpull.New(testutil.NewFakeClient(),
		func(context.Context, *appv1.ApplicationManager, *appcfg.ApplicationConfig) ([]appv1.Ref, error) {
			t.Fatal("resolver should not be called on nil AM")
			return nil, nil
		})
	_, err := p.ResolveRefs(context.Background(), nil, &appcfg.ApplicationConfig{})
	if err == nil {
		t.Fatal("expected error for nil AM")
	}
}

func TestResolveRefs_WrongAppCfgType(t *testing.T) {
	p := imgpull.New(testutil.NewFakeClient(),
		func(context.Context, *appv1.ApplicationManager, *appcfg.ApplicationConfig) ([]appv1.Ref, error) {
			t.Fatal("resolver should not be called on wrong appCfg type")
			return nil, nil
		})
	am := &appv1.ApplicationManager{ObjectMeta: metav1.ObjectMeta{Name: "am"}}
	_, err := p.ResolveRefs(context.Background(), am, "not a config")
	if err == nil {
		t.Fatal("expected error for wrong appCfg type")
	}
}

func TestResolveRefs_NilAppCfg(t *testing.T) {
	p := imgpull.New(testutil.NewFakeClient(),
		func(context.Context, *appv1.ApplicationManager, *appcfg.ApplicationConfig) ([]appv1.Ref, error) {
			t.Fatal("resolver should not be called on nil appCfg")
			return nil, nil
		})
	am := &appv1.ApplicationManager{ObjectMeta: metav1.ObjectMeta{Name: "am"}}
	// nil pointer of the CORRECT type is still rejected — same as
	// a wrong type. This is the strict variant documented on the
	// port method.
	var nilCfg *appcfg.ApplicationConfig
	_, err := p.ResolveRefs(context.Background(), am, nilCfg)
	if err == nil {
		t.Fatal("expected error for nil *appcfg.ApplicationConfig")
	}
}

func TestResolveRefs_HappyPath(t *testing.T) {
	called := false
	wantRefs := []appv1.Ref{{Name: "nginx:1.25"}, {Name: "redis:7"}}
	p := imgpull.New(testutil.NewFakeClient(),
		func(_ context.Context, am *appv1.ApplicationManager, cfg *appcfg.ApplicationConfig) ([]appv1.Ref, error) {
			called = true
			if am.Name != "am" {
				t.Errorf("resolver: AM name %q want am", am.Name)
			}
			if cfg == nil || cfg.AppName != "app-x" {
				t.Errorf("resolver: cfg AppName wrong: %+v", cfg)
			}
			return wantRefs, nil
		})
	am := &appv1.ApplicationManager{ObjectMeta: metav1.ObjectMeta{Name: "am"}}
	cfg := &appcfg.ApplicationConfig{AppName: "app-x"}

	got, err := p.ResolveRefs(context.Background(), am, cfg)
	if err != nil {
		t.Fatalf("ResolveRefs: %v", err)
	}
	if !called {
		t.Fatal("resolver was not called")
	}
	if len(got) != 2 || got[0].Name != "nginx:1.25" {
		t.Errorf("refs: %+v", got)
	}
}

func TestResolveRefs_ResolverError(t *testing.T) {
	wantErr := errors.New("boom")
	p := imgpull.New(testutil.NewFakeClient(),
		func(context.Context, *appv1.ApplicationManager, *appcfg.ApplicationConfig) ([]appv1.Ref, error) {
			return nil, wantErr
		})
	am := &appv1.ApplicationManager{ObjectMeta: metav1.ObjectMeta{Name: "am"}}
	cfg := &appcfg.ApplicationConfig{}
	_, err := p.ResolveRefs(context.Background(), am, cfg)
	if !errors.Is(err, wantErr) {
		t.Errorf("err: got %v want %v", err, wantErr)
	}
}

// -----------------------------------------------------------------
// CreateJob (delegation via fakeIM)
// -----------------------------------------------------------------

func TestCreateJob_NilAM(t *testing.T) {
	fim := &fakeIM{}
	p := imgpull.New(testutil.NewFakeClient(),
		func(context.Context, *appv1.ApplicationManager, *appcfg.ApplicationConfig) ([]appv1.Ref, error) {
			return nil, nil
		},
		imgpull.WithImageManager(fim),
	)
	err := p.CreateJob(context.Background(), nil, nil)
	if err == nil {
		t.Fatal("expected error for nil AM")
	}
	if len(fim.createCalls) != 0 {
		t.Errorf("nil-AM CreateJob leaked to underlying IM: %+v", fim.createCalls)
	}
}

func TestCreateJob_DelegatesToImageManager(t *testing.T) {
	fim := &fakeIM{}
	p := imgpull.New(testutil.NewFakeClient(),
		func(context.Context, *appv1.ApplicationManager, *appcfg.ApplicationConfig) ([]appv1.Ref, error) {
			return nil, nil
		},
		imgpull.WithImageManager(fim),
	)
	am := &appv1.ApplicationManager{ObjectMeta: metav1.ObjectMeta{Name: "am"}}
	refs := []appv1.Ref{{Name: "nginx:1.25"}}
	if err := p.CreateJob(context.Background(), am, refs); err != nil {
		t.Fatalf("CreateJob: %v", err)
	}
	if len(fim.createCalls) != 1 {
		t.Fatalf("createCalls: %d want 1", len(fim.createCalls))
	}
	if fim.createCalls[0].Name != "am" {
		t.Errorf("call name: %q want am", fim.createCalls[0].Name)
	}
	if len(fim.createCalls[0].Refs) != 1 || fim.createCalls[0].Refs[0].Name != "nginx:1.25" {
		t.Errorf("call refs: %+v", fim.createCalls[0].Refs)
	}
}

func TestCreateJob_PropagatesError(t *testing.T) {
	wantErr := errors.New("no schedulable nodes")
	fim := &fakeIM{createErr: wantErr}
	p := imgpull.New(testutil.NewFakeClient(),
		func(context.Context, *appv1.ApplicationManager, *appcfg.ApplicationConfig) ([]appv1.Ref, error) {
			return nil, nil
		},
		imgpull.WithImageManager(fim),
	)
	am := &appv1.ApplicationManager{ObjectMeta: metav1.ObjectMeta{Name: "am"}}
	err := p.CreateJob(context.Background(), am, nil)
	if !errors.Is(err, wantErr) {
		t.Errorf("err: %v want %v", err, wantErr)
	}
}

// -----------------------------------------------------------------
// CancelJob
// -----------------------------------------------------------------

func TestCancelJob_EmptyName(t *testing.T) {
	fim := &fakeIM{}
	p := imgpull.New(testutil.NewFakeClient(),
		func(context.Context, *appv1.ApplicationManager, *appcfg.ApplicationConfig) ([]appv1.Ref, error) {
			return nil, nil
		},
		imgpull.WithImageManager(fim),
	)
	err := p.CancelJob(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for empty name")
	}
	if len(fim.updateCalls) != 0 {
		t.Errorf("empty-name CancelJob leaked: %+v", fim.updateCalls)
	}
}

func TestCancelJob_SetsDownloadingCanceled(t *testing.T) {
	fim := &fakeIM{}
	p := imgpull.New(testutil.NewFakeClient(),
		func(context.Context, *appv1.ApplicationManager, *appcfg.ApplicationConfig) ([]appv1.Ref, error) {
			return nil, nil
		},
		imgpull.WithImageManager(fim),
	)
	if err := p.CancelJob(context.Background(), "am"); err != nil {
		t.Fatalf("CancelJob: %v", err)
	}
	if len(fim.updateCalls) != 1 {
		t.Fatalf("updateCalls: %d want 1", len(fim.updateCalls))
	}
	call := fim.updateCalls[0]
	wantState := appv1.DownloadingCanceled.String()
	if call.Name != "am" {
		t.Errorf("name: %q want am", call.Name)
	}
	if call.State != wantState {
		t.Errorf("state: %q want %q", call.State, wantState)
	}
	if call.Message != wantState {
		t.Errorf("message: %q want %q (legacy convention)", call.Message, wantState)
	}
}

// -----------------------------------------------------------------
// UpdateJobStatus
// -----------------------------------------------------------------

func TestUpdateJobStatus_EmptyName(t *testing.T) {
	fim := &fakeIM{}
	p := imgpull.New(testutil.NewFakeClient(),
		func(context.Context, *appv1.ApplicationManager, *appcfg.ApplicationConfig) ([]appv1.Ref, error) {
			return nil, nil
		},
		imgpull.WithImageManager(fim),
	)
	err := p.UpdateJobStatus(context.Background(), "", "downloading", "hi")
	if err == nil {
		t.Fatal("expected error for empty name")
	}
	if len(fim.updateCalls) != 0 {
		t.Errorf("empty-name UpdateJobStatus leaked: %+v", fim.updateCalls)
	}
}

func TestUpdateJobStatus_Passthrough(t *testing.T) {
	fim := &fakeIM{}
	p := imgpull.New(testutil.NewFakeClient(),
		func(context.Context, *appv1.ApplicationManager, *appcfg.ApplicationConfig) ([]appv1.Ref, error) {
			return nil, nil
		},
		imgpull.WithImageManager(fim),
	)
	if err := p.UpdateJobStatus(context.Background(), "am", "downloading", "starting"); err != nil {
		t.Fatalf("UpdateJobStatus: %v", err)
	}
	if len(fim.updateCalls) != 1 {
		t.Fatalf("updateCalls: %d want 1", len(fim.updateCalls))
	}
	call := fim.updateCalls[0]
	if call.Name != "am" || call.State != "downloading" || call.Message != "starting" {
		t.Errorf("passthrough: %+v", call)
	}
}
