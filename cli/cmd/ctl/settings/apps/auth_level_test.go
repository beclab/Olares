package apps

import (
	"context"
	"testing"
)

func TestRunAuthLevelSet_RejectsBadLevel(t *testing.T) {
	cases := []string{
		"",
		"private-but-not-really",
		"none",
		"admin",
	}
	for _, level := range cases {
		t.Run("level="+level, func(t *testing.T) {
			err := runAuthLevelSetWithDoer(context.Background(), &fakeDoer{}, "files", "file", level)
			if err == nil {
				t.Fatalf("want validation err for level=%q, got nil", level)
			}
		})
	}
}

func TestRunAuthLevelSet_HappyPathBodyShape(t *testing.T) {
	cases := []string{"private", "public", "internal"}
	for _, level := range cases {
		t.Run(level, func(t *testing.T) {
			doer := &fakeDoer{}
			doer.enqueueEmptyEnvelope()
			if err := runAuthLevelSetWithDoer(context.Background(), doer, "files", "file", level); err != nil {
				t.Fatalf("unexpected err: %v", err)
			}
			if len(doer.calls) != 1 {
				t.Fatalf("want 1 call, got %d", len(doer.calls))
			}
			c := doer.calls[0]
			if c.method != "POST" {
				t.Errorf("method=%q want POST", c.method)
			}
			wantPath := "/api/applications/files/file/setup/auth-level"
			if c.path != wantPath {
				t.Errorf("path=%q want %q", c.path, wantPath)
			}
			body := c.body.(map[string]string)
			if body["authorization_level"] != level {
				t.Errorf("body authorization_level=%q want %q", body["authorization_level"], level)
			}
		})
	}
}

func TestRunAuthLevelSet_NormalizesCase(t *testing.T) {
	doer := &fakeDoer{}
	doer.enqueueEmptyEnvelope()
	// Trailing whitespace + uppercase should be normalized before
	// validation; otherwise users get confused error messages when
	// they paste from a doc that mentions "Private".
	if err := runAuthLevelSetWithDoer(context.Background(), doer, "files", "file", " Private "); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	body := doer.calls[0].body.(map[string]string)
	if body["authorization_level"] != "private" {
		t.Errorf("got %q, want lowercase", body["authorization_level"])
	}
}

func TestRunAuthLevelSet_RejectsEmptyArgs(t *testing.T) {
	if err := runAuthLevelSetWithDoer(context.Background(), &fakeDoer{}, "", "file", "private"); err == nil {
		t.Error("want err for empty app")
	}
	if err := runAuthLevelSetWithDoer(context.Background(), &fakeDoer{}, "files", "", "private"); err == nil {
		t.Error("want err for empty entrance")
	}
}

