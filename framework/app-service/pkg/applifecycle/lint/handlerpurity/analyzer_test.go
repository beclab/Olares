package handlerpurity

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

// TestAnalyzer runs the handlerpurity analyzer over the testdata
// packages and asserts that:
//
//   - "ok" produces zero diagnostics (pure handler baseline)
//   - "goroutine" flags every `go` statement (CONTRACT.md rule 2)
//   - "sleep" flags time.Sleep but not unrelated time.* calls (rule 3.1)
//   - "poll" flags every wait.Poll* call via the pkg-suffix matcher,
//     using a stub "poll/wait" package that mirrors the production
//     k8s.io/apimachinery/pkg/util/wait surface (rule 3.2)
//
// The `// want "..."` comments in each testdata file drive the
// assertions; each comment is a regexp matched against the diagnostic
// message on that source line.
func TestAnalyzer(t *testing.T) {
	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, Analyzer,
		"ok",
		"goroutine",
		"sleep",
		"poll",
	)
}
