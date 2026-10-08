package commands

import (
	"context"
	"strings"
	"testing"
)

func TestRunAsyncWithResultReturnsCLIError(t *testing.T) {
	cmd := NewBaseCommand()
	result, err := cmd.RunAsyncWithResult_(context.Background(), "/bin/sh", "-c",
		"printf 'Error: failed to reach registry, token=private-value\\n' >&2; exit 7")
	if err != nil {
		t.Fatal(err)
	}
	got := <-result
	if got == nil || !strings.Contains(got.Error(), "exit status 7") || !strings.Contains(got.Error(), "failed to reach registry") {
		t.Fatalf("missing exit status or CLI detail: %v", got)
	}
	if strings.Contains(got.Error(), "private-value") || !strings.Contains(got.Error(), "token=[redacted]") {
		t.Fatalf("CLI detail was not redacted: %v", got)
	}
}

func TestRunAsyncWithResultSucceeds(t *testing.T) {
	cmd := NewBaseCommand()
	result, err := cmd.RunAsyncWithResult_(context.Background(), "/bin/sh", "-c", "exit 0")
	if err != nil {
		t.Fatal(err)
	}
	if got := <-result; got != nil {
		t.Fatalf("successful command returned an error: %v", got)
	}
}

func TestCommandErrorTailIsBounded(t *testing.T) {
	var tail commandErrorTail
	_, _ = tail.Write([]byte(strings.Repeat("x", commandErrorTailLimit) + "\nError: timeout"))
	if got := tail.summary(); len(got) > 1024 || !strings.HasSuffix(got, "Error: timeout") {
		t.Fatalf("unexpected bounded summary: %q", got)
	}
}

func TestCommandErrorTailRedactsAuthorizationAndCLIFlags(t *testing.T) {
	var tail commandErrorTail
	_, _ = tail.Write([]byte("Authorization: Bearer private-token\n--password private-password\nError: denied"))
	got := tail.summary()
	if strings.Contains(got, "private-token") || strings.Contains(got, "private-password") || !strings.Contains(got, "Error: denied") {
		t.Fatalf("unsafe or incomplete summary: %q", got)
	}
}
