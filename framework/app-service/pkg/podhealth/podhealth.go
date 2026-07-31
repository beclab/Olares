// Package podhealth detects pod conditions that will not heal on their own
// during install / upgrade / resume startup waits, so those flows can fail
// fast instead of polling until an outer TTL expires.
//
// It mirrors the chainable shape of pkg/compute/validation: each check is a
// small Checker that inspects an Input and returns a structured Signal. Run
// executes a chain and returns the first hit; adding a new detection later is
// just a new Checker appended to DefaultCheckers. Callers that poll feed each
// tick's Run result into a GraceTracker, which turns a momentary Signal into a
// fatal one only after it has persisted past the Signal's grace window.
package podhealth

import (
	"time"

	corev1 "k8s.io/api/core/v1"
)

// Reason codes surfaced on Signal.Reason. They are machine-readable and stable
// so callers/logs can key on them without parsing Message.
const (
	ReasonUnschedulable   = "Unschedulable"
	ReasonImagePull       = "ImagePullFailure"
	ReasonContainerConfig = "ContainerConfigError"
	ReasonCrashLoop       = "CrashLoopBackOff"
	ReasonPermanentMount  = "PermanentMountFailure"
)

// Grace defaults. Exported as vars (not consts) so tests can shorten them.
var (
	// HardErrorGrace is how long a hard, self-non-healing pod condition
	// (bad image, container config error, unschedulable, deep CrashLoopBackOff)
	// must persist before it is treated as fatal. It tolerates transient
	// crashes during normal startup.
	HardErrorGrace = 5 * time.Minute

	// MountFailureGrace is how long a permanent FailedMount signature must
	// persist before it is fatal. Shorter than HardErrorGrace because these
	// failures are deterministic; the grace only tolerates the brief window
	// where helm is still creating the referenced Secret/ConfigMap
	// concurrently with the workload.
	MountFailureGrace = 3 * time.Minute
)

const (
	// DefaultCrashLoopRestartThreshold is the minimum container restart count
	// for a CrashLoopBackOff pod to be treated as unrecoverable. CrashLoopBackOff
	// backoff caps at 300s, so >=5 restarts means several minutes of failing,
	// which filters out transient crashes during normal startup.
	DefaultCrashLoopRestartThreshold int32 = 5

	// DefaultMountEventRecency bounds how old a FailedMount event may be to
	// still count. It filters out stale events from an already-recovered mount
	// so only currently-recurring failures are considered.
	DefaultMountEventRecency = 90 * time.Second
)

// Signal is the structured outcome of a Checker. Reason is a stable code,
// Message is human-readable detail, and Grace is how long the condition must
// persist before it is fatal (0 means fail immediately). Checker is populated
// by Run with the name of the checker that produced the signal.
type Signal struct {
	Reason  string
	Message string
	Grace   time.Duration
	Checker string
}

// Input bundles everything the checkers inspect. Pods is required; the
// event-based mount checker is skipped unless FetchEvents is supplied, which
// lets each call site plug in its own (clientset) event source and keeps the
// pod-only checkers I/O-free and easy to unit test.
type Input struct {
	Pods []corev1.Pod

	// Now is the reference time for event-recency filtering; Run defaults it
	// to time.Now() when zero.
	Now time.Time

	// CrashLoopRestartThreshold overrides DefaultCrashLoopRestartThreshold when
	// non-zero.
	CrashLoopRestartThreshold int32

	// MountEventRecency overrides DefaultMountEventRecency when non-zero.
	MountEventRecency time.Duration

	// FetchEvents returns the events for a single pod. Optional: when nil the
	// event-based mount checker is a no-op. The mount checker only calls it for
	// pods stuck in ContainerCreating, so callers can back it with a plain
	// clientset List without an informer.
	FetchEvents func(pod corev1.Pod) ([]corev1.Event, error)
}

// Checker inspects an Input and reports whether it found an unrecoverable
// condition. Name identifies it in logs and Signal.Checker.
type Checker interface {
	Name() string
	Check(in Input) (Signal, bool)
}

// Run executes the checkers in order and returns the first hit. Passing no
// checkers runs DefaultCheckers. Defaults for Now / thresholds are applied
// once here so individual checkers stay simple.
func Run(in Input, checkers ...Checker) (Signal, bool) {
	if in.Now.IsZero() {
		in.Now = time.Now()
	}
	if in.CrashLoopRestartThreshold == 0 {
		in.CrashLoopRestartThreshold = DefaultCrashLoopRestartThreshold
	}
	if in.MountEventRecency == 0 {
		in.MountEventRecency = DefaultMountEventRecency
	}
	if len(checkers) == 0 {
		checkers = DefaultCheckers()
	}
	for _, c := range checkers {
		if sig, ok := c.Check(in); ok {
			sig.Checker = c.Name()
			return sig, true
		}
	}
	return Signal{}, false
}

// GraceTracker turns the per-tick Signal from Run into a fatal decision only
// after the signal has persisted past its Grace (or is immediate). It is the
// stateful companion to Run for callers that poll on a ticker. Not safe for
// concurrent use; use one per poll loop.
type GraceTracker struct {
	since  time.Time
	reason string
}

// Observe records a tick's Run result. It returns:
//   - fatal=true when the signal is immediate (Grace<=0) or has persisted past
//     its Grace, meaning the caller should abort;
//   - started=true on the tick that opens a new grace window, so the caller can
//     emit a single warning log rather than one per tick.
//
// A not-detected result resets the tracker so an intervening healthy tick
// clears the grace window.
func (g *GraceTracker) Observe(sig Signal, detected bool) (out Signal, fatal bool, started bool) {
	if !detected {
		g.since = time.Time{}
		g.reason = ""
		return Signal{}, false, false
	}
	if sig.Grace <= 0 {
		return sig, true, false
	}
	if g.since.IsZero() || g.reason != sig.Reason {
		g.since = time.Now()
		g.reason = sig.Reason
		return sig, false, true
	}
	if time.Since(g.since) > sig.Grace {
		return sig, true, false
	}
	return sig, false, false
}
