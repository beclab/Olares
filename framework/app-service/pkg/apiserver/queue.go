package apiserver

import (
	"context"
	"fmt"
	"runtime/debug"
	"time"

	"github.com/beclab/Olares/framework/app-service/pkg/apiserver/api"

	"github.com/emicklei/go-restful/v3"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"
)

type OpController struct {
	wq  workqueue.RateLimitingInterface
	ctx context.Context
}

type Task struct {
	exec func()
}

func (h *Handler) queued(next func(req *restful.Request, resp *restful.Response)) func(req *restful.Request, resp *restful.Response) {
	return func(req *restful.Request, resp *restful.Response) {
		app := req.PathParameter(ParamAppName)
		klog.Infof("enqueue queue %s .........", app)
		done := make(chan struct{})
		t := &Task{
			exec: func() {
				// always release the waiting HTTP goroutine, and never let a
				// handler panic propagate up to crash the single queue worker.
				defer close(done)
				defer func() {
					if r := recover(); r != nil {
						klog.Errorf("recovered from panic in queued task for app %s: %v\n%s", app, r, debug.Stack())
						api.HandleError(resp, req, fmt.Errorf("internal error processing %s: %v", app, r))
					}
				}()
				next(req, resp)
			},
		}
		h.opController.enqueue(t)
		<-done
	}
}

func (op *OpController) processNextWorkItem() bool {
	obj, shutdown := op.wq.Get()
	if shutdown {
		return false
	}
	defer op.wq.Done(obj)
	op.process(obj)
	op.wq.Forget(obj)
	return true
}

func (op *OpController) process(obj interface{}) {
	eobj, ok := obj.(*Task)
	if !ok {
		return
	}
	eobj.exec()
}

func (op *OpController) worker() {
	for op.processNextWorkItem() {
	}
}

func (op *OpController) run() {
	defer utilruntime.HandleCrash()
	defer op.wq.ShuttingDown()
	go wait.Until(op.worker, time.Second, op.ctx.Done())

	klog.Infof("started queue worker......")
	<-op.ctx.Done()
	klog.Infof("shutting down queue worker......")
}

func (op *OpController) enqueue(obj interface{}) {
	op.wq.Add(obj)
}

func NewQueue(ctx context.Context) *OpController {
	return &OpController{
		ctx: ctx,
		wq:  workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "opqueue"),
	}
}
