package router

import (
	"bytes"
	"strings"
	"testing"
)

// A locally installed application owns a model row from the moment it is
// installed, so `router model list` carries models of applications that are
// stopped, downloading or failed. It used to print those three switches green —
// offered, active, provider active — which is every word the reader needed to
// conclude the opposite of the truth.
//
// The weights are the second half of the same problem, and the harder half to
// see: a container reports running minutes before the model behind it can
// answer, so the rows that matter here are the ones where the application is up
// and the model still is not.
func TestCallableNoteNamesTheOneThingInTheWay(t *testing.T) {
	ptr := func(s string) *string { return &s }

	for _, tc := range []struct {
		name string
		row  adminModelRow
		want string
	}{
		{
			name: "a running application serving a ready model is callable",
			row: adminModelRow{
				ProviderSource: "olares", ProviderStatus: "active",
				ProviderOlaresStatus: ptr("running"), ModelConsoleStatus: ptr("ready"),
				Model: providerModelRow{Enabled: true, Status: "active"},
			},
			want: "yes",
		},
		{
			name: "a stopped application reports the phase, not a switch",
			row: adminModelRow{
				ProviderSource: "olares", ProviderStatus: "active",
				ProviderOlaresStatus: ptr("stopped"),
				Model:                providerModelRow{Enabled: true, Status: "active"},
			},
			want: "no · app stopped",
		},
		{
			name: "an installing application reports the phase it is in",
			row: adminModelRow{
				ProviderSource: "olares", ProviderStatus: "active",
				ProviderOlaresStatus: ptr("downloading"),
				Model:                providerModelRow{Enabled: true, Status: "active"},
			},
			want: "no · app downloading",
		},
		{
			name: "a failed application reports the failure",
			row: adminModelRow{
				ProviderSource: "olares", ProviderStatus: "active",
				ProviderOlaresStatus: ptr("failed"),
				Model:                providerModelRow{Enabled: true, Status: "active"},
			},
			want: "no · app failed",
		},
		{
			// The row this change exists for: every switch is on and the
			// container is up, and a call would still be refused.
			name: "a running application still fetching weights is not callable",
			row: adminModelRow{
				ProviderSource: "olares", ProviderStatus: "active",
				ProviderOlaresStatus: ptr("running"), ModelConsoleStatus: ptr("download"),
				Model: providerModelRow{Enabled: true, Status: "active"},
			},
			want: "no · fetching weights",
		},
		{
			// Not the same fix as the line above, which is the whole reason
			// the two are spelled apart.
			name: "an engine that has not started says so instead of downloading",
			row: adminModelRow{
				ProviderSource: "olares", ProviderStatus: "active",
				ProviderOlaresStatus: ptr("running"), ModelConsoleStatus: ptr("loading"),
				Model: providerModelRow{Enabled: true, Status: "active"},
			},
			want: "no · engine loading",
		},
		{
			name: "a model that could not load is distinguished from an app that could not install",
			row: adminModelRow{
				ProviderSource: "olares", ProviderStatus: "active",
				ProviderOlaresStatus: ptr("running"), ModelConsoleStatus: ptr("failed"),
				Model: providerModelRow{Enabled: true, Status: "active"},
			},
			want: "no · model load failed",
		},
		{
			// llm-init's health loop moves a model between ready and degraded
			// while the engine usually still answers; Router dispatches to it.
			name: "a degraded model is still callable",
			row: adminModelRow{
				ProviderSource: "olares", ProviderStatus: "active",
				ProviderOlaresStatus: ptr("running"), ModelConsoleStatus: ptr("degraded"),
				Model: providerModelRow{Enabled: true, Status: "active"},
			},
			want: "yes",
		},
		{
			// The vocabulary belongs to the application. Not recognising a
			// word is not evidence against the model behind it.
			name: "a phase this build has never heard of does not hide the model",
			row: adminModelRow{
				ProviderSource: "olares", ProviderStatus: "active",
				ProviderOlaresStatus: ptr("running"), ModelConsoleStatus: ptr("recalibrating"),
				Model: providerModelRow{Enabled: true, Status: "active"},
			},
			want: "yes",
		},
		{
			// An application running its own engine reports no phase at all,
			// permanently. An allow-list would hide it forever.
			name: "a running application that reports no phase is callable",
			row: adminModelRow{
				ProviderSource: "olares", ProviderStatus: "active",
				ProviderOlaresStatus: ptr("running"),
				Model:                providerModelRow{Enabled: true, Status: "active"},
			},
			want: "yes",
		},
		{
			// What an older Router sends, and what an application that has
			// reported nothing leaves behind.
			name: "no container status at all is named rather than guessed",
			row: adminModelRow{
				ProviderSource: "olares", ProviderStatus: "active",
				Model: providerModelRow{Enabled: true, Status: "active"},
			},
			want: "no · app state unknown",
		},
		{
			name: "a model switched off says so even while its application runs",
			row: adminModelRow{
				ProviderSource: "olares", ProviderStatus: "active",
				ProviderOlaresStatus: ptr("running"), ModelConsoleStatus: ptr("ready"),
				Model: providerModelRow{Enabled: false, Status: "active"},
			},
			want: "no · model disabled",
		},
		{
			name: "a model whose own status is not active is disabled",
			row: adminModelRow{
				ProviderSource: "manual", ProviderStatus: "active",
				Model: providerModelRow{Enabled: true, Status: "disabled"},
			},
			want: "no · model disabled",
		},
		{
			// Different fix from the line above: the provider, not the model.
			name: "a disabled provider is named separately",
			row: adminModelRow{
				ProviderSource: "manual", ProviderStatus: "disabled",
				Model: providerModelRow{Enabled: true, Status: "active"},
			},
			want: "no · provider disabled",
		},
		{
			name: "a cloud vendor is callable without any phase",
			row: adminModelRow{
				ProviderSource: "manual", ProviderStatus: "active",
				Model: providerModelRow{Enabled: true, Status: "active"},
			},
			want: "yes",
		},
		{
			// A phase on a manual row would be Router contradicting itself;
			// whatever it means, it does not decide callability here.
			name: "a phase on a manual provider does not hide the model",
			row: adminModelRow{
				ProviderSource: "manual", ProviderStatus: "active",
				ProviderOlaresStatus: ptr("stopped"), ModelConsoleStatus: ptr("download"),
				Model: providerModelRow{Enabled: true, Status: "active"},
			},
			want: "yes",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.row.callableNote(); got != tc.want {
				t.Errorf("callableNote() = %q, want %q", got, tc.want)
			}
			if got := tc.row.callable(); got != (tc.want == "yes") {
				t.Errorf("callable() = %v but the note reads %q", got, tc.want)
			}
		})
	}
}

