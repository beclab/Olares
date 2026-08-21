package router

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
	"testing"
)

// providerModelModes is a copy of a list Router owns, and the copy goes stale:
// it sat two modes behind for a release. That on its own is survivable — a
// stale hint is a hint. What made it a bug is that three verbs were using it to
// refuse input, so `--mode music_generation` was rejected here for a value the
// database accepts, with a message that read as though the CLI knew.
//
// The rule that replaced it is that this list may be shown and never consulted.
// Router validates the mode on all three routes and names its own list when it
// refuses, so removing the local gate costs nothing and cannot go stale.
//
// This test is over the syntax tree because the distinction is structural: a
// grep for the identifier cannot tell `strings.Join(providerModelModes, ", ")`
// in a help string from `containsString(providerModelModes, mode)` guarding a
// return.
func TestTheModeListIsNeverUsedToRefuseAnything(t *testing.T) {
	// The predicates a gate would be built out of. Joining the list into a
	// message is fine; asking it whether a word is allowed is not.
	deciders := map[string]bool{"containsString": true, "slices.Contains": true}

	for _, path := range packageFiles(t) {
		if strings.HasSuffix(path, "_test.go") {
			continue
		}
		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, path, nil, 0)
		if err != nil {
			t.Fatalf("parse %s: %v", path, err)
		}
		ast.Inspect(file, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			name := callName(call.Fun)
			if !deciders[name] {
				return true
			}
			for _, arg := range call.Args {
				id, ok := arg.(*ast.Ident)
				if !ok || id.Name != "providerModelModes" {
					continue
				}
				t.Errorf("%s:%d: %s(providerModelModes, …) makes the local copy decide. "+
					"Send the value and let Router refuse it — its message names the real list",
					path, fset.Position(call.Pos()).Line, name)
			}
			return true
		})
	}
}

func callName(fun ast.Expr) string {
	switch f := fun.(type) {
	case *ast.Ident:
		return f.Name
	case *ast.SelectorExpr:
		if pkg, ok := f.X.(*ast.Ident); ok {
			return pkg.Name + "." + f.Sel.Name
		}
		return f.Sel.Name
	}
	return ""
}
