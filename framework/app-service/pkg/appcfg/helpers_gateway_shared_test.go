package appcfg

import "testing"

func TestSharedEntranceID(t *testing.T) {
	if got, _ := SharedEntranceID("ab12", 0, 1); got != "ab12" {
		t.Errorf("count=1 got %q, want ab12", got)
	}
	if got, _ := SharedEntranceID("ab12", 1, 3); got != "ab121" {
		t.Errorf("count=3 idx=1 got %q, want ab121", got)
	}
	if _, err := SharedEntranceID("", 0, 1); err == nil {
		t.Error("empty appid should error")
	}
	if _, err := SharedEntranceID("ab12", 5, 3); err == nil {
		t.Error("index out of range should error")
	}
}

func TestLogicalHostPattern(t *testing.T) {
	got, err := LogicalHostPattern("ab12", 0, 1, "Olares.COM.")
	if err != nil {
		t.Fatalf("LogicalHostPattern: %v", err)
	}
	if got != "ab12.*.olares.com" {
		t.Errorf("got %q, want ab12.*.olares.com", got)
	}
	if _, err := LogicalHostPattern("ab12", 0, 1, ""); err == nil {
		t.Error("empty platformDomain should error")
	}
}
