package natsrecorder

import (
	"context"
	"fmt"

	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/ports"
	"github.com/beclab/Olares/framework/app-service/pkg/event"
	"github.com/beclab/Olares/framework/app-service/pkg/utils"
)

// QueueEnqueuer is the tiny fan-in surface Recorder depends on.
// *event.QueuedEventController satisfies it via the Enqueue method
// added alongside this package. Kept as an interface so tests can
// substitute a slice-backed fake without spinning up NATS or a
// controller-runtime workqueue.
type QueueEnqueuer interface {
	Enqueue(*event.QueueEvent)
}

// Recorder is the production ports.EventRecorder adapter.
type Recorder struct {
	queue QueueEnqueuer
}

// New constructs a Recorder. `queue` MUST be non-nil — a nil queue
// would silently drop every event the state machine emits, which
// is worse than a startup panic.
func New(queue QueueEnqueuer) *Recorder {
	if queue == nil {
		panic("natsrecorder.New: queue is nil")
	}
	return &Recorder{queue: queue}
}

// RecordAppEvent implements ports.EventRecorder. It maps an
// AppLifecycleEvent onto the legacy utils.Event payload shape and
// enqueues it on the per-owner subject `os.application.{User}`.
//
// The subject computation is preserved from
// event.PublishAppEventToQueue verbatim — bfl subscribes by exact
// subject, so any drift here becomes a silent no-op on the UI.
//
// ctx is intentionally unused: the queue owns its own lifecycle
// (cancellation via the QueuedEventController's own ctx passed at
// construction time). Adding a per-call ctx here would suggest
// per-event cancellation that the queue does not honour.
func (r *Recorder) RecordAppEvent(_ context.Context, ev ports.AppLifecycleEvent) error {
	if ev.User == "" {
		// A per-user subject with empty user would end up as
		// "os.application." which bfl never subscribes to.
		// Fail loudly rather than silently misroute — this is
		// almost certainly a handler bug (User not populated
		// from AM.Spec.AppOwner).
		return fmt.Errorf("natsrecorder: RecordAppEvent requires User to be set")
	}
	subject := fmt.Sprintf("os.application.%s", ev.User)
	payload := utils.Event{
		EventID:          ev.EventID,
		CreateTime:       ev.CreateTime,
		Name:             ev.Name,
		RawAppName:       ev.RawAppName,
		Type:             ev.Type,
		OpType:           ev.OpType,
		OpID:             ev.OpID,
		State:            ev.State,
		Progress:         ev.Progress,
		User:             ev.User,
		EntranceStatuses: ev.EntranceStatuses,
		Title:            ev.Title,
		Icon:             ev.Icon,
		Reason:           ev.Reason,
		Message:          ev.Message,
		MarketSource:     ev.MarketSource,
		ChartOwner:       ev.ChartOwner,
		// SharedEntrances intentionally not copied — no ported
		// handler emits shared-entrance metadata; see doc.go
		// "Field mapping" for the sync-point contract if a
		// handler ever needs it.
	}
	r.queue.Enqueue(&event.QueueEvent{Subject: subject, Data: payload})
	return nil
}

// RecordNotification implements ports.EventRecorder. Non-app subjects
// pass through untouched — the payload schema is defined by the
// downstream consumer per port doc, not by this adapter.
func (r *Recorder) RecordNotification(_ context.Context, subject string, payload any) error {
	if subject == "" {
		return fmt.Errorf("natsrecorder: RecordNotification requires a non-empty subject")
	}
	r.queue.Enqueue(&event.QueueEvent{Subject: subject, Data: payload})
	return nil
}

var _ ports.EventRecorder = (*Recorder)(nil)
