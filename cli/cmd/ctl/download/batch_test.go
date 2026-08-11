package download

import (
	"bytes"
	"strings"
	"testing"
)

func TestParseTaskIDs(t *testing.T) {
	ids, err := parseTaskIDs([]string{"1", "2", "2", "3"})
	if err != nil {
		t.Fatal(err)
	}
	if len(ids) != 3 || ids[0] != 1 || ids[1] != 2 || ids[2] != 3 {
		t.Fatalf("got %v", ids)
	}
	if _, err := parseTaskIDs(nil); err == nil {
		t.Fatal("expected empty error")
	}
	if _, err := parseTaskIDs([]string{"0"}); err == nil {
		t.Fatal("expected invalid id")
	}
	tooMany := make([]string, batchMaxTasks+1)
	for i := range tooMany {
		tooMany[i] = "1"
	}
	if _, err := parseTaskIDs(tooMany); err == nil || !strings.Contains(err.Error(), "too many") {
		t.Fatalf("expected too many, got %v", err)
	}
}

func TestValidateHFDest(t *testing.T) {
	for _, raw := range []string{"", "cache", "CACHE", " local "} {
		if _, err := validateHFDest(raw); err != nil {
			t.Fatalf("%q: %v", raw, err)
		}
	}
	if _, err := validateHFDest("disk"); err == nil || !strings.Contains(err.Error(), "unsupported --hf-dest") {
		t.Fatalf("expected unsupported, got %v", err)
	}
}

func TestInspectHFTokenHint(t *testing.T) {
	if got := inspectHFTokenHint(InspectData{ErrorCategory: "gated"}); got == "" || !strings.Contains(got, `--extra '{"token":"hf_xxx"}'`) {
		t.Fatalf("gated hint: %q", got)
	}
	if got := inspectHFTokenHint(InspectData{ErrorCategory: "PRIVATE"}); !strings.Contains(got, "token") {
		t.Fatalf("private hint: %q", got)
	}
	if got := inspectHFTokenHint(InspectData{ErrorCategory: "timeout"}); got != "" {
		t.Fatalf("unexpected hint: %q", got)
	}
}

func TestRenderBatchResult(t *testing.T) {
	var buf bytes.Buffer
	err := renderBatchResult(&buf, FormatTable, "pause", BatchResult{
		Succeeded: []int64{1, 2},
		Failed:    []BatchItemError{{TaskID: 3, Error: "not found"}},
	})
	if err == nil || !strings.Contains(err.Error(), "1 of 3") {
		t.Fatalf("expected partial failure, got %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "2 succeeded, 1 failed") || !strings.Contains(out, "3: not found") {
		t.Fatalf("table output: %q", out)
	}

	buf.Reset()
	if err := renderBatchResult(&buf, FormatTable, "pause", BatchResult{Succeeded: []int64{1}}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "1 succeeded, 0 failed") {
		t.Fatalf("got %q", buf.String())
	}
}
