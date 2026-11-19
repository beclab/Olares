package ports

import (
	"context"
	"time"

	appv1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
)

// EventRecorder is the sole publisher of user-facing app lifecycle
// events. It replaces the process-global pkg/event.AppEventQueue
// singleton and the various free-form utils.PublishEvent helpers.
//
// Legacy pipeline (to be dismantled in the cutover step):
//
//	handler / controller
//	        │  push
//	        ▼
//	pkg/event.AppEventQueue (global)   ── worker goroutine (per-process)
//	                                   ── utils.PublishEvent → NATS JetStream
//
// New pipeline:
//
//	handler returns domain.EmitEvent  ── Runner  ── EventRecorder port
//	                                                       │
//	                                              adapters/natsrecorder
//	                                                       │
//	                                                     NATS
//
// The port intentionally exposes TWO methods rather than a generic
// Publish because the two legacy usages have different subjects,
// payload contracts, and downstream consumers:
//
//   - RecordAppEvent  → subject "os.application.{owner}"
//     payload = utils.Event shape (preserved for UI)
//     consumers: per-user NATS subscribers in bfl
//   - RecordNotification → subject "os.notification" (compute alerts)
//     payload = compute/notification resourceNotification
//     consumers: system notification service
//
// Splitting them at the port level keeps the schema contract in
// front of the reader and prevents accidental cross-wiring.
type EventRecorder interface {
	// RecordAppEvent enqueues an AppLifecycleEvent for the
	// per-owner subject. RUNNER-ONLY WRITE (dispatched from
	// domain.EmitEvent). Non-blocking — the adapter owns any
	// internal queue and retry loop.
	RecordAppEvent(ctx context.Context, ev AppLifecycleEvent) error

	// RecordNotification publishes an arbitrary payload to `subject`.
	// RUNNER-ONLY WRITE (dispatched from InstallValidator or a
	// dedicated Effect variant). Kept generic because the payload
	// schema is defined by the subject's downstream consumer, not
	// by this package.
	RecordNotification(ctx context.Context, subject string, payload any) error
}

// AppLifecycleEvent mirrors utils.Event (pkg/utils/event.go). Field
// names are preserved so the JSON serialization on the NATS subject
// stays byte-compatible with existing UI consumers — do NOT rename
// without coordinating with bfl / frontend.
type AppLifecycleEvent struct {
	EventID          string
	Name             string
	RawAppName       string
	Type             string
	OpType           string
	OpID             string
	State            string
	Progress         string
	User             string
	Title            string
	Icon             string
	Reason           string
	Message          string
	MarketSource     string
	ChartOwner       string
	EntranceStatuses []appv1.EntranceStatus
	CreateTime       time.Time
}
