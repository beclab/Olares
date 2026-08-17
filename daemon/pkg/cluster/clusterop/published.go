package clusterop

import "sync/atomic"

// published is the orchestrator this daemon is running, if it built one.
//
// It exists because the two things that need the manager come from opposite
// directions: the HTTP routes are wired at startup and hold it directly, while
// the upgrade watcher reaches it from inside a command that knows nothing
// about the API server. Threading it through that path would mean giving every
// upgrade command a handle to the cluster orchestrator so that one of them
// may use it.
//
// It is not the module registry, which is a different thing with a similar
// name: that one holds what this daemon can do, this one holds the single
// manager doing it.
//
// A nil value is the normal state on a compute node, and on a control node
// whose state directory could not be opened. Callers must treat it as "this
// daemon does not orchestrate" and fall back, never as an error.
var published atomic.Pointer[Manager]

// Publish makes m the orchestrator this daemon is running.
func Publish(m *Manager) { published.Store(m) }

// Current returns the orchestrator, or nil if this daemon has none.
func Current() *Manager { return published.Load() }
