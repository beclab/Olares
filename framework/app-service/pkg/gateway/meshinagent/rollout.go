package meshinagent

import (
	"sync"
	"time"
)

const (
	// MaxConcurrentRuleBumpRollouts is U3 K=2.
	MaxConcurrentRuleBumpRollouts = 2
	RolloutMaxRetries             = 5
	RolloutBackoffInitial         = 10 * time.Second
	RolloutBackoffCap             = 300 * time.Second

	// AnnotRolloutFingerprint records last successful rollout fingerprint (idempotency).
	AnnotRolloutFingerprint = "gateway.olares.io/mesh-in-rollout-fp"
	// AnnotRolloutStatus is "ok" or ErrRolloutFailed after a rollout attempt.
	AnnotRolloutStatus = "gateway.olares.io/mesh-in-rollout-status"
	// AnnotRolloutReason is the last enqueue reason (audit).
	AnnotRolloutReason = "gateway.olares.io/mesh-in-rollout-reason"

	// ErrRolloutFailed is the stable failure code (ARCH R7-1); declare facts are not rolled back.
	ErrRolloutFailed = "MESH_IN_ROLLOUT_FAILED"

	ReasonDecideFalseToTrue = "decide_false_to_true"
	ReasonDecideTrueToFalse = "decide_true_to_false"
	ReasonDecideEdges       = "decide_edges_changed"
	ReasonMeshReady  = "mesh_ready"
	ReasonAppCreateInject   = "app_create_inject"

	RestartedAtAnnotation = "kubectl.kubernetes.io/restartedAt"

	RolloutStatusOK = "ok"

	// MeshInjectRolloutStateCM lives in os-framework (app-service home).
	MeshInjectRolloutStateCMNamespace = "os-framework"
	MeshInjectRolloutStateCMName      = "olares-mesh-inject-rollout"
	MeshInjectStateReadyKey           = "ready"
	MeshInjectStateEpochKey           = "epoch"

	EGDataplaneNamespace = "os-gateway"
)

// RolloutQueue limits concurrent rule-bump / decide-change rollouts.
type RolloutQueue struct {
	mu      sync.Mutex
	active  int
	max     int
	waiting []string
	// granted holds keys whose slot was already counted into active by Release,
	// so the waiter picking that key up must not consume a second slot.
	granted map[string]struct{}
}

// NewRolloutQueue returns a queue with the given concurrency cap.
func NewRolloutQueue(max int) *RolloutQueue {
	if max <= 0 {
		max = MaxConcurrentRuleBumpRollouts
	}
	return &RolloutQueue{max: max, granted: map[string]struct{}{}}
}

// DefaultRolloutQueue is the process-wide bump queue (K=2).
var DefaultRolloutQueue = NewRolloutQueue(MaxConcurrentRuleBumpRollouts)

// TryAcquire returns true if a rollout slot was acquired for key.
func (q *RolloutQueue) TryAcquire(key string) bool {
	q.mu.Lock()
	defer q.mu.Unlock()
	if _, ok := q.granted[key]; ok {
		delete(q.granted, key)
		return true
	}
	if q.active >= q.max {
		for _, w := range q.waiting {
			if w == key {
				return false
			}
		}
		q.waiting = append(q.waiting, key)
		return false
	}
	q.active++
	return true
}

// Release frees a slot and optionally returns the next waiting key.
func (q *RolloutQueue) Release() (next string, ok bool) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.active > 0 {
		q.active--
	}
	if len(q.waiting) == 0 {
		return "", false
	}
	next = q.waiting[0]
	q.waiting = q.waiting[1:]
	q.active++
	if q.granted == nil {
		q.granted = map[string]struct{}{}
	}
	q.granted[next] = struct{}{}
	return next, true
}

// ActiveCount is for tests.
func (q *RolloutQueue) ActiveCount() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.active
}

// WaitingCount is for tests.
func (q *RolloutQueue) WaitingCount() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return len(q.waiting)
}

// RetryBackoff returns the sleep duration for attempt (0-based) with exponential *2 capped.
func RetryBackoff(attempt int) time.Duration {
	d := RolloutBackoffInitial
	for i := 0; i < attempt; i++ {
		d *= 2
		if d >= RolloutBackoffCap {
			return RolloutBackoffCap
		}
	}
	if d > RolloutBackoffCap {
		return RolloutBackoffCap
	}
	return d
}

// RolloutFingerprint builds an idempotency token for decide+edges+ready epoch.
func RolloutFingerprint(inject bool, edges string, readyEpoch string) string {
	inj := "false"
	if inject {
		inj = "true"
	}
	if readyEpoch == "" {
		readyEpoch = "0"
	}
	return "inject=" + inj + "|edges=" + edges + "|epoch=" + readyEpoch
}
