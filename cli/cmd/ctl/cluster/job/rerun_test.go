package job

import "testing"

// TestJobTerminalState pins the predicate that decides whether
// `cluster job rerun` should refuse a Job. We treat ANY condition
// with status=True named "Complete" or "Failed" (case-insensitive,
// matching K8s's ConditionType convention) as terminal — the Job
// controller stops scheduling once it sets either, so a client-side
// "delete pods" rerun would be a no-op.
//
// The reason string is propagated for Failed conditions so the user
// sees "Failed: BackoffLimitExceeded" rather than a bare "Failed".
// For Complete we deliberately drop the reason (it's almost always
// blank in practice and adds noise to the error message).
func TestJobTerminalState(t *testing.T) {
	tests := []struct {
		name       string
		conditions []JobCondition
		wantTerm   bool
		wantReason string
	}{
		{
			name:       "no conditions: not terminal",
			conditions: nil,
			wantTerm:   false,
		},
		{
			name: "Complete=True is terminal",
			conditions: []JobCondition{
				{Type: "Complete", Status: "True"},
			},
			wantTerm:   true,
			wantReason: "Complete",
		},
		{
			name: "Failed=True with reason is terminal, reason propagated",
			conditions: []JobCondition{
				{Type: "Failed", Status: "True", Reason: "BackoffLimitExceeded"},
			},
			wantTerm:   true,
			wantReason: "Failed: BackoffLimitExceeded",
		},
		{
			name: "Failed=True without reason renders bare 'Failed'",
			conditions: []JobCondition{
				{Type: "Failed", Status: "True"},
			},
			wantTerm:   true,
			wantReason: "Failed",
		},
		{
			// Suspended Jobs aren't terminal — the controller resumes
			// scheduling when spec.suspend flips back to false, so a
			// rerun could meaningfully delete in-flight pods. We
			// intentionally don't gate on Suspended.
			name: "Suspended=True is NOT terminal",
			conditions: []JobCondition{
				{Type: "Suspended", Status: "True"},
			},
			wantTerm: false,
		},
		{
			// Status=False / Unknown means the condition doesn't
			// apply (K8s convention) — don't treat a "Complete with
			// status=False" Job as terminal just because the type
			// name matches.
			name: "Complete=False is NOT terminal",
			conditions: []JobCondition{
				{Type: "Complete", Status: "False"},
			},
			wantTerm: false,
		},
		{
			// Real-world Jobs carry multiple conditions; the
			// predicate must find a terminal one regardless of
			// position in the slice.
			name: "terminal condition mixed with non-terminal ones",
			conditions: []JobCondition{
				{Type: "Suspended", Status: "False"},
				{Type: "Complete", Status: "True"},
			},
			wantTerm:   true,
			wantReason: "Complete",
		},
		{
			// K8s ConditionType strings are PascalCase, but be lenient
			// in case the apiserver flavor changes — `strings.ToLower`
			// keeps the predicate robust.
			name: "lower-case type name still matches",
			conditions: []JobCondition{
				{Type: "complete", Status: "True"},
			},
			wantTerm:   true,
			wantReason: "Complete",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			j := Job{Status: JobStatus{Conditions: tc.conditions}}
			gotTerm, gotReason := jobTerminalState(j)
			if gotTerm != tc.wantTerm {
				t.Fatalf("terminal: got %v, want %v (reason=%q)", gotTerm, tc.wantTerm, gotReason)
			}
			if !tc.wantTerm {
				return
			}
			if gotReason != tc.wantReason {
				t.Fatalf("reason: got %q, want %q", gotReason, tc.wantReason)
			}
		})
	}
}

