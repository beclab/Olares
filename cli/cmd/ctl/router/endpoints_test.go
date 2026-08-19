package router

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// A route spelled inline is a route nothing can check, and the failure it
// produces is the least legible one this tree has: a 404 with no envelope,
// which reads as "Router is not installed" rather than as "this path moved".
// endpoints.go exists so a rename is one edit; this test is what keeps it that
// way, because adding a request is exactly the moment somebody reaches for a
// literal.
//
// The scan is over the syntax tree rather than the file's text, so a path
// named in a comment — which the doc comments do constantly, and should — is
// not mistaken for one being sent.
func TestNoRouteIsSpelledOutsideEndpointsGo(t *testing.T) {
	// Prefixes that can only be a route on one of the three surfaces this
	// tree addresses. A bare "/" is not among them: a lone slash is a path
	// separator far more often than a request.
	routish := []string{"/console/api", "/v1/", "/api/", "/healthz"}

	for _, path := range packageFiles(t) {
		base := filepath.Base(path)
		if base == "endpoints.go" || strings.HasSuffix(base, "_test.go") {
			continue
		}
		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, path, nil, 0)
		if err != nil {
			t.Fatalf("parse %s: %v", path, err)
		}
		ast.Inspect(file, func(n ast.Node) bool {
			lit, ok := n.(*ast.BasicLit)
			if !ok || lit.Kind != token.STRING {
				return true
			}
			value, err := strconv.Unquote(lit.Value)
			if err != nil {
				return true
			}
			for _, prefix := range routish {
				if !strings.HasPrefix(value, prefix) {
					continue
				}
				// The Olares Desktop is a fourth host and a different
				// product: discovery asks it which applications are
				// installed, and that route is not Router's to declare.
				if base == "discovery.go" && value == "/api/myapps" {
					return true
				}
				t.Errorf("%s: %q is a route written inline; declare it in endpoints.go and use the name",
					fset.Position(lit.Pos()), value)
			}
			return true
		})
	}
}

// The two prefixes decide which credential a request carries — the console
// plane runs on the profile's session and /v1 refuses one — so a route
// declared without either is a route whose authentication is a guess.
func TestNothingReferencesThePlanePrefixesOutsideEndpointsGo(t *testing.T) {
	for _, path := range packageFiles(t) {
		base := filepath.Base(path)
		if base == "endpoints.go" || strings.HasSuffix(base, "_test.go") {
			continue
		}
		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, path, nil, 0)
		if err != nil {
			t.Fatalf("parse %s: %v", path, err)
		}
		ast.Inspect(file, func(n ast.Node) bool {
			id, ok := n.(*ast.Ident)
			if !ok {
				return true
			}
			if id.Name == "consoleAPI" || id.Name == "dataPlaneAPI" {
				t.Errorf("%s: %s is concatenated here; endpoints.go is where a path is assembled",
					fset.Position(id.Pos()), id.Name)
			}
			return true
		})
	}
}

// An id reaches these constructors from a Router row or from an argument, and
// one carrying a slash would otherwise silently address a different route.
func TestPathSegmentsAreEscaped(t *testing.T) {
	cases := []struct {
		name string
		got  string
		want string
	}{
		{"provider", epProvider("a/b"), consoleAPI + "/providers/a%2Fb"},
		{"provider model", epProviderModel("p", "meta-llama/Llama-3"),
			consoleAPI + "/providers/p/models/meta-llama%2FLlama-3"},
		{"api key", epAPIKey("k/1"), consoleAPI + "/api-keys/k%2F1"},
		{"audit log", epAuditLog("x y"), consoleAPI + "/audit-logs/x%20y"},
		{"ocr task", epOCRTask("t/2"), dataPlaneAPI + "/ocr/tasks/t%2F2"},
	}
	for _, c := range cases {
		if c.got != c.want {
			t.Errorf("%s: got %q want %q", c.name, c.got, c.want)
		}
	}
}

// withQuery leaves a path alone when there is nothing to append, so a caller
// assembling optional filters does not have to know whether it ended up with
// any — and a path that reached the wire with a bare "?" would be a different
// request than the one intended.
func TestWithQueryOmitsAnEmptyQuery(t *testing.T) {
	if got := withQuery(epProviders, nil); got != epProviders {
		t.Errorf("nil values: got %q want %q", got, epProviders)
	}
	if got := withQuery(epProviders, map[string][]string{}); got != epProviders {
		t.Errorf("empty values: got %q want %q", got, epProviders)
	}
	if got := withQuery(epSpendLogs, map[string][]string{"model": {"a/b"}}); got != epSpendLogs+"?model=a%2Fb" {
		t.Errorf("one value: got %q", got)
	}
}

func packageFiles(t *testing.T) []string {
	t.Helper()
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("read package directory: %v", err)
	}
	var found []string
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".go" {
			continue
		}
		found = append(found, entry.Name())
	}
	if len(found) < 30 {
		t.Fatalf("found only %d Go files; the package moved", len(found))
	}
	return found
}
