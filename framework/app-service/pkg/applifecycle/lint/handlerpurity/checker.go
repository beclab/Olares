package handlerpurity

import (
	"fmt"
	"go/ast"
	"go/token"

	"golang.org/x/tools/go/packages"
)

// Violation is a single CONTRACT.md rule violation surfaced by
// CheckDir. It carries the resolved source position and the same
// message string the Analyzer would emit through analysis.Diagnostic,
// so callers can format either as a test failure or a lint report.
type Violation struct {
	Pos     token.Position
	Message string
}

// String renders v like a compiler diagnostic ("file:line:col: msg").
func (v Violation) String() string {
	return fmt.Sprintf("%s: %s", v.Pos, v.Message)
}

// CheckDir loads every Go package rooted at dir (recursively via
// "./...") with full type information, then reports every CONTRACT.md
// rule 1-3 violation it finds outside *_test.go. It reuses the exact
// matching logic the go/analysis Analyzer uses (findForbiddenCall +
// the *ast.GoStmt rule), so unit-tested behaviour and integration
// behaviour cannot drift apart.
//
// This exists so a plain `go test ./pkg/applifecycle/handlers/...`
// enforces the handler purity contract without requiring
// golangci-lint or a custom cmd/lint binary. See
// pkg/applifecycle/handlers/purity_test.go for the wiring.
func CheckDir(dir string) ([]Violation, error) {
	cfg := &packages.Config{
		Mode: packages.NeedName |
			packages.NeedFiles |
			packages.NeedCompiledGoFiles |
			packages.NeedSyntax |
			packages.NeedTypes |
			packages.NeedTypesInfo,
		Dir:  dir,
		Fset: token.NewFileSet(),
	}
	pkgs, err := packages.Load(cfg, "./...")
	if err != nil {
		return nil, fmt.Errorf("packages.Load: %w", err)
	}

	// Fail loud on load errors: if the code doesn't type-check the
	// analyzer can't produce trustworthy results, and silently
	// returning zero violations would be worse than surfacing the
	// build failure.
	var loadErrs []string
	for _, pkg := range pkgs {
		for _, e := range pkg.Errors {
			loadErrs = append(loadErrs, e.Error())
		}
	}
	if len(loadErrs) > 0 {
		return nil, fmt.Errorf("package load errors:\n  %s",
			joinLines(loadErrs))
	}

	var out []Violation
	for _, pkg := range pkgs {
		for _, file := range pkg.Syntax {
			if isTestFile(pkg.Fset.Position(file.Pos()).Filename) {
				continue
			}
			ast.Inspect(file, func(n ast.Node) bool {
				switch node := n.(type) {
				case *ast.GoStmt:
					out = append(out, Violation{
						Pos:     pkg.Fset.Position(node.Pos()),
						Message: goStmtMessage,
					})
				case *ast.CallExpr:
					if msg := findForbiddenCall(pkg.TypesInfo, node); msg != "" {
						out = append(out, Violation{
							Pos:     pkg.Fset.Position(node.Pos()),
							Message: msg,
						})
					}
				}
				return true
			})
		}
	}
	return out, nil
}

// goStmtMessage is the diagnostic emitted for every `go` statement in
// a handler. It is a package-level constant so Analyzer.run and
// CheckDir report identically.
const goStmtMessage = "handler must not spawn goroutines (CONTRACT.md rule 2): return a CallWorkload effect and let the Runner manage the goroutine"

func joinLines(ss []string) string {
	out := ""
	for i, s := range ss {
		if i > 0 {
			out += "\n  "
		}
		out += s
	}
	return out
}
