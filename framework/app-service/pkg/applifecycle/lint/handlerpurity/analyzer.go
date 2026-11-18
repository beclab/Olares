package handlerpurity

import (
	"go/ast"
	"go/types"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// Analyzer is the exported go/analysis.Analyzer that enforces
// CONTRACT.md rules 1-3 on handler functions. See the package doc for
// exactly which forbidden constructs are detected.
var Analyzer = &analysis.Analyzer{
	Name: "handlerpurity",
	Doc: "enforce CONTRACT.md rules 1-3: no goroutines, no blocking " +
		"waits, no direct cluster mutations inside HandlerFunc bodies " +
		"(pkg/applifecycle/handlers)",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

// forbiddenCall is a single forbidden (package-suffix, identifier) pair.
// When a package-qualified call expression matches one of these, the
// analyzer reports it with the message the caller sees.
//
// pkgSuffix uses last-segment matching so the same rule applies to both
// the real K8s package (e.g. "k8s.io/apimachinery/pkg/util/wait") and
// any stub package used in analyzer tests (e.g. "poll/wait"). See
// matchesPkgSuffix.
type forbiddenCall struct {
	// pkgSuffix is the trailing segment of the callee package's
	// import path (e.g. "time" or "wait"). Matching is:
	//   path == pkgSuffix || strings.HasSuffix(path, "/"+pkgSuffix)
	pkgSuffix string

	// name is the identifier of the callee (function or method name).
	name string

	// message is what the analyzer prints when the call is found.
	message string
}

// blockingCalls covers CONTRACT.md rule 3: no time.Sleep / wait.Poll*.
var blockingCalls = []forbiddenCall{
	{
		pkgSuffix: "time",
		name:      "Sleep",
		message:   "handler must not call time.Sleep (CONTRACT.md rule 3): return a ScheduleTick effect instead",
	},
	{
		pkgSuffix: "wait",
		name:      "Poll",
		message:   "handler must not call wait.Poll (CONTRACT.md rule 3): return a ScheduleTick effect instead",
	},
	{
		pkgSuffix: "wait",
		name:      "PollImmediate",
		message:   "handler must not call wait.PollImmediate (CONTRACT.md rule 3): return a ScheduleTick effect instead",
	},
	{
		pkgSuffix: "wait",
		name:      "PollUntil",
		message:   "handler must not call wait.PollUntil (CONTRACT.md rule 3): return a ScheduleTick effect instead",
	},
	{
		pkgSuffix: "wait",
		name:      "PollImmediateUntil",
		message:   "handler must not call wait.PollImmediateUntil (CONTRACT.md rule 3): return a ScheduleTick effect instead",
	},
	{
		pkgSuffix: "wait",
		name:      "PollUntilContextCancel",
		message:   "handler must not call wait.PollUntilContextCancel (CONTRACT.md rule 3): return a ScheduleTick effect instead",
	},
	{
		pkgSuffix: "wait",
		name:      "PollUntilContextTimeout",
		message:   "handler must not call wait.PollUntilContextTimeout (CONTRACT.md rule 3): return a ScheduleTick effect instead",
	},
}

// clientWriteMethods covers CONTRACT.md rule 1: no direct client
// mutations. The set matches sigs.k8s.io/controller-runtime's
// client.Client (and its SubResourceWriter) interface method names.
// Because method-name uniqueness is not perfect across packages, the
// analyzer additionally checks the receiver's package path is one of
// clientReceiverPackages before reporting.
var clientWriteMethods = map[string]string{
	"Update":      "handler must not call client.Update (CONTRACT.md rule 1): return a PatchAM effect instead",
	"Patch":       "handler must not call client.Patch (CONTRACT.md rule 1): return a PatchAM effect instead",
	"Create":      "handler must not call client.Create (CONTRACT.md rule 1): return a CallWorkload effect (or a PatchAM if the target is the AM itself)",
	"Delete":      "handler must not call client.Delete (CONTRACT.md rule 1): return a CallWorkload effect",
	"DeleteAllOf": "handler must not call client.DeleteAllOf (CONTRACT.md rule 1): return a CallWorkload effect",
}

// clientReceiverPackages restricts clientWriteMethods matching to
// receivers imported from controller-runtime's client package (which
// covers both Client and SubResourceWriter, since Status() returns the
// latter). Anything else with a matching method name is not flagged.
var clientReceiverPackages = []string{
	"sigs.k8s.io/controller-runtime/pkg/client",
}

// run implements analysis.Analyzer.Run.
func run(pass *analysis.Pass) (any, error) {
	insp := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	filter := []ast.Node{
		(*ast.GoStmt)(nil),
		(*ast.CallExpr)(nil),
	}

	insp.Preorder(filter, func(n ast.Node) {
		// Tests sleep and start goroutines to observe the Runner.
		// The contract applies to handler source, not *_test.go.
		if isTestFile(pass.Fset.Position(n.Pos()).Filename) {
			return
		}
		switch node := n.(type) {
		case *ast.GoStmt:
			// CONTRACT.md rule 2: no goroutine spawning.
			pass.Report(analysis.Diagnostic{
				Pos:     node.Pos(),
				Message: goStmtMessage,
			})
		case *ast.CallExpr:
			checkForbiddenCall(pass, node)
		}
	})

	return nil, nil
}

// checkForbiddenCall reports node when it matches an entry in
// blockingCalls or, for methods whose receiver is imported from
// controller-runtime's client package, an entry in clientWriteMethods.
func checkForbiddenCall(pass *analysis.Pass, call *ast.CallExpr) {
	if msg := findForbiddenCall(pass.TypesInfo, call); msg != "" {
		pass.Report(analysis.Diagnostic{
			Pos:     call.Pos(),
			Message: msg,
		})
	}
}

// findForbiddenCall returns the diagnostic message for call when it
// matches a forbidden pattern, or "" otherwise. Extracted so both
// Analyzer.Run (via checkForbiddenCall) and CheckDir (see checker.go)
// can share the same matching logic.
func findForbiddenCall(info *types.Info, call *ast.CallExpr) string {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return ""
	}
	name := sel.Sel.Name

	// Blocking calls first — these are package-scoped functions
	// (time.Sleep, wait.Poll*) so the qualifier resolves through
	// pkgPathOfSelector.
	calleePath := pkgPathOfSelector(info, sel)
	for _, fc := range blockingCalls {
		if fc.name != name {
			continue
		}
		if matchesPkgSuffix(calleePath, fc.pkgSuffix) {
			return fc.message
		}
	}

	// Client write methods — Selections carries the receiver type
	// for genuine method calls; only match when the receiver's
	// declaring package is in clientReceiverPackages, to keep
	// false positives (e.g. someType.Update on unrelated types)
	// from firing.
	if msg, ok := clientWriteMethods[name]; ok && info != nil {
		if sels := info.Selections[sel]; sels != nil {
			recv := sels.Recv()
			if recv != nil && isClientReceiver(recv) {
				return msg
			}
		}
	}

	return ""
}

// isTestFile reports whether filename is a Go test file. The check
// is the path suffix, so a directory named "_test.go" is not required
// to be special-cased.
func isTestFile(filename string) bool {
	return strings.HasSuffix(filename, "_test.go")
}

// matchesPkgSuffix reports whether path is exactly suffix or ends with
// "/"+suffix. Used to match callee package paths against the trailing
// segment (e.g. "wait" matches both "wait" and any "…/wait").
func matchesPkgSuffix(path, suffix string) bool {
	if suffix == "" {
		return false
	}
	if path == suffix {
		return true
	}
	return strings.HasSuffix(path, "/"+suffix)
}

// pkgPathOfSelector returns the import path of the package the
// selector's qualifier resolves to, or "" when it is not a package
// qualifier (e.g. a method call whose receiver is a local variable).
func pkgPathOfSelector(info *types.Info, sel *ast.SelectorExpr) string {
	ident, ok := sel.X.(*ast.Ident)
	if !ok {
		return ""
	}
	if info == nil {
		return ""
	}
	obj := info.Uses[ident]
	if obj == nil {
		return ""
	}
	pkgName, ok := obj.(*types.PkgName)
	if !ok {
		return ""
	}
	return pkgName.Imported().Path()
}

// isClientReceiver reports whether t is a value whose declared type
// (unwrapping pointers and named types) originates from one of the
// controller-runtime client packages listed in clientReceiverPackages.
func isClientReceiver(t types.Type) bool {
	for {
		switch tt := t.(type) {
		case *types.Pointer:
			t = tt.Elem()
			continue
		case *types.Named:
			pkg := tt.Obj().Pkg()
			if pkg == nil {
				return false
			}
			for _, allowed := range clientReceiverPackages {
				if pkg.Path() == allowed ||
					strings.HasPrefix(pkg.Path(), allowed+"/") {
					return true
				}
			}
			return false
		default:
			return false
		}
	}
}
