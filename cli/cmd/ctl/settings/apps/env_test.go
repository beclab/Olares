package apps

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

func TestParseVarFlags(t *testing.T) {
	cases := []struct {
		name    string
		in      []string
		want    map[string]string
		wantErr string
	}{
		{
			name: "splits_on_first_equals_only",
			in:   []string{"GREETING=hi=there"},
			want: map[string]string{"GREETING": "hi=there"},
		},
		{
			name: "empty_value_is_allowed",
			in:   []string{"LOG_LEVEL="},
			want: map[string]string{"LOG_LEVEL": ""},
		},
		{
			name: "trims_the_key_but_not_the_value",
			in:   []string{"  API_URL  = https://x "},
			want: map[string]string{"API_URL": " https://x "},
		},
		{
			name: "last_wins_on_a_repeated_key",
			in:   []string{"A=1", "A=2"},
			want: map[string]string{"A": "2"},
		},
		{name: "missing_equals", in: []string{"NOPE"}, wantErr: "expected KEY=VALUE"},
		{name: "empty_key", in: []string{"=1"}, wantErr: "expected KEY=VALUE"},
		{name: "blank_key", in: []string{"   =1"}, wantErr: "empty key"},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseVarFlags(tc.in)
			if tc.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
					t.Fatalf("err = %v; want it to mention %q", err, tc.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(got) != len(tc.want) {
				t.Fatalf("got %v; want %v", got, tc.want)
			}
			for k, v := range tc.want {
				if got[k] != v {
					t.Errorf("key %q = %q; want %q", k, got[k], v)
				}
			}
		})
	}
}

func TestEnvUpdateBodyCarriesOnlyTheRequestedNames(t *testing.T) {
	body := envUpdateBody(map[string]string{"ZED": "3", "ALPHA": "1"})
	if len(body) != 2 {
		t.Fatalf("body has %d entries; want only the 2 requested", len(body))
	}
	if body[0]["envName"] != "ALPHA" || body[1]["envName"] != "ZED" {
		t.Errorf("body is not sorted by name: %v", body)
	}
	if body[0]["value"] != "1" || body[1]["value"] != "3" {
		t.Errorf("values do not match the request: %v", body)
	}
}

func TestSplitAppliedEnvKeys(t *testing.T) {
	// The upstream echoes the app's full vector, so a requested name
	// missing from it was silently dropped rather than created.
	after := []baseEnv{{EnvName: "API_URL"}, {EnvName: "LOG_LEVEL"}}
	applied, ignored := splitAppliedEnvKeys(
		map[string]string{"LOG_LEVEL": "debug", "TYPO": "x", "API_URL": "y"},
		after,
	)
	if strings.Join(applied, ",") != "API_URL,LOG_LEVEL" {
		t.Errorf("applied = %v; want the two declared names, sorted", applied)
	}
	if strings.Join(ignored, ",") != "TYPO" {
		t.Errorf("ignored = %v; want only the undeclared name", ignored)
	}
}

func TestRunAppEnvSetSendsOnlyTheNamedVariables(t *testing.T) {
	d := &fakeDoer{}
	// A read-only variable in the app's vector must not end up in the
	// request: the upstream 400s the whole call if it is named.
	d.enqueueEnvelope([]baseEnv{
		{EnvName: "API_URL"},
		{EnvName: "OLARES_SECRET"},
	})

	if err := runAppEnvSetWithDoer(context.Background(), d, "my-app", map[string]string{"API_URL": "https://x"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(d.calls) != 1 {
		t.Fatalf("made %d wire calls; want a single PUT with no preparatory GET", len(d.calls))
	}
	call := d.lastCall()
	if call.method != "PUT" || call.path != "/api/env/apps/my-app/env" {
		t.Errorf("call = %s %s; want PUT /api/env/apps/my-app/env", call.method, call.path)
	}
	raw, err := json.Marshal(call.body)
	if err != nil {
		t.Fatalf("marshal body: %v", err)
	}
	if got := string(raw); got != `[{"envName":"API_URL","value":"https://x"}]` {
		t.Errorf("body = %s; want only the named variable", got)
	}
}

func TestRunAppEnvSetReportsAnUndeclaredName(t *testing.T) {
	d := &fakeDoer{}
	d.enqueueEnvelope([]baseEnv{{EnvName: "API_URL"}})

	err := runAppEnvSetWithDoer(context.Background(), d, "my-app", map[string]string{"TYPO": "x"})
	if err == nil {
		t.Fatal("setting an undeclared variable reported success")
	}
	for _, sub := range []string{"TYPO", "does not declare"} {
		if !strings.Contains(err.Error(), sub) {
			t.Errorf("error %q does not mention %q", err, sub)
		}
	}
}
