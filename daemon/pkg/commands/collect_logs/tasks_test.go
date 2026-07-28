package collectlogs

import (
	"fmt"
	"testing"
	"time"
)

func resetTasks() {
	tasksMu.Lock()
	defer tasksMu.Unlock()
	tasks = map[string]*TaskStatus{}
	taskOrder = nil
}

func TestRegisterAndGetTask(t *testing.T) {
	resetTasks()
	registerTask(&TaskStatus{RunID: "r1", Caller: "alice", State: "in-progress", StartedAt: time.Now()})

	got, ok := GetTask("r1")
	if !ok {
		t.Fatalf("task r1 not found after register")
	}
	if got.Caller != "alice" || got.State != "in-progress" {
		t.Errorf("unexpected task: %+v", got)
	}
}

func TestGetUnknownTask(t *testing.T) {
	resetTasks()
	if _, ok := GetTask("nope"); ok {
		t.Fatalf("expected unknown runID to return false")
	}
}

func TestUpdateTask(t *testing.T) {
	resetTasks()
	registerTask(&TaskStatus{RunID: "r1", State: "in-progress"})
	finished := time.Now()
	updateTask("r1", func(t *TaskStatus) {
		t.State = "completed"
		t.Error = "partial: [worker-1 timeout]"
		t.FinishedAt = &finished
		t.Nodes = []TaskNodeStatus{{Name: "master", Status: "ok"}}
	})

	got, _ := GetTask("r1")
	if got.State != "completed" || got.Error == "" || got.FinishedAt == nil {
		t.Errorf("update not applied: %+v", got)
	}
	if len(got.Nodes) != 1 || got.Nodes[0].Name != "master" {
		t.Errorf("nodes not applied: %+v", got.Nodes)
	}

	// Updating an evicted/unknown run is a no-op, not a panic.
	updateTask("missing", func(t *TaskStatus) { t.State = "x" })
}

func TestGetTaskReturnsCopy(t *testing.T) {
	resetTasks()
	finished := time.Now()
	registerTask(&TaskStatus{
		RunID:      "r1",
		State:      "completed",
		FinishedAt: &finished,
		Nodes:      []TaskNodeStatus{{Name: "master", Status: "ok"}},
	})

	got, _ := GetTask("r1")
	got.State = "tampered"
	got.Nodes[0].Name = "tampered"
	*got.FinishedAt = time.Unix(0, 0)

	fresh, _ := GetTask("r1")
	if fresh.State != "completed" {
		t.Errorf("returned copy aliased State: %q", fresh.State)
	}
	if fresh.Nodes[0].Name != "master" {
		t.Errorf("returned copy aliased Nodes slice: %q", fresh.Nodes[0].Name)
	}
	if fresh.FinishedAt.Equal(time.Unix(0, 0)) {
		t.Errorf("returned copy aliased FinishedAt pointer")
	}
}

func TestEvictionKeepsNewest(t *testing.T) {
	resetTasks()
	total := maxTrackedTasks + 10
	for i := 0; i < total; i++ {
		registerTask(&TaskStatus{RunID: fmt.Sprintf("r%03d", i), State: "in-progress"})
	}

	tasksMu.Lock()
	n := len(tasks)
	order := len(taskOrder)
	tasksMu.Unlock()
	if n != maxTrackedTasks || order != maxTrackedTasks {
		t.Fatalf("registry not bounded: tasks=%d order=%d want %d", n, order, maxTrackedTasks)
	}

	// Oldest 10 evicted, newest maxTrackedTasks retained.
	for i := 0; i < 10; i++ {
		if _, ok := GetTask(fmt.Sprintf("r%03d", i)); ok {
			t.Errorf("expected r%03d evicted", i)
		}
	}
	for i := total - maxTrackedTasks; i < total; i++ {
		if _, ok := GetTask(fmt.Sprintf("r%03d", i)); !ok {
			t.Errorf("expected r%03d retained", i)
		}
	}
}
