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
)

// RolloutQueue limits concurrent rule-bump / decide-change rollouts.
type RolloutQueue struct {
	mu        sync.Mutex
	active    int
	max       int
	waiting   []string
}

// NewRolloutQueue returns a queue with the given concurrency cap.
func NewRolloutQueue(max int) *RolloutQueue {
	if max <= 0 {
		max = MaxConcurrentRuleBumpRollouts
	}
	return &RolloutQueue{max: max}
}

// DefaultRolloutQueue is the process-wide bump queue (K=2).
var DefaultRolloutQueue = NewRolloutQueue(MaxConcurrentRuleBumpRollouts)

// TryAcquire returns true if a rollout slot was acquired for key.
func (q *RolloutQueue) TryAcquire(key string) bool {
	q.mu.Lock()
	defer q.mu.Unlock()
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
