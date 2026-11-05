package fake

import (
	"context"
	"sync"

	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/ports"
)

// EventRecorder is an in-memory ports.EventRecorder. It stores every
// call it receives, in call order, so tests can assert both:
//
//   - that a specific event was emitted (AssertAppEvent)
//   - that no unexpected events were emitted (Events(); len==N)
//   - that the event contents exactly match a per-field expectation
//
// Zero value is legal and starts empty. All methods are safe for
// concurrent use — the Runner routes EmitEvent effects in the same
// goroutine as CallWorkload callback dispatch, so a fake operating
// under contention is a realistic test surface.
type EventRecorder struct {
	mu sync.Mutex

	appEvents     []ports.AppLifecycleEvent
	notifications []RecordedNotification

	// RecordAppEventFunc, when non-nil, is invoked before the event
	// is stored. Returning a non-nil error skips the append and
	// surfaces the error to the caller. Use it to inject failures
	// (e.g. NATS unavailability) into a specific test case without
	// forking the fake.
	RecordAppEventFunc func(ctx context.Context, ev ports.AppLifecycleEvent) error

	// RecordNotificationFunc is the analog for RecordNotification.
	RecordNotificationFunc func(ctx context.Context, subject string, payload any) error
}

// RecordedNotification is the tuple RecordNotification receives.
// Exposed so tests can range over the recorder's Notifications()
// output with named fields.
type RecordedNotification struct {
	Subject string
	Payload any
}

// RecordAppEvent implements ports.EventRecorder.
func (r *EventRecorder) RecordAppEvent(ctx context.Context, ev ports.AppLifecycleEvent) error {
	if r.RecordAppEventFunc != nil {
		if err := r.RecordAppEventFunc(ctx, ev); err != nil {
			return err
		}
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.appEvents = append(r.appEvents, ev)
	return nil
}

// RecordNotification implements ports.EventRecorder.
func (r *EventRecorder) RecordNotification(ctx context.Context, subject string, payload any) error {
	if r.RecordNotificationFunc != nil {
		if err := r.RecordNotificationFunc(ctx, subject, payload); err != nil {
			return err
		}
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.notifications = append(r.notifications, RecordedNotification{Subject: subject, Payload: payload})
	return nil
}

// Events returns a copy of every RecordAppEvent call, in call order.
// A copy (not a slice into the internal buffer) so tests can range
// over it without holding the fake's mutex.
func (r *EventRecorder) Events() []ports.AppLifecycleEvent {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]ports.AppLifecycleEvent, len(r.appEvents))
	copy(out, r.appEvents)
	return out
}

// Notifications returns a copy of every RecordNotification call.
func (r *EventRecorder) Notifications() []RecordedNotification {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]RecordedNotification, len(r.notifications))
	copy(out, r.notifications)
	return out
}

// Reset clears every recorded call. Convenient between sub-tests
// that share a recorder.
func (r *EventRecorder) Reset() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.appEvents = nil
	r.notifications = nil
}

// FindAppEvent returns the first event whose Name and State fields
// match, and a boolean indicating a hit. Kept minimal — tests
// looking for richer matches should range over Events() and match
// per field, since bolting more predicates onto FindAppEvent turns
// this file into a DSL.
func (r *EventRecorder) FindAppEvent(name, state string) (ports.AppLifecycleEvent, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, ev := range r.appEvents {
		if ev.Name == name && ev.State == state {
			return ev, true
		}
	}
	return ports.AppLifecycleEvent{}, false
}

// Interface conformance guard.
var _ ports.EventRecorder = (*EventRecorder)(nil)
