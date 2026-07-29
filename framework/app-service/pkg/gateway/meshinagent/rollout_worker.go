package meshinagent

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"

	appv1alpha1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
)

// RolloutJob is one unit of work for the rate-limited rollout worker.
type RolloutJob struct {
	Key           string
	Reason        string
	AppNamespace  string
	AppName       string // Spec.Name (workload label), empty for shared/eg-only
	AppObjectName string // Application CR name when patching fingerprint
	Workloads     []WorkloadRef
	Fingerprint   string
	Inject        bool
	Edges         string
	ReadyEpoch   string
}

// RolloutWorker processes mesh-in rollouts with concurrency K and retry backoff.
type RolloutWorker struct {
	Client client.Client
	Queue  *RolloutQueue

	mu      sync.Mutex
	pending map[string]RolloutJob
	wake    chan struct{}
	started bool

	// sleepFn allows tests to skip real backoff sleeps.
	sleepFn func(context.Context, time.Duration)
	// completeFn reports rollout completion; tests substitute a deterministic one.
	completeFn func(context.Context, WorkloadRef) (bool, error)
}

var (
	defaultWorkerMu sync.RWMutex
	defaultWorker   *RolloutWorker
)

// SetDefaultWorker installs the process-wide worker (called from main).
func SetDefaultWorker(w *RolloutWorker) {
	defaultWorkerMu.Lock()
	defer defaultWorkerMu.Unlock()
	defaultWorker = w
}

// DefaultWorker returns the process-wide worker (may be nil before main wires it).
func DefaultWorker() *RolloutWorker {
	defaultWorkerMu.RLock()
	defer defaultWorkerMu.RUnlock()
	return defaultWorker
}

// NewRolloutWorker constructs a worker using DefaultRolloutQueue when q is nil.
func NewRolloutWorker(c client.Client, q *RolloutQueue) *RolloutWorker {
	if q == nil {
		q = DefaultRolloutQueue
	}
	w := &RolloutWorker{
		Client:  c,
		Queue:   q,
		pending: map[string]RolloutJob{},
		wake:    make(chan struct{}, 1),
		sleepFn: func(ctx context.Context, d time.Duration) {
			t := time.NewTimer(d)
			defer t.Stop()
			select {
			case <-ctx.Done():
			case <-t.C:
			}
		},
	}
	w.completeFn = func(ctx context.Context, ref WorkloadRef) (bool, error) {
		return WorkloadRolloutComplete(ctx, w.Client, ref)
	}
	return w
}

// Start runs the worker loop until ctx is cancelled.
func (w *RolloutWorker) Start(ctx context.Context) {
	if w == nil {
		return
	}
	w.mu.Lock()
	if w.started {
		w.mu.Unlock()
		return
	}
	w.started = true
	w.mu.Unlock()
	go w.loop(ctx)
}

func (w *RolloutWorker) loop(ctx context.Context) {
	for {
		job, ok := w.dequeue()
		if !ok {
			select {
			case <-ctx.Done():
				return
			case <-w.wake:
			}
			continue
		}
		// Bound concurrency: wait until TryAcquire succeeds.
		for {
			if w.Queue.TryAcquire(job.Key) {
				break
			}
			select {
			case <-ctx.Done():
				return
			case <-time.After(500 * time.Millisecond):
			}
		}
		go func(j RolloutJob) {
			defer func() {
				w.Queue.Release()
				w.signal()
			}()
			w.runJob(ctx, j)
		}(job)
	}
}

func (w *RolloutWorker) signal() {
	select {
	case w.wake <- struct{}{}:
	default:
	}
}

func (w *RolloutWorker) dequeue() (RolloutJob, bool) {
	w.mu.Lock()
	defer w.mu.Unlock()
	for k, j := range w.pending {
		delete(w.pending, k)
		return j, true
	}
	return RolloutJob{}, false
}

// Enqueue stores/replaces a job by key and wakes the loop.
func (w *RolloutWorker) Enqueue(job RolloutJob) {
	if w == nil || job.Key == "" {
		return
	}
	w.mu.Lock()
	w.pending[job.Key] = job
	w.mu.Unlock()
	w.signal()
}

