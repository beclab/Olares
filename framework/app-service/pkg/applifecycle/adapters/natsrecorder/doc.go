// Package natsrecorder implements ports.EventRecorder on top of the
// legacy pkg/event.QueuedEventController — the same NATS-JetStream
// publisher the existing (non-Runner) code path uses.
//
// # What this replaces
//
// Legacy fan-in to the queue:
//
//	handler / controller
//	   ├─ event.PublishAppEventToQueue(utils.EventParams{…})
//	   └─ event.PublishToQueue(subject, payload)
//	         │
//	         ▼
//	QueuedEventController.enqueue → worker → utils.PublishEvent → NATS
//
// New Runner-side fan-in via this adapter:
//
//	handler returns domain.EmitEvent{Kind, Payload}
//	   │
//	Runner (effects loop)
//	   │
//	natsrecorder.Recorder.RecordAppEvent / .RecordNotification
//	   │
//	QueuedEventController.Enqueue → worker → utils.PublishEvent → NATS
//
// The adapter is deliberately a thin translator; the queue owns NATS
// connectivity, backoff, and retry (see pkg/event/event.go). Two
// consequences worth flagging:
//
//  1. RecordAppEvent / RecordNotification are non-blocking: they
//     return nil once the QueueEvent has been added to the worker
//     queue. Delivery failures surface in the queue worker's logs
//     (retry loop), not in the Runner. This matches every existing
//     legacy call site (all of which are fire-and-forget) so the
//     observable behaviour on the NATS side is unchanged.
//
//  2. The adapter accepts a QueueEnqueuer interface, not the
//     concrete *event.QueuedEventController, so tests can inject a
//     fake queue that records the enqueued events without spinning
//     up NATS. *event.QueuedEventController naturally satisfies the
//     interface via the new Enqueue method added alongside this
//     adapter.
//
// # Field mapping
//
// ports.AppLifecycleEvent → utils.Event, one field at a time. The
// two shapes are almost identical; the one shape diff is that
// utils.Event carries SharedEntrances (a v3-shared-app fanout hint)
// which no ported handler currently emits. If a handler ever needs
// to emit shared-entrance metadata, add the field to
// ports.AppLifecycleEvent AND to the copy in RecordAppEvent below —
// the two must stay lockstep so the JSON on the NATS subject stays
// byte-compatible with the bfl / frontend consumers.
//
// # When to wire this into main.go
//
// NOT NOW. Runner effects fire on every state transition the state
// machine decides to make. As long as k8sstore is in shadow-mode
// (writes suppressed), the state machine can decide to transition
// but nothing actually persists — wiring natsrecorder live would
// therefore emit real NATS events for ghost state changes the rest
// of the system never sees. Duplicated/spurious "app is now
// Downloading" events would surface on the UI.
//
// The correct wire-in point is the step-3 cutover PR that ALSO
// flips k8sstore from shadow-mode to live. Wiring both in the same
// PR keeps the "Runner-emitted event ↔ Runner-persisted state"
// invariant true throughout the flip. Until then main.go keeps
// using &fake.EventRecorder{} and this adapter exists as a tested,
// ready-to-swap unit.
package natsrecorder
