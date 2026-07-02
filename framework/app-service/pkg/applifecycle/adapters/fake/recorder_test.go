package fake_test

import (
	"context"
	"errors"
	"testing"

	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/adapters/fake"
	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/ports"
)

func TestEventRecorder_RecordAndFind(t *testing.T) {
	var r fake.EventRecorder

	if err := r.RecordAppEvent(context.Background(), ports.AppLifecycleEvent{
		Name: "myapp", State: "Installing",
	}); err != nil {
		t.Fatalf("RecordAppEvent: %v", err)
	}
	if err := r.RecordAppEvent(context.Background(), ports.AppLifecycleEvent{
		Name: "myapp", State: "Running",
	}); err != nil {
		t.Fatalf("RecordAppEvent: %v", err)
	}

	if got := r.Events(); len(got) != 2 {
		t.Fatalf("Events() len = %d, want 2", len(got))
	}

	// Find hits both by (name, state) — pins the primary usage
	// pattern in handler tests.
	if _, ok := r.FindAppEvent("myapp", "Running"); !ok {
		t.Fatal(`FindAppEvent("myapp","Running") = false, want true`)
	}
	if _, ok := r.FindAppEvent("myapp", "Failed"); ok {
		t.Fatal(`FindAppEvent("myapp","Failed") = true, want false`)
	}
}

// TestEventRecorder_HookOverride pins the "inject failure per test"
// pattern that Runner error-path tests will depend on.
func TestEventRecorder_HookOverride(t *testing.T) {
	sentinel := errors.New("nats unavailable")
	r := &fake.EventRecorder{
		RecordAppEventFunc: func(_ context.Context, _ ports.AppLifecycleEvent) error {
			return sentinel
		},
	}
	err := r.RecordAppEvent(context.Background(), ports.AppLifecycleEvent{Name: "x"})
	if !errors.Is(err, sentinel) {
		t.Fatalf("RecordAppEvent err = %v, want %v", err, sentinel)
	}
	if got := r.Events(); len(got) != 0 {
		t.Fatalf("Events() len = %d, want 0 (hook rejected the append)", len(got))
	}
}

// TestEventRecorder_Reset pins the between-cases cleanup helper.
func TestEventRecorder_Reset(t *testing.T) {
	var r fake.EventRecorder
	_ = r.RecordAppEvent(context.Background(), ports.AppLifecycleEvent{Name: "a"})
	_ = r.RecordNotification(context.Background(), "os.notification", "hi")

	r.Reset()
	if got := r.Events(); len(got) != 0 {
		t.Fatalf("after Reset, Events() len = %d, want 0", len(got))
	}
	if got := r.Notifications(); len(got) != 0 {
		t.Fatalf("after Reset, Notifications() len = %d, want 0", len(got))
	}
}

// TestEventRecorder_EventsCopy pins that Events() returns a copy —
// mutating the returned slice must NOT affect subsequent calls.
func TestEventRecorder_EventsCopy(t *testing.T) {
	var r fake.EventRecorder
	_ = r.RecordAppEvent(context.Background(), ports.AppLifecycleEvent{Name: "a"})

	got := r.Events()
	got[0].Name = "hacked"

	fresh := r.Events()
	if fresh[0].Name != "a" {
		t.Fatalf("Events() copy leaked mutation: got %q, want %q", fresh[0].Name, "a")
	}
}
