package event

import (
	"context"
	"fmt"
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
	wq  workqueue.RateLimitingInterface
	ctx context.Context
	nc  *nats.Conn
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

	if err := qe.process(obj); err != nil {
		// Requeue with rate limiting on failure
		qe.wq.AddRateLimited(obj)
		return true
	}

	// Successfully processed, forget about it
	qe.wq.Forget(obj)
	return true
}

func (qe *QueuedEventController) process(obj interface{}) error {
	eobj, ok := obj.(*QueueEvent)
	if !ok {
		return nil
	}
	err := qe.publish(eobj.Subject, eobj.Data)
	if err != nil {
		klog.Errorf("async publish subject %s, data %v, failed %v, will retry", eobj.Subject, eobj.Data, err)
		return err
	}
	klog.Infof("publish event success data: %#v", eobj.Data)
	return nil
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
	return utils.PublishEvent(qe.nc, subject, data)
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
	subject := fmt.Sprintf("os.application.%s", p.Owner)

	now := time.Now()
	data := utils.Event{
		EventID:    fmt.Sprintf("%s-%s-%d", p.Owner, p.Name, now.UnixMilli()),
		CreateTime: now,
		Name:       p.Name,
		Type:       p.Type,
		OpType:     p.OpType,
		OpID:       p.OpID,
		State:      p.State,
		Progress:   p.Progress,
		User:       p.Owner,
		RawAppName: func() string {
			if p.RawAppName == "" {
				return p.Name
			}
			return p.RawAppName
		}(),
		Title:           p.Title,
		Reason:          p.Reason,
		Message:         p.Message,
		SharedEntrances: p.SharedEntrances,
	}
	if len(p.EntranceStatuses) > 0 {
		data.EntranceStatuses = p.EntranceStatuses
	}

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
