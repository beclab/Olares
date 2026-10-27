package podhealth

import (
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func waitingPod(name, container, reason string, restarts int32) corev1.Pod {
	return corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec:       corev1.PodSpec{Containers: []corev1.Container{{Name: container}}},
		Status: corev1.PodStatus{
			ContainerStatuses: []corev1.ContainerStatus{{
				Name:         container,
				RestartCount: restarts,
				State:        corev1.ContainerState{Waiting: &corev1.ContainerStateWaiting{Reason: reason}},
			}},
		},
	}
}

func TestRun_PodStatusTiers(t *testing.T) {
	tests := []struct {
		name       string
		pods       []corev1.Pod
		wantOK     bool
		wantReason string
	}{
		{"empty", nil, false, ""},
		{"image pull backoff", []corev1.Pod{waitingPod("p", "c", "ImagePullBackOff", 0)}, true, ReasonImagePull},
		{"err image pull", []corev1.Pod{waitingPod("p", "c", "ErrImagePull", 0)}, true, ReasonImagePull},
		{"invalid image", []corev1.Pod{waitingPod("p", "c", "InvalidImageName", 0)}, true, ReasonImagePull},
		{"create container config error", []corev1.Pod{waitingPod("p", "c", "CreateContainerConfigError", 0)}, true, ReasonContainerConfig},
		{"create container error", []corev1.Pod{waitingPod("p", "c", "CreateContainerError", 0)}, true, ReasonContainerConfig},
		{"run container error", []corev1.Pod{waitingPod("p", "c", "RunContainerError", 0)}, true, ReasonContainerConfig},
		{"crashloop below threshold", []corev1.Pod{waitingPod("p", "c", "CrashLoopBackOff", 4)}, false, ""},
		{"crashloop at threshold", []corev1.Pod{waitingPod("p", "c", "CrashLoopBackOff", 5)}, true, ReasonCrashLoop},
		{"container creating is not unrecoverable", []corev1.Pod{waitingPod("p", "c", "ContainerCreating", 0)}, false, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sig, ok := Run(Input{Pods: tt.pods})
			if ok != tt.wantOK {
				t.Fatalf("Run ok = %v, want %v", ok, tt.wantOK)
			}
			if ok && sig.Reason != tt.wantReason {
				t.Fatalf("Run reason = %q, want %q", sig.Reason, tt.wantReason)
			}
			if ok && sig.Grace != HardErrorGrace {
				t.Fatalf("Run grace = %v, want %v", sig.Grace, HardErrorGrace)
			}
		})
	}
}

func TestRun_Unschedulable(t *testing.T) {
	pod := corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "p"},
		Status: corev1.PodStatus{
			Conditions: []corev1.PodCondition{{
				Type:    corev1.PodScheduled,
				Status:  corev1.ConditionFalse,
				Reason:  corev1.PodReasonUnschedulable,
				Message: "0/3 nodes available",
			}},
		},
	}
	sig, ok := Run(Input{Pods: []corev1.Pod{pod}})
	if !ok || sig.Reason != ReasonUnschedulable {
		t.Fatalf("expected unschedulable hit, got ok=%v reason=%q", ok, sig.Reason)
	}
}

func TestRun_InitContainer(t *testing.T) {
	pod := corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "p"},
		Status: corev1.PodStatus{
			InitContainerStatuses: []corev1.ContainerStatus{{
				Name:  "init",
				State: corev1.ContainerState{Waiting: &corev1.ContainerStateWaiting{Reason: "ImagePullBackOff"}},
			}},
		},
	}
	if _, ok := Run(Input{Pods: []corev1.Pod{pod}}); !ok {
		t.Fatal("expected init container ImagePullBackOff to be a hit")
	}
}

