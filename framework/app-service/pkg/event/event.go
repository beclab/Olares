package event

import (
	"context"
	"fmt"
	"time"

	"bytetrade.io/web3os/app-service/pkg/utils"

	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"
)

var AppEventQueue *QueuedEventController

type QueuedEventController struct {
	wq  workqueue.RateLimitingInterface
	ctx context.Context
}

type QueueEvent struct {
	Subject string
	Data    interface{}
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
	err := utils.PublishToNats(eobj.Subject, eobj.Data)
	if err != nil {
		klog.Errorf("async publish subject %s,data %v, failed %v", eobj.Subject, eobj.Data, err)
	} else {
		klog.Infof("publish event success data: %#v", eobj.Data)
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

func NewAppEventQueue(ctx context.Context) *QueuedEventController {
	return &QueuedEventController{
		ctx: ctx,
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