// Readiness answers a narrower question than callable: only the weights, in the
// words `router call models` prints, so a reader who saw one can recognise the
// other. An admin switch is not a readiness problem.
func TestReadinessDescribesTheWeightsAndNothingElse(t *testing.T) {
	ptr := func(s string) *string { return &s }

	for _, tc := range []struct {
		name string
		row  adminModelRow
		want string
	}{
		{
			name: "a remote vendor has no weights to wait for",
			row: adminModelRow{
				ProviderSource: "manual", ProviderStatus: "active",
				Model: providerModelRow{Enabled: true, Status: "active"},
			},
			want: "ready",
		},
		{
			name: "a disabled model is still ready if its weights are",
			row: adminModelRow{
				ProviderSource: "olares", ProviderStatus: "active",
				ProviderOlaresStatus: ptr("running"), ModelConsoleStatus: ptr("ready"),
				Model: providerModelRow{Enabled: false, Status: "active"},
			},
			want: "ready",
		},
		{
			name: "an application that is not running has nothing to ask",
			row: adminModelRow{
				ProviderSource: "olares", ProviderOlaresStatus: ptr("stopped"), ProviderStatus: "active",
				Model: providerModelRow{Enabled: true, Status: "active"},
			},
			want: "unknown",
		},
		{
			name: "an application reporting no phase cannot be judged",
			row: adminModelRow{
				ProviderSource: "olares", ProviderOlaresStatus: ptr("running"), ProviderStatus: "active",
				Model: providerModelRow{Enabled: true, Status: "active"},
			},
			want: "unknown",
		},
		{
			name: "fetching weights is warming",
			row: adminModelRow{
				ProviderSource: "olares", ProviderStatus: "active",
				ProviderOlaresStatus: ptr("running"), ModelConsoleStatus: ptr("download"),
				Model: providerModelRow{Enabled: true, Status: "active"},
			},
			want: "warming",
		},
		{
			name: "a starting engine is warming",
			row: adminModelRow{
				ProviderSource: "olares", ProviderStatus: "active",
				ProviderOlaresStatus: ptr("running"), ModelConsoleStatus: ptr("loading"),
				Model: providerModelRow{Enabled: true, Status: "active"},
			},
			want: "warming",
		},
		{
			name: "a model that could not load is failed",
			row: adminModelRow{
				ProviderSource: "olares", ProviderStatus: "active",
				ProviderOlaresStatus: ptr("running"), ModelConsoleStatus: ptr("failed"),
				Model: providerModelRow{Enabled: true, Status: "active"},
			},
			want: "failed",
		},
		{
			name: "degraded and unrecognised both read ready",
			row: adminModelRow{
				ProviderSource: "olares", ProviderStatus: "active",
				ProviderOlaresStatus: ptr("running"), ModelConsoleStatus: ptr("degraded"),
				Model: providerModelRow{Enabled: true, Status: "active"},
			},
			want: "ready",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.row.readiness(); got != tc.want {
				t.Errorf("readiness() = %q, want %q", got, tc.want)
			}
		})
	}
}

