package commands

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
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

func TestCommandSummaryRedactsStructuredAndURLCredentials(t *testing.T) {
	for _, input := range []string{
		`Error: denied {"password":"multi word password","access_token":"private-token"}`,
		`Error: Get "https://name:private-password@registry.example/v2/": denied`,
		"Error: password='multi word password' token=private-token",
		`Error: --password "multi word password"`,
		"Authorization: Basic private-token\nError: denied",
	} {
		got := CommandErrorSummary(input)
		for _, secret := range []string{"multi word password", "private-password", "private-token"} {
			if strings.Contains(got, secret) {
				t.Fatalf("credential leaked: %q", got)
			}
		}
		if !strings.Contains(got, "Error:") {
			t.Fatalf("lost diagnostic: %q", got)
		}
	}
}

func TestUpgradeCommandTimeoutAlsoStopsItsChildren(t *testing.T) {
	marker := filepath.Join(t.TempDir(), "orphan-ran")
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	cmd := NewCommandContext(ctx, "/bin/sh", "-c", `(sleep 0.3; printf orphan > "$1") & wait`, "sh", marker)
	if err := cmd.Run(); err == nil {
		t.Fatal("timed-out command succeeded")
	}
	time.Sleep(400 * time.Millisecond)
	if _, err := os.Stat(marker); !os.IsNotExist(err) {
		t.Fatalf("child continued after cancellation: %v", err)
	}
}
