// Package handlerpurity provides a go/analysis pass that enforces
// CONTRACT.md rules 1-3 on every function in pkg/applifecycle/handlers:
//
//   - rule 2: no `go func` inside a handler (spawning goroutines)
//   - rule 3.1: no time.Sleep calls
//   - rule 3.2: no k8s.io/apimachinery/pkg/util/wait.Poll* calls
//   - rule 1 (partial): no obvious controller-runtime client write
//     selectors (Update / Patch / Create / Delete / DeleteAllOf, plus
//     the same on the Status() sub-resource writer)
//
// The analyzer is intentionally conservative: it flags an expression
// only when both the identifier name AND the receiver-package path
// match one of the forbidden signatures. This keeps false positives
// low without needing full SSA data-flow analysis.
//
// Rules 4 and 5 are NOT covered here:
//
//   - rule 4 (undeclared Next) is enforced at runtime by
//     domain.Engine.Step's ErrUndeclaredTransition. A static
//     equivalent would require constant-propagation across the
//     handler body, which yields little marginal value over the
//     runtime check.
//   - rule 5 (CallWorkload must carry OnSuccess/OnFailure) is
//     currently enforced by code review; a future analyzer can add
//     static coverage by walking Effect literals in Result.Effects.
//
// Files named *_test.go are skipped. Tests may sleep or start
// goroutines while observing the Runner; the contract applies to
// handler source.
//
// The analyzer runs in two modes:
//
//  1. As a standalone go/analysis pass driven by golangci-lint or a
//     dedicated cmd/lint binary — this is the CI mode.
//  2. Inline from pkg/applifecycle/handlers/purity_test.go, which
//     loads its own package with go/packages and runs the analyzer
//     — this makes `go test ./pkg/applifecycle/...` a hard gate so
//     no CI setup is required to catch violations locally.
package handlerpurity