// Three columns asking the reader to combine them is the reading this table
// removes, so it must not offer them again.
func TestTheModelTableStatesReachabilityOnce(t *testing.T) {
	stopped := "stopped"
	var buf bytes.Buffer
	rows := []adminModelRow{
		{
			ProviderName: "Olares", ProviderSource: "olares", ProviderStatus: "active",
			ProviderOlaresStatus: &stopped,
			Model:                providerModelRow{Name: "qwen3-8b", Mode: "chat", Enabled: true, Status: "active"},
		},
	}
	if err := renderModelList(&buf, rows, 1, 100, 0); err != nil {
		t.Fatalf("render: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "CALLABLE") {
		t.Errorf("table does not state reachability: %s", out)
	}
	for _, gone := range []string{"OFFERED", "PROVIDER STATUS"} {
		if strings.Contains(out, gone) {
			t.Errorf("table still carries the %s column: %s", gone, out)
		}
	}
	if !strings.Contains(out, "app stopped") {
		t.Errorf("a stopped application's model does not say so: %s", out)
	}
}

// A model on its way up is the one case where the answer changes by itself, so
// the table says how to watch it rather than leaving "no" looking terminal.
func TestAWarmingModelIsToldItWillGetThere(t *testing.T) {
	running, phase := "running", "loading"
	var buf bytes.Buffer
	rows := []adminModelRow{
		{
			ProviderName: "Olares", ProviderSource: "olares", ProviderStatus: "active",
			ProviderOlaresStatus: &running, ModelConsoleStatus: &phase,
			Model: providerModelRow{Name: "qwen3-8b", Mode: "chat", Enabled: true, Status: "active"},
		},
	}
	if err := renderModelList(&buf, rows, 1, 100, 0); err != nil {
		t.Fatalf("render: %v", err)
	}
	if out := buf.String(); !strings.Contains(out, "model progress") {
		t.Errorf("a warming model is not told how to follow it: %s", out)
	}
}