// EnqueueAppRollout lists app workloads and enqueues a decide-driven rollout.
func (w *RolloutWorker) EnqueueAppRollout(ctx context.Context, app *appv1alpha1.Application, reason, readyEpoch string) {
	if w == nil || app == nil {
		return
	}
	inject := DeclaresSharedCaller(app.Spec.Settings)
	if !inject && app.Annotations != nil {
		inject = strings.EqualFold(strings.TrimSpace(app.Annotations[AnnotDecide]), "true")
	}
	edges := ""
	if app.Spec.Settings != nil {
		edges = strings.TrimSpace(app.Spec.Settings[AnnotDecideEdges])
	}
	if edges == "" && app.Annotations != nil {
		edges = strings.TrimSpace(app.Annotations[AnnotDecideEdges])
	}
	fp := RolloutFingerprint(inject, edges, readyEpoch)
	if app.Annotations != nil &&
		app.Annotations[AnnotRolloutFingerprint] == fp &&
		app.Annotations[AnnotRolloutStatus] == RolloutStatusOK {
		klog.V(2).Infof("mesh-in-rollout: skip app %s fp unchanged", app.Name)
		return
	}
	refs, err := ListAppWorkloads(ctx, w.Client, app.Spec.Namespace, app.Spec.Name)
	if err != nil {
		klog.Errorf("mesh-in-rollout: list workloads for app %s failed: %v", app.Name, err)
		return
	}
	refs = FilterWorkloadsNeedingInject(ctx, w.Client, refs, SharedCallerInject(inject))
	if len(refs) == 0 {
		klog.V(2).Infof("mesh-in-rollout: nothing to roll for app %s/%s", app.Spec.Namespace, app.Spec.Name)
		return
	}
	w.Enqueue(RolloutJob{
		Key:           "app/" + app.Namespace + "/" + app.Name,
		Reason:        reason,
		AppNamespace:  app.Spec.Namespace,
		AppName:       app.Spec.Name,
		AppObjectName: app.Name,
		Workloads:     refs,
		Fingerprint:   fp,
		Inject:        inject,
		Edges:         edges,
		ReadyEpoch:   readyEpoch,
	})
}

// EnqueueMeshReadySweep enqueues decide=true Applications, Shared inject NS, and EG dataplane.
func (w *RolloutWorker) EnqueueMeshReadySweep(ctx context.Context, epoch string) {
	if w == nil {
		return
	}
	apps, err := ListDecideTrueApplications(ctx, w.Client)
	if err != nil {
		klog.Errorf("mesh-in-rollout: mesh-ready sweep list apps failed: %v", err)
	} else {
		for i := range apps {
			w.EnqueueAppRollout(ctx, &apps[i], ReasonMeshReady, epoch)
		}
	}
	shared, err := ListSharedInjectWorkloads(ctx, w.Client)
	if err != nil {
		klog.Errorf("mesh-in-rollout: mesh-ready sweep shared failed: %v", err)
	}
	eg, err := ListEGDataplaneWorkloads(ctx, w.Client)
	if err != nil {
		klog.Errorf("mesh-in-rollout: mesh-ready sweep eg failed: %v", err)
	}
	// One job per workload keeps a large Shared namespace from occupying a whole
	// slot and rolling every workload in it at once.
	dataplane := FilterWorkloadsNeedingInject(ctx, w.Client, append(shared, eg...), LinkerdOnlyInject())
	for _, ref := range dataplane {
		w.Enqueue(RolloutJob{
			Key:         "dataplane/" + epoch + "/" + ref.Key(),
			Reason:      ReasonMeshReady,
			Workloads:   []WorkloadRef{ref},
			ReadyEpoch: epoch,
		})
	}
	klog.Infof("mesh-in-rollout: sweep enqueued epoch=%s apps=%d dataplane=%d",
		epoch, len(apps), len(dataplane))
}

