package router

import (
	"bytes"
	"strings"
	"testing"
)

// A locally installed application owns a model row from the moment it is
// installed, so `router list` carries models of applications that are stopped,
// downloading or failed. It used to print those three switches green — offered,
// active, provider active — which is every word the reader needed to conclude
// the opposite of the truth.
func TestStateSeparatesAStoppedApplicationFromADisabledModel(t *testing.T) {
	phase := func(s string) *string { return &s }

	for _, tc := range []struct {
		name string
		row  adminModelRow
		want string
	}{
		{
			name: "a running application is callable",
			row: adminModelRow{
				ProviderSource: "olares", ProviderStatus: "active",
				ProviderOlaresStatus: phase("running"),
				Model:                providerModelRow{Enabled: true, Status: "active"},
			},
			want: "callable",
		},
		{
			// The row this whole change exists for.
			name: "a stopped application reports the phase, not a switch",
			row: adminModelRow{
				ProviderSource: "olares", ProviderStatus: "active",
				ProviderOlaresStatus: phase("stopped"),
				Model:                providerModelRow{Enabled: true, Status: "active"},
			},
			want: "stopped",
		},
		{
			name: "an installing application reports the phase it is in",
			row: adminModelRow{
				ProviderSource: "olares", ProviderStatus: "active",
				ProviderOlaresStatus: phase("downloading"),
				Model:                providerModelRow{Enabled: true, Status: "active"},
			},
			want: "downloading",
		},
		{
			name: "a failed application reports the failure",
			row: adminModelRow{
				ProviderSource: "olares", ProviderStatus: "active",
				ProviderOlaresStatus: phase("failed"),
				Model:                providerModelRow{Enabled: true, Status: "active"},
			},
			want: "failed",
		},
		{
			// What an older Router sends, and what an application that has
			// reported nothing leaves behind. Silence is not a phase to name.
			name: "no phase at all keeps the generic verdict",
			row: adminModelRow{
				ProviderSource: "olares", ProviderStatus: "active",
				Model: providerModelRow{Enabled: true, Status: "active"},
			},
			want: "disabled",
		},
		{
			name: "a model switched off says so even while its application runs",
			row: adminModelRow{
				ProviderSource: "olares", ProviderStatus: "active",
				ProviderOlaresStatus: phase("running"),
				Model:                providerModelRow{Enabled: false, Status: "active"},
			},
			want: "disabled",
		},
		{
			name: "a model whose own status is not active is disabled",
			row: adminModelRow{
				ProviderSource: "manual", ProviderStatus: "active",
				Model: providerModelRow{Enabled: true, Status: "disabled"},
			},
			want: "disabled",
		},
		{
			// Different fix from the line above: the provider, not the model.
			name: "a disabled provider is named separately",
			row: adminModelRow{
				ProviderSource: "manual", ProviderStatus: "disabled",
				Model: providerModelRow{Enabled: true, Status: "active"},
			},
			want: "provider disabled",
		},
		{
			name: "a cloud vendor is callable without any phase",
			row: adminModelRow{
				ProviderSource: "manual", ProviderStatus: "active",
				Model: providerModelRow{Enabled: true, Status: "active"},
			},
			want: "callable",
		},
		{
			// A phase on a manual row would be Router contradicting itself;
			// whatever it means, it does not decide callability here.
			name: "a phase on a manual provider does not hide the model",
			row: adminModelRow{
				ProviderSource: "manual", ProviderStatus: "active",
				ProviderOlaresStatus: phase("stopped"),
				Model:                providerModelRow{Enabled: true, Status: "active"},
			},
			want: "callable",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.row.state(); got != tc.want {
				t.Errorf("state() = %q, want %q", got, tc.want)
			}
		})
	}
}

// Three columns asking the reader to combine them is the reading this change
// removes, so the table must not offer them again.
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
	if !strings.Contains(out, "STATE") {
		t.Errorf("table does not state reachability: %s", out)
	}
	for _, gone := range []string{"OFFERED", "PROVIDER STATUS"} {
		if strings.Contains(out, gone) {
			t.Errorf("table still carries the %s column: %s", gone, out)
		}
	}
	if !strings.Contains(out, "stopped") {
		t.Errorf("a stopped application's model does not say so: %s", out)
	}
}
