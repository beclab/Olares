package collectlogs

import (
	"sync"
	"time"
)

// maxTrackedTasks bounds the in-memory task registry so a long-lived daemon
// does not accumulate run records without limit. Oldest entries are evicted
// first.
const maxTrackedTasks = 50

// StateNotFound is reported for a runID the master no longer tracks (evicted,
// never existed, or lost on master restart). Returned with HTTP 200 so the
// client handles it as a normal terminal state rather than a transport error.
const StateNotFound = "not-found"

// StateForbidden is reported when the caller may not view a tracked run (not
// the run's owner and not owner/admin). Returned with HTTP 200 carrying no
// run details, so the client handles it uniformly as a state rather than a
// transport error.
const StateForbidden = "forbidden"

// TaskNodeStatus is the per-node outcome recorded for a run.
type TaskNodeStatus struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	Err    string `json:"err,omitempty"`
}

// TaskStatus is the per-runID view of a collection, queryable while runs
// execute concurrently. State reuses the global ProcessingState values
// (in-progress/completed/failed).
type TaskStatus struct {
	RunID      string           `json:"runID"`
	Caller     string           `json:"caller"`
	State      string           `json:"state"`
	Error      string           `json:"error,omitempty"`
	File       string           `json:"file"`
	Path       string           `json:"path"`
	StartedAt  time.Time        `json:"startedAt"`
	FinishedAt *time.Time       `json:"finishedAt,omitempty"`
	Nodes      []TaskNodeStatus `json:"nodes,omitempty"`
}

var (
	tasksMu   sync.Mutex
	tasks     = map[string]*TaskStatus{}
	taskOrder []string
)

// registerTask records a freshly started run, evicting the oldest tracked
// tasks past maxTrackedTasks.
func registerTask(t *TaskStatus) {
	tasksMu.Lock()
	defer tasksMu.Unlock()

	if _, exists := tasks[t.RunID]; !exists {
		taskOrder = append(taskOrder, t.RunID)
	}
	tasks[t.RunID] = t

	for len(taskOrder) > maxTrackedTasks {
		oldest := taskOrder[0]
		taskOrder = taskOrder[1:]
		delete(tasks, oldest)
	}
}

// updateTask applies fn to the tracked task under lock. It is a no-op if the
// run was already evicted.
func updateTask(runID string, fn func(*TaskStatus)) {
	tasksMu.Lock()
	defer tasksMu.Unlock()
	if t, ok := tasks[runID]; ok {
		fn(t)
	}
}

// GetTask returns a deep copy of the tracked task so callers never observe
// concurrent mutation by the orchestration goroutine.
func GetTask(runID string) (TaskStatus, bool) {
	tasksMu.Lock()
	defer tasksMu.Unlock()
	t, ok := tasks[runID]
	if !ok {
		return TaskStatus{}, false
	}
	out := *t
	if t.FinishedAt != nil {
		ft := *t.FinishedAt
		out.FinishedAt = &ft
	}
	if t.Nodes != nil {
		out.Nodes = append([]TaskNodeStatus(nil), t.Nodes...)
	}
	return out, true
}