// runJob bumps the pod templates once, then polls until the replacement pods are
// ready. Success is only recorded after completion so a stalled rollout keeps its
// retry budget instead of being fingerprinted as done.
func (w *RolloutWorker) runJob(ctx context.Context, job RolloutJob) {
	var lastErr error
	bumped := false
	for attempt := 0; attempt < RolloutMaxRetries; attempt++ {
		if attempt > 0 {
			d := RetryBackoff(attempt - 1)
			klog.Warningf("mesh-in-rollout: retry key=%s attempt=%d backoff=%s lastErr=%v",
				job.Key, attempt, d, lastErr)
			w.sleepFn(ctx, d)
			if ctx.Err() != nil {
				klog.Warningf("mesh-in-rollout: abort key=%s: %v", job.Key, ctx.Err())
				return
			}
		}
		if !bumped {
			if lastErr = w.bumpAll(ctx, job.Workloads); lastErr != nil {
				continue
			}
			bumped = true
		}
		done, err := w.rolloutComplete(ctx, job.Workloads)
		if err != nil {
			lastErr = err
			continue
		}
		if done {
			w.markSuccess(ctx, job)
			return
		}
		lastErr = fmt.Errorf("rollout still in progress")
	}
	klog.Errorf("mesh-in-rollout: %s key=%s reason=%s after %d attempts: %v",
		ErrRolloutFailed, job.Key, job.Reason, RolloutMaxRetries, lastErr)
	w.markFailed(ctx, job)
}

func (w *RolloutWorker) bumpAll(ctx context.Context, refs []WorkloadRef) error {
	var first error
	for _, ref := range refs {
		if err := BumpWorkload(ctx, w.Client, ref); err != nil {
			if first == nil {
				first = err
			}
		}
	}
	return first
}

func (w *RolloutWorker) rolloutComplete(ctx context.Context, refs []WorkloadRef) (bool, error) {
	if w.completeFn == nil {
		return true, nil
	}
	for _, ref := range refs {
		done, err := w.completeFn(ctx, ref)
		if err != nil {
			klog.Warningf("mesh-in-rollout: %s rollout status unknown: %v", ref.Key(), err)
			return false, err
		}
		if !done {
			return false, nil
		}
	}
	return true, nil
}

func (w *RolloutWorker) markSuccess(ctx context.Context, job RolloutJob) {
	if job.AppObjectName == "" || w.Client == nil {
		return
	}
	var app appv1alpha1.Application
	if err := w.Client.Get(ctx, types.NamespacedName{Name: job.AppObjectName}, &app); err != nil {
		klog.Errorf("mesh-in-rollout: get app %s to mark ok failed: %v", job.AppObjectName, err)
		return
	}
	base := app.DeepCopy()
	if app.Annotations == nil {
		app.Annotations = map[string]string{}
	}
	app.Annotations[AnnotRolloutFingerprint] = job.Fingerprint
	app.Annotations[AnnotRolloutStatus] = RolloutStatusOK
	app.Annotations[AnnotRolloutReason] = job.Reason
	if err := w.Client.Patch(ctx, &app, client.MergeFrom(base)); err != nil {
		klog.Errorf("mesh-in-rollout: patch app %s success annot failed: %v", job.AppObjectName, err)
	}
}

func (w *RolloutWorker) markFailed(ctx context.Context, job RolloutJob) {
	if job.AppObjectName == "" || w.Client == nil {
		return
	}
	var app appv1alpha1.Application
	if err := w.Client.Get(ctx, types.NamespacedName{Name: job.AppObjectName}, &app); err != nil {
		klog.Errorf("mesh-in-rollout: get app %s to mark failed failed: %v", job.AppObjectName, err)
		return
	}
	base := app.DeepCopy()
	if app.Annotations == nil {
		app.Annotations = map[string]string{}
	}
	// Keep fingerprint unset/mismatched so a later reconcile can retry; do NOT rollback Decide.
	app.Annotations[AnnotRolloutStatus] = ErrRolloutFailed
	app.Annotations[AnnotRolloutReason] = job.Reason
	if err := w.Client.Patch(ctx, &app, client.MergeFrom(base)); err != nil {
		klog.Errorf("mesh-in-rollout: patch app %s fail annot failed: %v", job.AppObjectName, err)
	}
}

// LoadMeshInjectRolloutState reads ready flag and epoch from the state ConfigMap.
func LoadMeshInjectRolloutState(ctx context.Context, c client.Client) (ready bool, epoch string, err error) {
	if c == nil {
		return false, "0", nil
	}
	var cm corev1.ConfigMap
	if err := c.Get(ctx, types.NamespacedName{
		Namespace: MeshInjectRolloutStateCMNamespace,
		Name:      MeshInjectRolloutStateCMName,
	}, &cm); err != nil {
		if apierrors.IsNotFound(err) {
			return false, "0", nil
		}
		klog.Errorf("mesh-in-rollout: get mesh-inject state cm failed: %v", err)
		return false, "0", err
	}
	ready = strings.EqualFold(strings.TrimSpace(cm.Data[MeshInjectStateReadyKey]), "true")
	epoch = strings.TrimSpace(cm.Data[MeshInjectStateEpochKey])
	if epoch == "" {
		epoch = "0"
	}
	return ready, epoch, nil
}

