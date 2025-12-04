package task

import (
	"sync"
	"time"

	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// Type is a multi-user workqueue defined by terminus.
// It's implements workqueue.RateLimitingInterface.
type Type struct {
	// queue is a map that stores workqueue for different keys, that key is username in terminus.
	queue map[string]workqueue.RateLimitingInterface
	cond  *sync.Cond

	indexes      map[string][]int64
	index        int64
	gM           map[string]bool
	shuttingDown bool
}

// Add insert an item to the queue, marks item as needing processing.
func (q *Type) Add(item interface{}) {
	req, _ := item.(reconcile.Request)
	username := req.Namespace
	q.cond.L.Lock()
	defer q.cond.L.Unlock()
	if _, ok := q.queue[username]; !ok {
		q.queue[username] = workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), username)
	}
	total := q.len()
	q.queue[username].Add(item)
	if q.len() > total {
		q.indexes[username] = append(q.indexes[username], q.index)
		q.index++
		if !q.gM[username] {
			q.cond.Signal()
		}
	}
}

func (q *Type) len() int {
	total := 0
	for i := range q.queue {
		total += q.queue[i].Len()
	}
	return total
}

// Get blocks until it can return an item to be processed.
func (q *Type) Get() (item interface{}, shutdown bool) {
	q.cond.L.Lock()
	defer q.cond.L.Unlock()

	for q.len() == 0 && !q.shuttingDown {
		q.cond.Wait()
	}
	if q.len() == 0 {
		return nil, true
	}
	total := q.len()
	for k, v := range q.indexes {
		if q.gM[k] {
			total -= len(v)
		}
	}
	if total == 0 {
		q.cond.Wait()
	}

	minIndex := int64(1<<63 - 1)
	qKey := ""
	for k, v := range q.indexes {
		if len(v) == 0 || q.gM[k] {
			continue
		}
		if v[0] < minIndex {
			minIndex = v[0]
			qKey = k
		}
	}
	q.indexes[qKey] = q.indexes[qKey][1:]
	v, t := q.queue[qKey].Get()
	q.gM[qKey] = true
	return v, t
}

// SetCompleted marks this item has been processed.
func (q *Type) SetCompleted(item interface{}) {
	q.cond.L.Lock()
	defer q.cond.L.Unlock()
	req, _ := item.(reconcile.Request)
	username := req.Namespace
	delete(q.gM, username)
	if q.queue[username].Len() > 0 {
		q.cond.Signal()
	}
}

// Len returns the sum of all queue.
func (q *Type) Len() int {
	q.cond.L.Lock()
	defer q.cond.L.Unlock()
	return q.len()
}

// Done marks item as done processing, and if it has been marked as dirty again
// while it was being processed, it will be re-added to the queue for
// re-processing.
func (q *Type) Done(item interface{}) {
	q.cond.L.Lock()
	defer q.cond.L.Unlock()
	req, _ := item.(reconcile.Request)
	u := req.Namespace
	total := q.queue[u].Len()
	q.queue[u].Done(item)
	if q.queue[u].Len() > total {
		q.indexes[u] = append(q.indexes[u], q.index)
		q.index++
	}

}

// ShutDown will cause q to ignore all new items added to it and
// immediately instruct the worker goroutines to exit.
func (q *Type) ShutDown() {
	q.cond.L.Lock()
	defer q.cond.L.Unlock()
	for i := range q.queue {
		q.queue[i].ShutDown()
	}
	q.shuttingDown = true
}

// ShutDownWithDrain will cause all q to ignore all new items added to it.
func (q *Type) ShutDownWithDrain() {
	q.cond.L.Lock()
	defer q.cond.L.Unlock()
	for i := range q.queue {
		q.queue[i].ShutDownWithDrain()
	}
	q.shuttingDown = true
}

// ShuttingDown returns true if q is shuttingdown.
func (q *Type) ShuttingDown() bool {
	q.cond.L.Lock()
	defer q.cond.L.Unlock()
	return q.shuttingDown
}

// AddAfter adds an item to the workqueue after the indicated duration has passed
func (q *Type) AddAfter(item interface{}, duration time.Duration) {
	if q.ShuttingDown() {
		return
	}
	req, _ := item.(reconcile.Request)
	q.queue[req.Namespace].AddAfter(item, duration)
}

// AddRateLimited adds an item to the workqueue after the rate limiter says it's ok
func (q *Type) AddRateLimited(item interface{}) {
	req, _ := item.(reconcile.Request)
	q.queue[req.Namespace].AddRateLimited(item)
}

// Forget indicates that an item is finished being retried.  Doesn't matter whether it's for failing
// or for success, we'll stop tracking it.
func (q *Type) Forget(item interface{}) {
	req, _ := item.(reconcile.Request)
	username := req.Namespace
	q.queue[username].Forget(item)
}

// NumRequeues returns back how many times the item was requeued.
func (q *Type) NumRequeues(item interface{}) int {
	req, _ := item.(reconcile.Request)
	return q.queue[req.Namespace].NumRequeues(item)
}

// NewQ initializes the Type which implements workqueue.RateLimitingInterface.
func NewQ() workqueue.RateLimitingInterface {
	return &Type{
		queue:   make(map[string]workqueue.RateLimitingInterface),
		cond:    sync.NewCond(&sync.Mutex{}),
		indexes: make(map[string][]int64),
		gM:      make(map[string]bool),
	}
}

// WQueue is used for application install/upgrade.
var WQueue workqueue.RateLimitingInterface

func init() {
	WQueue = NewQ()
}
