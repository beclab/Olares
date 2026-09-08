package router

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

// A relaunch is signalled by one request and observed by the application long
// after that request is over, so the POST cannot be the whole answer.
func TestOnlyASettledObservationEndsTheWait(t *testing.T) {
	cases := map[string]bool{
		"":                       false,
		restartStateIdle:         false,
		restartStateWatching:     false,
		restartStateConfirmed:    true,
		restartStateNoSupervisor: true,
		restartStateUnverified:   true,
	}
	for state, want := range cases {
		got := (&engineRestartStatus{State: state}).settled()
		if got != want {
			t.Fatalf("state %q: settled() = %v, want %v", state, got, want)
		}
	}
}

func TestTheWaitEndsOnTheApplicationsVerdict(t *testing.T) {
	reads := 0
	status := awaitEngineRestart(context.Background(), func(context.Context) (*engineRestartStatus, error) {
		reads++
		return &engineRestartStatus{State: restartStateConfirmed}, nil
	})
	if reads != 1 {
		t.Fatalf("a settled answer should be asked for once, got %d reads", reads)
	}
	if status == nil || status.State != restartStateConfirmed {
		t.Fatalf("got %+v, want the confirmed observation", status)
	}
}

// An application too old to have the route answers 404, and a Router that
// cannot be reached answers nothing. Neither unsends the signal, so the
// command reports what it has rather than failing.
func TestAnUnreadableOutcomeIsNotAFailure(t *testing.T) {
	status := awaitEngineRestart(context.Background(), func(context.Context) (*engineRestartStatus, error) {
		return nil, errors.New("404")
	})
	if status != nil {
		t.Fatalf("got %+v, want no observation rather than an invented one", status)
	}

	var buf bytes.Buffer
	if err := reportEngineRestart(&buf, nil, "Olares/qwen3-4b", ""); err != nil {
		t.Fatalf("reportEngineRestart: %v", err)
	}
	if !strings.Contains(buf.String(), "model status") {
		t.Fatalf("with no verdict the reader needs somewhere to look, got:\n%s", buf.String())
	}
}

// Giving up when the caller does, rather than holding the whole window open.
func TestTheWaitStopsWithItsContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	started := time.Now()
	status := awaitEngineRestart(ctx, func(context.Context) (*engineRestartStatus, error) {
		cancel()
		return &engineRestartStatus{State: restartStateWatching}, nil
	})
	if elapsed := time.Since(started); elapsed > engineRestartPoll {
		t.Fatalf("the wait outlived its context by %s", elapsed)
	}
	if status == nil || status.State != restartStateWatching {
		t.Fatalf("got %+v, want what was known when the caller left", status)
	}
}

// Three settled states, three different situations. Only one of them is the
// relaunch working, and the other two need the reader to do something else.
func TestEachOutcomeSaysWhatToDoNext(t *testing.T) {
	back := time.Now()
	cases := map[string]struct {
		status engineRestartStatus
		app    string
		want   string
	}{
		"loading again": {
			status: engineRestartStatus{State: restartStateConfirmed},
			want:   "loading the weights",
		},
		"answering again": {
			status: engineRestartStatus{State: restartStateConfirmed, ObservedBackAt: &back},
			want:   "answering again",
		},
		"nothing was watching": {
			status: engineRestartStatus{State: restartStateNoSupervisor},
			app:    "llamacppqwen3627bggufv3",
			want:   "market restart llamacppqwen3627bggufv3",
		},
		"nothing was watching, and the application is not named here": {
			status: engineRestartStatus{State: restartStateNoSupervisor},
			want:   "names the application",
		},
		"no verdict was possible": {
			status: engineRestartStatus{State: restartStateUnverified},
			want:   "could not tell",
		},
		"still looking when the window closed": {
			status: engineRestartStatus{State: restartStateWatching},
			want:   "still watching",
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			var buf bytes.Buffer
			if err := reportEngineRestart(&buf, &tc.status, "Olares/qwen3-4b", tc.app); err != nil {
				t.Fatalf("reportEngineRestart: %v", err)
			}
			if !strings.Contains(buf.String(), tc.want) {
				t.Fatalf("should mention %q, got:\n%s", tc.want, buf.String())
			}
		})
	}
}

// The outcome read is a GET, so it must not inherit the warning the signal
// carries: a repeated read relaunches nothing.
func TestReadingTheOutcomeIsNotAWriteToRetryCarefully(t *testing.T) {
	if note := repeatNote("GET", epForModel(epEngineRestart, "Olares/qwen3-4b")); note != "" {
		t.Fatalf("a read should carry no repeat warning, got %q", note)
	}
	if note := repeatNote("POST", epEngineRestart); !strings.Contains(note, "restarts") {
		t.Fatalf("the signal still needs its warning, got %q", note)
	}
}