// StoreMeshInjectRolloutState persists ready/epoch (create or update).
func StoreMeshInjectRolloutState(ctx context.Context, c client.Client, ready bool, epoch string) error {
	if c == nil {
		return nil
	}
	if epoch == "" {
		epoch = "0"
	}
	readyStr := "false"
	if ready {
		readyStr = "true"
	}
	var cm corev1.ConfigMap
	err := c.Get(ctx, types.NamespacedName{
		Namespace: MeshInjectRolloutStateCMNamespace,
		Name:      MeshInjectRolloutStateCMName,
	}, &cm)
	if apierrors.IsNotFound(err) {
		cm = corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: MeshInjectRolloutStateCMNamespace,
				Name:      MeshInjectRolloutStateCMName,
			},
			Data: map[string]string{
				MeshInjectStateReadyKey: readyStr,
				MeshInjectStateEpochKey: epoch,
			},
		}
		if cerr := c.Create(ctx, &cm); cerr != nil {
			klog.Errorf("mesh-in-rollout: create mesh-inject state cm failed: %v", cerr)
			return cerr
		}
		return nil
	}
	if err != nil {
		klog.Errorf("mesh-in-rollout: get mesh-inject state cm failed: %v", err)
		return err
	}
	base := cm.DeepCopy()
	if cm.Data == nil {
		cm.Data = map[string]string{}
	}
	cm.Data[MeshInjectStateReadyKey] = readyStr
	cm.Data[MeshInjectStateEpochKey] = epoch
	if err := c.Patch(ctx, &cm, client.MergeFrom(base)); err != nil {
		klog.Errorf("mesh-in-rollout: patch mesh-inject state cm failed: %v", err)
		return err
	}
	return nil
}

// NextMeshReadyEpoch increments a numeric epoch string.
func NextMeshReadyEpoch(cur string) string {
	n, err := strconv.Atoi(strings.TrimSpace(cur))
	if err != nil || n < 0 {
		n = 0
	}
	return strconv.Itoa(n + 1)
}

// MaybeEnqueueAfterDecide is called from Application reconcile after Decide patch.
// prevInject/prevEdges are pre-patch values; inject/edges are post-Decide.
func MaybeEnqueueAfterDecide(ctx context.Context, app *appv1alpha1.Application, prevInject, inject bool, prevEdges, edges, reasonHint string) {
	w := DefaultWorker()
	if w == nil || app == nil {
		return
	}
	var reason string
	switch {
	case inject && !prevInject:
		reason = ReasonDecideFalseToTrue
	case !inject && prevInject:
		reason = ReasonDecideTrueToFalse
	case inject && prevInject && prevEdges != edges:
		reason = ReasonDecideEdges
	default:
		return
	}
	if strings.TrimSpace(reasonHint) != "" {
		reason = reasonHint
	}
	_, epoch, err := LoadMeshInjectRolloutState(ctx, w.Client)
	if err != nil {
		klog.Warningf("mesh-in-rollout: load mesh-ready epoch for decide enqueue: %v", err)
		epoch = "0"
	}
	w.EnqueueAppRollout(ctx, app, reason, epoch)
}

// EnqueueCreateIfInject enqueues when a newly created Application declares inject.
func EnqueueCreateIfInject(ctx context.Context, app *appv1alpha1.Application) {
	if app == nil {
		return
	}
	inject := DeclaresSharedCaller(app.Spec.Settings)
	if !inject {
		return
	}
	MaybeEnqueueAfterDecide(ctx, app, false, true, "", strings.TrimSpace(app.Spec.Settings[AnnotDecideEdges]), ReasonAppCreateInject)
}

// FormatRolloutFailedMessage returns a stable diagnostic string.
func FormatRolloutFailedMessage(key string, err error) string {
	return fmt.Sprintf("%s key=%s err=%v", ErrRolloutFailed, key, err)
}
