package apps

import (
	"context"
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func TestParseSubPolicySpec(t *testing.T) {
	cases := []struct {
		name    string
		spec    string
		want    EntrancePolicy
		wantErr string
	}{
		{
			name: "minimal",
			spec: "uri=/admin,policy=two_factor",
			want: EntrancePolicy{URI: "/admin", Policy: "two_factor"},
		},
		{
			name: "full",
			spec: "uri=/admin,policy=one_factor,one_time=true,valid_duration=600",
			want: EntrancePolicy{URI: "/admin", Policy: "one_factor", OneTime: true, ValidDuration: 600},
		},
		{
			name: "tolerates whitespace",
			spec: "  uri = /admin , policy = public ",
			want: EntrancePolicy{URI: "/admin", Policy: "public"},
		},
		{
			name: "rejects missing uri",
			spec: "policy=two_factor",
			wantErr: "uri=",
		},
		{
			name: "rejects missing policy",
			spec: "uri=/admin",
			wantErr: "policy=",
		},
		{
			name: "rejects unknown policy value",
			spec: "uri=/admin,policy=bogus",
			wantErr: "system|one_factor|two_factor|public",
		},
		{
			name: "rejects unknown key",
			spec: "uri=/admin,policy=public,frobnicate=true",
			wantErr: "unknown key",
		},
		{
			name: "rejects malformed kv pair",
			spec: "uri=/admin,policy",
			wantErr: "expected key=value",
		},
		{
			name: "rejects bad one_time value",
			spec: "uri=/x,policy=public,one_time=truthy",
			wantErr: "one_time",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := parseSubPolicySpec(c.spec)
			if c.wantErr != "" {
				if err == nil {
					t.Fatalf("want err containing %q, got nil (got=%+v)", c.wantErr, got)
				}
				if !strings.Contains(err.Error(), c.wantErr) {
					t.Fatalf("err=%q, want substring %q", err.Error(), c.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected err: %v", err)
			}
			if !reflect.DeepEqual(got, c.want) {
				t.Fatalf("got=%+v want=%+v", got, c.want)
			}
		})
	}
}

func TestParseSubPolicySpecs(t *testing.T) {
	got, err := parseSubPolicySpecs([]string{
		"uri=/admin,policy=two_factor",
		"uri=/healthz,policy=public",
	})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	want := []EntrancePolicy{
		{URI: "/admin", Policy: "two_factor"},
		{URI: "/healthz", Policy: "public"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got=%+v want=%+v", got, want)
	}

	// Index propagates through the error message so users can locate the
	// bad spec in their flag list.
	_, err = parseSubPolicySpecs([]string{"uri=/ok,policy=public", "garbage"})
	if err == nil {
		t.Fatal("want err for second spec, got nil")
	}
	if !strings.Contains(err.Error(), "[1]") {
		t.Fatalf("err=%q does not include the index", err.Error())
	}
}

func TestSummarizeSubPolicies(t *testing.T) {
	if got := summarizeSubPolicies(nil); got != "null" {
		t.Errorf("nil pointer should render as null, got %q", got)
	}
	empty := []EntrancePolicy{}
	if got := summarizeSubPolicies(&empty); got != "[]" {
		t.Errorf("empty slice should render as [], got %q", got)
	}
	some := []EntrancePolicy{
		{URI: "/admin", Policy: "two_factor"},
		{URI: "/healthz", Policy: "public"},
	}
	got := summarizeSubPolicies(&some)
	want := "[/admin=two_factor, /healthz=public]"
	if got != want {
		t.Errorf("got=%q want=%q", got, want)
	}
}

func TestSubPoliciesPtr_PreservesEmpty(t *testing.T) {
	// Nil input must NOT roundtrip back as a nil pointer — that would
	// mean we send sub_policies: null in the wire body when the user
	// hasn't asked to clear them. Subtlety matters here because the
	// SPA uses null = "drop all" and [] = "no overrides yet, please
	// preserve" interchangeably; we err on the side of explicit-empty.
	p := subPoliciesPtr(nil)
	if p == nil {
		t.Fatal("nil input should produce a non-nil pointer to []")
	}
	if len(*p) != 0 {
		t.Fatalf("want empty slice, got %v", *p)
	}
}

func TestRunPolicySet_ReadModifyWritePreservesUntouchedFields(t *testing.T) {
	// GET response: existing config the user is NOT asking to change
	// for one_time / sub_policies — only --default-policy is passed.
	doer := &fakeDoer{}
	doer.enqueueEnvelope(SetupPolicy{
		DefaultPolicy: "one_factor",
		OneTime:       true,
		ValidDuration: 300,
		SubPolicies: []EntrancePolicy{
			{URI: "/admin", Policy: "two_factor", OneTime: true, ValidDuration: 600},
		},
	})
	doer.enqueueEmptyEnvelope() // POST response

	flags := policySetFlags{
		defaultPolicy:    "two_factor",
		defaultPolicySet: true,
	}
	if err := runPolicySetWithDoer(context.Background(), doer, "files", "file", flags); err != nil {
		t.Fatalf("runPolicySetWithDoer: %v", err)
	}

	if len(doer.calls) != 2 {
		t.Fatalf("want 2 calls (GET + POST), got %d", len(doer.calls))
	}
	if doer.calls[0].method != "GET" || doer.calls[1].method != "POST" {
		t.Fatalf("call sequence wrong: %+v", doer.calls)
	}

	post := doer.calls[1].body.(setupPolicyBody)
	if post.DefaultPolicy != "two_factor" {
		t.Errorf("DefaultPolicy not applied: got %q", post.DefaultPolicy)
	}
	if !post.OneTime {
		t.Errorf("OneTime should be preserved as true (RMW), got false")
	}
	if post.ValidDuration != 300 {
		t.Errorf("ValidDuration should be preserved as 300 (RMW), got %d", post.ValidDuration)
	}
	if post.SubPolicies == nil {
		t.Fatal("SubPolicies should not be nil — it was preserved from current")
	}
	if len(*post.SubPolicies) != 1 || (*post.SubPolicies)[0].URI != "/admin" {
		t.Errorf("SubPolicies not preserved, got %+v", *post.SubPolicies)
	}
}

func TestRunPolicySet_ClearSubPoliciesPostsNull(t *testing.T) {
	doer := &fakeDoer{}
	doer.enqueueEnvelope(SetupPolicy{
		DefaultPolicy: "public",
		SubPolicies: []EntrancePolicy{
			{URI: "/admin", Policy: "two_factor"},
		},
	})
	doer.enqueueEmptyEnvelope()

	flags := policySetFlags{clearSubPolicies: true}
	if err := runPolicySetWithDoer(context.Background(), doer, "files", "file", flags); err != nil {
		t.Fatalf("runPolicySetWithDoer: %v", err)
	}
	post := doer.calls[1].body.(setupPolicyBody)
	if post.SubPolicies != nil {
		t.Errorf("clear-sub-policies must set SubPolicies pointer to nil so JSON marshals as null, got %+v", post.SubPolicies)
	}
	// Verify the JSON marshal actually emits null for sub_policies so
	// the upstream sees the SPA's "drop all" signal.
	raw, err := json.Marshal(post)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if !strings.Contains(string(raw), `"sub_policies":null`) {
		t.Errorf("expected sub_policies:null in body, got %s", raw)
	}
}

func TestRunPolicySet_RejectsConflictingFlags(t *testing.T) {
	cases := []struct {
		name  string
		flags policySetFlags
	}{
		{
			name: "sub-policy and sub-policies-file are mutex",
			flags: policySetFlags{
				subPolicySpecs:  []string{"uri=/x,policy=public"},
				subPolicySet:    true,
				subPoliciesFile: "/dev/null",
			},
		},
		{
			name: "clear and sub-policy are mutex",
			flags: policySetFlags{
				clearSubPolicies: true,
				subPolicySpecs:   []string{"uri=/x,policy=public"},
				subPolicySet:     true,
			},
		},
		{
			name: "default-policy must be valid",
			flags: policySetFlags{
				defaultPolicy:    "bogus",
				defaultPolicySet: true,
			},
		},
		{
			name:  "must change at least one knob",
			flags: policySetFlags{},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			doer := &fakeDoer{}
			err := runPolicySetWithDoer(context.Background(), doer, "files", "file", c.flags)
			if err == nil {
				t.Fatal("want validation err, got nil")
			}
			if len(doer.calls) != 0 {
				t.Errorf("validation should reject before any wire call; got calls=%+v", doer.calls)
			}
		})
	}
}
