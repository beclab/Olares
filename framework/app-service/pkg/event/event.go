package event

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/beclab/Olares/framework/app-service/pkg/utils"

	"github.com/nats-io/nats.go"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"
)

var AppEventQueue *QueuedEventController

type QueuedEventController struct {
	wq    workqueue.RateLimitingInterface
	ctx   context.Context
	nc    *nats.Conn
	ncMux sync.Mutex
}

type QueueEvent struct {
	Subject string
	Data    interface{}
}

type UserEvent struct {
	Topic   string  `json:"topic"`
	Payload Payload `json:"payload"`
}

type Payload struct {
	User      string    `json:"user"`
	Operator  string    `json:"operator"`
	Timestamp time.Time `json:"timestamp"`
}

func (qe *QueuedEventController) processNextWorkItem() bool {
	obj, shutdown := qe.wq.Get()
	if shutdown {
		return false
	}
	defer qe.wq.Done(obj)
	qe.process(obj)
	qe.wq.Forget(obj)
	return true
}

func (qe *QueuedEventController) process(obj interface{}) {
	eobj, ok := obj.(*QueueEvent)
	if !ok {
		return
	}
	for {
		err := qe.publish(eobj.Subject, eobj.Data)
		if err == nil {
			klog.Infof("publish event success data: %#v", eobj.Data)
			return
		}
		klog.Errorf("publish subject %s, data %v failed: %v", eobj.Subject, eobj.Data, err)
		select {
		case <-qe.ctx.Done():
			return
		case <-time.After(time.Second):

		}
	}
}

func (qe *QueuedEventController) worker() {
	for qe.processNextWorkItem() {

	}
}

func (qe *QueuedEventController) Run() {
	defer utilruntime.HandleCrash()
	defer qe.wq.ShuttingDown()
	go wait.Until(qe.worker, time.Second, qe.ctx.Done())
	klog.Infof("started event publish worker......")
	<-qe.ctx.Done()
	klog.Infof("shutting down queue worker......")
}

func (qe *QueuedEventController) enqueue(obj interface{}) {
	qe.wq.Add(obj)
}

func (qe *QueuedEventController) publish(subject string, data interface{}) error {
	if err := qe.ensureNatsConnected(); err != nil {
		return fmt.Errorf("failed to ensure NATS connection: %w", err)
	}
	return utils.PublishEvent(qe.nc, subject, data)
}

func (qe *QueuedEventController) ensureNatsConnected() error {
	qe.ncMux.Lock()
	defer qe.ncMux.Unlock()

	if qe.nc != nil && qe.nc.IsConnected() {
		return nil
	}
	if qe.nc != nil {
		qe.nc.Close()
	}

	klog.Info("NATS connection not established, attempting to connect...")
	nc, err := utils.NewNatsConn()
	if err != nil {
		klog.Errorf("Failed to connect to NATS: %v", err)
		return err
	}

	qe.nc = nc
	klog.Info("Successfully connected to NATS")
	return nil
}

func (qe *QueuedEventController) GetNatsConn() *nats.Conn {
	qe.ncMux.Lock()
	defer qe.ncMux.Unlock()
	return qe.nc
}

func NewAppEventQueue(ctx context.Context, nc *nats.Conn) *QueuedEventController {
	return &QueuedEventController{
		ctx: ctx,
		nc:  nc,
		wq:  workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "app-event-queue"),
	}
}

func SetAppEventQueue(q *QueuedEventController) {
	AppEventQueue = q
}

func PublishAppEventToQueue(p utils.EventParams) {
	now := time.Now()
	rawAppName := p.RawAppName
	if rawAppName == "" {
		rawAppName = p.Name
	}

	// buildEvent constructs the per-recipient event payload. The
	// `recipient` is the user the event is being routed to; it becomes
	// both the User field and the NATS subject suffix. For per-user
	// events (v1 and v3+per-user) this is always the nominal Owner; for
	// shared-app fan-out it is each activated user in turn.
	buildEvent := func(recipient string) utils.Event {
		ev := utils.Event{
			EventID:         fmt.Sprintf("%s-%s-%d", recipient, p.Name, now.UnixMilli()),
			CreateTime:      now,
			Name:            p.Name,
			Type:            p.Type,
			OpType:          p.OpType,
			OpID:            p.OpID,
			State:           p.State,
			Progress:        p.Progress,
			User:            recipient,
			RawAppName:      rawAppName,
			Title:           p.Title,
			Icon:            p.Icon,
			Reason:          p.Reason,
			Message:         p.Message,
			SharedEntrances: p.SharedEntrances,
			MarketSource:    p.MarketSource,
			ChartOwner:      p.ChartOwner,
		}
		if len(p.EntranceStatuses) > 0 {
			ev.EntranceStatuses = p.EntranceStatuses
		}
		return ev
	}

	// Shared apps are cluster-wide singletons; their lifecycle is
	// relevant to every user that can open the app, not just the
	// nominal owner. Fan the event out to every activated user so each
	// per-user NATS subscriber sees the state change. Unactivated users
	// (mid-wizard) are intentionally skipped — they have no UI to
	// surface the message yet.
	//
	// The activated-user set comes from the in-memory activeusers cache
	// (kept current by the UserController informer handler) so this
	// path performs no kube API I/O. If the cache happens to be empty
	// (e.g. an event was somehow published before the User informer
	// finished its initial sync), we deliberately drop the shared event
	// rather than fall back to a single-owner publish: a per-user style
	// publish to a shared app's nominal Owner would be silently
	// misrouted for everyone else.
	//if p.IsShared {
	//	recipients := activeusers.List()
	//	if len(recipients) == 0 {
	//		klog.Infof("shared-app fan-out: no activated users (cache empty), dropping event for app %s", p.Name)
	//		return
	//	}
	//	for _, u := range recipients {
	//		AppEventQueue.enqueue(&QueueEvent{
	//			Subject: fmt.Sprintf("os.application.%s", u),
	//			Data:    buildEvent(u),
	//		})
	//	}
	//	return
	//}

	AppEventQueue.enqueue(&QueueEvent{
		Subject: fmt.Sprintf("os.application.%s", p.Owner),
		Data:    buildEvent(p.Owner),
	})
}

// PublishToQueue enqueues an arbitrary message for asynchronous, retried
// delivery to NATS. It reuses the single shared JetStream connection held
// by AppEventQueue (see ensureNatsConnected) instead of dialing a new one,
// so callers that previously opened their own connection should use this.
func PublishToQueue(subject string, data interface{}) {
	AppEventQueue.enqueue(&QueueEvent{Subject: subject, Data: data})
}

func PublishUserEventToQueue(topic, user, operator string) {
	subject := "os.users"
	data := UserEvent{
		Topic: topic,
		Payload: Payload{
			User:      user,
			Operator:  operator,
			Timestamp: time.Now(),
		},
	}
	AppEventQueue.enqueue(&QueueEvent{Subject: subject, Data: data})
}