func TestRun_PermanentMount(t *testing.T) {
	now := time.Now()
	mkEvent := func(reason, msg string, age time.Duration) corev1.Event {
		return corev1.Event{
			InvolvedObject: corev1.ObjectReference{Name: "p"},
			Reason:         reason,
			Message:        msg,
			LastTimestamp:  metav1.NewTime(now.Add(-age)),
		}
	}
	// A pod that has not started any container, so the mount checker inspects
	// its events.
	stuckPod := corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "p"},
		Spec:       corev1.PodSpec{Containers: []corev1.Container{{Name: "c"}}},
		Status: corev1.PodStatus{
			Phase: corev1.PodPending,
			ContainerStatuses: []corev1.ContainerStatus{{
				Name:  "c",
				State: corev1.ContainerState{Waiting: &corev1.ContainerStateWaiting{Reason: "ContainerCreating"}},
			}},
		},
	}
	run := func(events []corev1.Event) (Signal, bool) {
		return Run(Input{
			Pods:        []corev1.Pod{stuckPod},
			Now:         now,
			FetchEvents: func(corev1.Pod) ([]corev1.Event, error) { return events, nil },
		})
	}

	t.Run("graced failure has mount grace", func(t *testing.T) {
		sig, ok := run([]corev1.Event{mkEvent("FailedMount", `secret "s" not found`, 10*time.Second)})
		if !ok || sig.Reason != ReasonPermanentMount || sig.Grace != MountFailureGrace {
			t.Fatalf("expected graced mount, got ok=%v reason=%q grace=%v", ok, sig.Reason, sig.Grace)
		}
	})

	t.Run("hostpath failure is immediate", func(t *testing.T) {
		sig, ok := run([]corev1.Event{mkEvent("FailedMount", `hostPath type check failed: /dev/ttyUSB0 is not a character device`, 10*time.Second)})
		if !ok || sig.Grace != 0 {
			t.Fatalf("expected immediate mount (grace 0), got ok=%v grace=%v", ok, sig.Grace)
		}
	})

	t.Run("stale event ignored", func(t *testing.T) {
		if _, ok := run([]corev1.Event{mkEvent("FailedMount", `secret "s" not found`, 10*time.Minute)}); ok {
			t.Fatal("expected stale FailedMount to be ignored")
		}
	})

	t.Run("nil FetchEvents skips mount tier", func(t *testing.T) {
		if _, ok := Run(Input{Pods: []corev1.Pod{stuckPod}, Now: now}); ok {
			t.Fatal("expected no hit when FetchEvents is nil")
		}
	})

	// Regression: a pod with init containers blocked on volume mounting reports
	// every container as PodInitializing, never ContainerCreating. Gating the
	// event scan on ContainerCreating missed this entirely.
	t.Run("init-container pod stuck in PodInitializing is scanned", func(t *testing.T) {
		initializing := func(name string) corev1.ContainerStatus {
			return corev1.ContainerStatus{
				Name:  name,
				State: corev1.ContainerState{Waiting: &corev1.ContainerStateWaiting{Reason: "PodInitializing"}},
			}
		}
		pod := corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{Name: "homeassistant-557f6444f-fb4d6"},
			Spec: corev1.PodSpec{
				InitContainers: []corev1.Container{{Name: "check-auth"}},
				Containers:     []corev1.Container{{Name: "homeassistant"}},
			},
			Status: corev1.PodStatus{
				Phase:                 corev1.PodPending,
				InitContainerStatuses: []corev1.ContainerStatus{initializing("check-auth")},
				ContainerStatuses:     []corev1.ContainerStatus{initializing("homeassistant")},
			},
		}
		events := []corev1.Event{mkEvent("FailedMount",
			`MountVolume.SetUp failed for volume "usb-device" : hostPath type check failed: /dev/ttyUSB0 is not a character device`,
			63*time.Second)}
		sig, ok := Run(Input{
			Pods:        []corev1.Pod{pod},
			Now:         now,
			FetchEvents: func(corev1.Pod) ([]corev1.Event, error) { return events, nil },
		})
		if !ok || sig.Reason != ReasonPermanentMount || sig.Grace != 0 {
			t.Fatalf("expected immediate permanent-mount hit, got ok=%v reason=%q grace=%v", ok, sig.Reason, sig.Grace)
		}
	})

	t.Run("pod with a running container is skipped", func(t *testing.T) {
		pod := corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{Name: "p"},
			Spec:       corev1.PodSpec{Containers: []corev1.Container{{Name: "c"}}},
			Status: corev1.PodStatus{
				Phase: corev1.PodPending,
				ContainerStatuses: []corev1.ContainerStatus{{
					Name:  "c",
					State: corev1.ContainerState{Running: &corev1.ContainerStateRunning{}},
				}},
			},
		}
		called := false
		_, ok := Run(Input{
			Pods: []corev1.Pod{pod},
			Now:  now,
			FetchEvents: func(corev1.Pod) ([]corev1.Event, error) {
				called = true
				return nil, nil
			},
		})
		if ok || called {
			t.Fatalf("expected running-container pod to be skipped, ok=%v fetched=%v", ok, called)
		}
	})
}

func TestClassifyMountFailure(t *testing.T) {
	tests := []struct {
		name string
		msg  string
		want mountKind
	}{
		{"secret not found is graced", `MountVolume.SetUp failed for volume "cfg" : secret "my-secret" not found`, mountGraced},
		{"configmap not found is graced", `MountVolume.SetUp failed for volume "cfg" : configmap "my-cm" not found`, mountGraced},
		{"hostpath type check is immediate", `MountVolume.SetUp failed for volume "usb-device" : hostPath type check failed: /dev/ttyUSB0 is not a character device`, mountImmediate},
		{"subpath prepare is immediate", `failed to prepare subPath for volumeMount "data" of container "app"`, mountImmediate},
		{"subpath no such file is immediate", `MountVolume.SetUp failed for volume "data" : subPath "conf" not found: no such file or directory`, mountImmediate},
		{"attach timeout is transient", `Unable to attach or mount volumes: unmounted volumes=[data], timed out waiting for the condition`, mountNone},
		{"multi-attach is transient", `Multi-Attach error for volume "pvc-x" Volume is already exclusively attached to one node`, mountNone},
		{"unrelated", `Successfully pulled image`, mountNone},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := classifyMountFailure(tt.msg); got != tt.want {
				t.Fatalf("classifyMountFailure(%q) = %v, want %v", tt.msg, got, tt.want)
			}
		})
	}
}

