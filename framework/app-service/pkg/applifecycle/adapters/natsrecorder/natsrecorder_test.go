package natsrecorder_test

import (
	"context"
	"testing"
	"time"

	appv1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"

	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/adapters/natsrecorder"
	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/ports"
	"github.com/beclab/Olares/framework/app-service/pkg/event"
	"github.com/beclab/Olares/framework/app-service/pkg/utils"
)

// fakeQueue records every Enqueue call. Deliberately does no
// synchronisation — the Recorder is single-goroutine per call and
// tests are sequential.
type fakeQueue struct {
	events []*event.QueueEvent
}

func (f *fakeQueue) Enqueue(qe *event.QueueEvent) {
	f.events = append(f.events, qe)
}

func TestNew_NilQueuePanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("New(nil): expected panic")
		}
	}()
	_ = natsrecorder.New(nil)
}

func TestRecordAppEvent_SubjectAndPayload(t *testing.T) {
	q := &fakeQueue{}
	r := natsrecorder.New(q)

	created := time.Date(2026, 7, 12, 8, 0, 0, 0, time.UTC)
	ev := ports.AppLifecycleEvent{
		EventID:      "u-app-x-123",
		Name:         "app-x",
		RawAppName:   "app-x-raw",
		Type:         "app",
		OpType:       "install",
		OpID:         "1782000000",
		State:        "downloading",
		Progress:     "42.00",
		User:         "u",
		Title:        "App X",
		Icon:         "https://example.test/x.png",
		Reason:       "InProgress",
		Message:      "pulling image",
		MarketSource: "upload",
		ChartOwner:   "u",
		EntranceStatuses: []appv1.EntranceStatus{
			{Name: "web", State: "running"},
		},
		CreateTime: created,
	}
	if err := r.RecordAppEvent(context.Background(), ev); err != nil {
		t.Fatalf("RecordAppEvent: %v", err)
	}
	if len(q.events) != 1 {
		t.Fatalf("expected 1 enqueue, got %d", len(q.events))
	}
	got := q.events[0]
	if got.Subject != "os.application.u" {
		t.Errorf("subject: got %q want os.application.u", got.Subject)
	}
	payload, ok := got.Data.(utils.Event)
	if !ok {
		t.Fatalf("payload type: got %T want utils.Event", got.Data)
	}
	// Spot-check every field mapping. Full-shape identity would
	// couple the test to any future field addition on utils.Event
	// (which we deliberately DO NOT block), so we check the fields
	// the port promised.
	if payload.EventID != ev.EventID {
		t.Errorf("EventID: %q want %q", payload.EventID, ev.EventID)
	}
	if !payload.CreateTime.Equal(ev.CreateTime) {
		t.Errorf("CreateTime: %v want %v", payload.CreateTime, ev.CreateTime)
	}
	if payload.Name != ev.Name {
		t.Errorf("Name: %q want %q", payload.Name, ev.Name)
	}
	if payload.RawAppName != ev.RawAppName {
		t.Errorf("RawAppName: %q want %q", payload.RawAppName, ev.RawAppName)
	}
	if payload.Type != ev.Type {
		t.Errorf("Type: %q want %q", payload.Type, ev.Type)
	}
	if payload.OpType != ev.OpType {
		t.Errorf("OpType: %q want %q", payload.OpType, ev.OpType)
	}
	if payload.OpID != ev.OpID {
		t.Errorf("OpID: %q want %q", payload.OpID, ev.OpID)
	}
	if payload.State != ev.State {
		t.Errorf("State: %q want %q", payload.State, ev.State)
	}
	if payload.Progress != ev.Progress {
		t.Errorf("Progress: %q want %q", payload.Progress, ev.Progress)
	}
	if payload.User != ev.User {
		t.Errorf("User: %q want %q", payload.User, ev.User)
	}
	if payload.Title != ev.Title {
		t.Errorf("Title: %q want %q", payload.Title, ev.Title)
	}
	if payload.Icon != ev.Icon {
		t.Errorf("Icon: %q want %q", payload.Icon, ev.Icon)
	}
	if payload.Reason != ev.Reason {
		t.Errorf("Reason: %q want %q", payload.Reason, ev.Reason)
	}
	if payload.Message != ev.Message {
		t.Errorf("Message: %q want %q", payload.Message, ev.Message)
	}
	if payload.MarketSource != ev.MarketSource {
		t.Errorf("MarketSource: %q want %q", payload.MarketSource, ev.MarketSource)
	}
	if payload.ChartOwner != ev.ChartOwner {
		t.Errorf("ChartOwner: %q want %q", payload.ChartOwner, ev.ChartOwner)
	}
	if len(payload.EntranceStatuses) != 1 || payload.EntranceStatuses[0].Name != "web" {
		t.Errorf("EntranceStatuses: %+v", payload.EntranceStatuses)
	}
}

func TestRecordAppEvent_EmptyUserRejected(t *testing.T) {
	q := &fakeQueue{}
	r := natsrecorder.New(q)

	err := r.RecordAppEvent(context.Background(), ports.AppLifecycleEvent{
		Name:  "app-x",
		State: "downloading",
		// User omitted on purpose.
	})
	if err == nil {
		t.Fatal("expected error on empty User")
	}
	if len(q.events) != 0 {
		t.Errorf("empty-user event should not enqueue, got %d", len(q.events))
	}
}

func TestRecordNotification_PayloadPassthrough(t *testing.T) {
	q := &fakeQueue{}
	r := natsrecorder.New(q)

	type customPayload struct {
		Kind   string
		Detail string
	}
	payload := customPayload{Kind: "compute", Detail: "hami-unschedulable"}

	if err := r.RecordNotification(context.Background(), "os.notification", payload); err != nil {
		t.Fatalf("RecordNotification: %v", err)
	}
	if len(q.events) != 1 {
		t.Fatalf("expected 1 enqueue, got %d", len(q.events))
	}
	got := q.events[0]
	if got.Subject != "os.notification" {
		t.Errorf("subject: got %q want os.notification", got.Subject)
	}
	// Payload passes through unwrapped — port promises the
	// downstream consumer defines the schema, so the adapter must
	// not shape-cast it.
	if got.Data != payload {
		t.Errorf("payload not passthrough: got %+v want %+v", got.Data, payload)
	}
}

func TestRecordNotification_EmptySubjectRejected(t *testing.T) {
	q := &fakeQueue{}
	r := natsrecorder.New(q)

	err := r.RecordNotification(context.Background(), "", struct{}{})
	if err == nil {
		t.Fatal("expected error on empty subject")
	}
	if len(q.events) != 0 {
		t.Errorf("empty-subject event should not enqueue, got %d", len(q.events))
	}
}

// TestProductionQueueSatisfiesInterface pins the type assertion at
// compile time — a signature drift on *event.QueuedEventController
// breaks THIS test rather than only surfacing when main.go wires
// the real queue in. Ordering the check here (adapter test, not
// pkg/event test) keeps pkg/event free of any dependency on this
// adapter.
func TestProductionQueueSatisfiesInterface(t *testing.T) {
	var _ natsrecorder.QueueEnqueuer = (*event.QueuedEventController)(nil)
}
