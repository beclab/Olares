// Command handlerpurity is a standalone runner for the
// pkg/applifecycle/lint/handlerpurity analyzer. It exists so CI and
// local `make handler-contract-check` invocations do not require
// golangci-lint to be installed — this binary is buildable from the
// module's own go.sum with no extra tooling.
//
// Typical usage:
//
//	go run ./cmd/handlerpurity ./pkg/applifecycle/handlers/...
//
// or, once built into ./bin:
//
//	./bin/handlerpurity ./pkg/applifecycle/handlers/...
//
// The handlerpurity analyzer runs on every function in every package
// passed to it, except files named *_test.go. It is *not*
// self-scoping — callers are expected to point it at
// pkg/applifecycle/handlers/... (and any future handler subtree)
// rather than the whole repo.
//
// Rule coverage (see pkg/applifecycle/CONTRACT.md):
//
//   - rule 2: no `go` statements
//   - rule 3.1: no time.Sleep
//   - rule 3.2: no wait.Poll / PollImmediate / PollUntil /
//     PollImmediateUntil / PollUntilContextCancel / PollUntilContextTimeout
//   - rule 1 (partial): no client.Update / Patch / Create / Delete /
//     DeleteAllOf on receivers imported from
//     sigs.k8s.io/controller-runtime/pkg/client
//
// Rules 4 and 5 are enforced elsewhere: rule 4 by domain.Engine.Step
// at runtime (ErrUndeclaredTransition), rule 5 by code review + the
// fact that Action is a closed sum type so unresolved callback
// identifiers fail to compile.
package main

import (
	"golang.org/x/tools/go/analysis/singlechecker"

	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/lint/handlerpurity"
)

func main() {
	singlechecker.Main(handlerpurity.Analyzer)
}