func TestScanPermanentMountFailure(t *testing.T) {
	now := time.Now()
	mkEvent := func(reason, msg string, age time.Duration) corev1.Event {
		return corev1.Event{
			InvolvedObject: corev1.ObjectReference{Name: "p"},
			Reason:         reason,
			Message:        msg,
			LastTimestamp:  metav1.NewTime(now.Add(-age)),
		}
	}

	t.Run("recent graced failure hits as graced", func(t *testing.T) {
		events := []corev1.Event{mkEvent("FailedMount", `secret "s" not found`, 10*time.Second)}
		_, kind, ok := scanPermanentMountFailure(events, now, DefaultMountEventRecency)
		if !ok || kind != mountGraced {
			t.Fatalf("expected graced hit, got kind=%v ok=%v", kind, ok)
		}
	})

	t.Run("immediate wins over graced", func(t *testing.T) {
		events := []corev1.Event{
			mkEvent("FailedMount", `secret "s" not found`, 5*time.Second),
			mkEvent("FailedMount", `failed to prepare subPath for volumeMount "data"`, 5*time.Second),
		}
		_, kind, ok := scanPermanentMountFailure(events, now, DefaultMountEventRecency)
		if !ok || kind != mountImmediate {
			t.Fatalf("expected immediate to win, got kind=%v ok=%v", kind, ok)
		}
	})

	t.Run("stale event ignored", func(t *testing.T) {
		events := []corev1.Event{mkEvent("FailedMount", `secret "s" not found`, 10*time.Minute)}
		if _, _, ok := scanPermanentMountFailure(events, now, DefaultMountEventRecency); ok {
			t.Fatal("expected stale FailedMount to be ignored")
		}
	})

	t.Run("transient reason ignored", func(t *testing.T) {
		events := []corev1.Event{mkEvent("FailedMount", `timed out waiting for the condition`, 5*time.Second)}
		if _, _, ok := scanPermanentMountFailure(events, now, DefaultMountEventRecency); ok {
			t.Fatal("expected transient FailedMount to be ignored")
		}
	})

	t.Run("non FailedMount reason ignored", func(t *testing.T) {
		events := []corev1.Event{mkEvent("FailedScheduling", `secret "s" not found`, 5*time.Second)}
		if _, _, ok := scanPermanentMountFailure(events, now, DefaultMountEventRecency); ok {
			t.Fatal("expected non-FailedMount reason to be ignored")
		}
	})

	t.Run("eventTime fallback when lastTimestamp zero", func(t *testing.T) {
		e := corev1.Event{
			InvolvedObject: corev1.ObjectReference{Name: "p"},
			Reason:         "FailedMount",
			Message:        `configmap "c" not found`,
			EventTime:      metav1.NewMicroTime(now.Add(-5 * time.Second)),
		}
		if _, _, ok := scanPermanentMountFailure([]corev1.Event{e}, now, DefaultMountEventRecency); !ok {
			t.Fatal("expected eventTime fallback to be honored")
		}
	})
}

func TestGraceTracker(t *testing.T) {
	t.Run("immediate signal is fatal at once", func(t *testing.T) {
		var g GraceTracker
		_, fatal, started := g.Observe(Signal{Reason: ReasonPermanentMount, Grace: 0}, true)
		if !fatal || started {
			t.Fatalf("immediate: fatal=%v started=%v, want true/false", fatal, started)
		}
	})

	t.Run("graced signal waits then fires", func(t *testing.T) {
		var g GraceTracker
		sig := Signal{Reason: ReasonImagePull, Grace: 10 * time.Millisecond}
		if _, fatal, started := g.Observe(sig, true); fatal || !started {
			t.Fatalf("first tick: fatal=%v started=%v, want false/true", fatal, started)
		}
		time.Sleep(20 * time.Millisecond)
		if _, fatal, _ := g.Observe(sig, true); !fatal {
			t.Fatal("expected fatal after grace elapsed")
		}
	})

	t.Run("healthy tick resets the window", func(t *testing.T) {
		var g GraceTracker
		sig := Signal{Reason: ReasonImagePull, Grace: time.Hour}
		g.Observe(sig, true)
		if _, _, started := g.Observe(Signal{}, false); started {
			t.Fatal("reset tick should not report started")
		}
		if _, _, started := g.Observe(sig, true); !started {
			t.Fatal("expected a fresh window after reset")
		}
	})

	t.Run("changed reason restarts the window", func(t *testing.T) {
		var g GraceTracker
		g.Observe(Signal{Reason: ReasonImagePull, Grace: time.Hour}, true)
		_, fatal, started := g.Observe(Signal{Reason: ReasonCrashLoop, Grace: time.Hour}, true)
		if fatal || !started {
			t.Fatalf("reason change: fatal=%v started=%v, want false/true", fatal, started)
		}
	})
}
