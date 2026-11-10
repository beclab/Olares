package handlers_test

import (
	"testing"

	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/lint/handlerpurity"
)

// TestHandlersPurity enforces CONTRACT.md rules 1-3 on every Go file
// under pkg/applifecycle/handlers/... . It is the CI gate that keeps
// handlers pure without requiring golangci-lint to be installed.
//
// If this test fails, DO NOT skip or add exclusions — the failing
// call needs to be replaced with an Effect (PatchAM / CallWorkload /
// ScheduleTick) returned from the handler, per pkg/applifecycle/
// CONTRACT.md.
//
// The analyzer itself is unit-tested in
// pkg/applifecycle/lint/handlerpurity/analyzer_test.go against
// synthetic fixtures; this test just points it at the real handler
// source tree.
func TestHandlersPurity(t *testing.T) {
	// "." resolves to the directory of this test file's package,
	// which packages.Load walks recursively via "./...". So
	// dropping a subdir like handlers/downloading/ under this
	// package is automatically covered.
	violations, err := handlerpurity.CheckDir(".")
	if err != nil {
		t.Fatalf("handlerpurity.CheckDir failed: %v", err)
	}
	if len(violations) == 0 {
		return
	}
	for _, v := range violations {
		t.Errorf("handler purity violation: %s", v)
	}
}
