package search

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

type ragFakeDoer struct {
	method string
	path   string
	body   interface{}
	raw    string
}

func (d *ragFakeDoer) DoJSON(_ context.Context, method, path string, body, out interface{}) error {
	d.method = method
	d.path = path
	d.body = body
	return json.Unmarshal([]byte(d.raw), out)
}

func ragBuildCommand(t *testing.T, changed map[string]string) *cobra.Command {
	t.Helper()
	cmd := &cobra.Command{}
	cmd.Flags().Int("candidate-k", 0, "")
	cmd.Flags().Int("top-k", 0, "")
	cmd.Flags().Bool("rerank", false, "")
	cmd.Flags().Int("rerank-top-n", 0, "")
	cmd.Flags().Int("deadline-ms", 0, "")
	cmd.Flags().Bool("highlights", false, "")
	for name, value := range changed {
		if err := cmd.Flags().Set(name, value); err != nil {
			t.Fatalf("set %s: %v", name, err)
		}
	}
	return cmd
}

func TestBuildRAGRequestLeavesDefaultsToServer(t *testing.T) {
	o := &ragSearchOptions{output: "json"}
	request, err := buildRAGRequest(ragBuildCommand(t, nil), "recovery plan", o)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(request)
	if err != nil {
		t.Fatal(err)
	}
	got := string(raw)
	if got != `{"query":"recovery plan"}` {
		t.Fatalf("unexpected wire request: %s", got)
	}
}

func TestBuildRAGRequestForwardsOnlyClientFacingOptions(t *testing.T) {
	o := &ragSearchOptions{
		roots:             []string{"drive/Home/Documents/"},
		mode:              "hybrid",
		candidateK:        40,
		topK:              10,
		rerank:            true,
		rerankTopN:        5,
		deadlineMS:        5000,
		includeHighlights: true,
		output:            "json",
	}
	cmd := ragBuildCommand(t, map[string]string{
		"candidate-k": "40", "top-k": "10", "rerank": "true",
		"rerank-top-n": "5", "deadline-ms": "5000", "highlights": "true",
	})
	request, err := buildRAGRequest(cmd, "北京大学", o)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(request)
	if err != nil {
		t.Fatal(err)
	}
	got := string(raw)
	for _, required := range []string{
		`"directoryRoots":["drive/Home/Documents/"]`,
		`"retrievalMode":"hybrid"`,
		`"candidateK":40`, `"topK":10`, `"enabled":true`,
		`"topN":5`, `"deadlineMs":5000`, `"includeHighlights":true`,
	} {
		if !strings.Contains(got, required) {
			t.Fatalf("wire request %s does not contain %s", got, required)
		}
	}
	for _, forbidden := range []string{"trusted_owner", "authorized_directory_roots", "files_v2:"} {
		if strings.Contains(got, forbidden) {
			t.Fatalf("wire request must not contain trusted field %q: %s", forbidden, got)
		}
	}
}

func TestBuildRAGRequestRejectsCanonicalAndUnboundedInputs(t *testing.T) {
	tests := []struct {
		name string
		o    ragSearchOptions
		set  map[string]string
	}{
		{name: "canonical root", o: ragSearchOptions{roots: []string{"files_v2:bob:drive/Home/"}, output: "json"}},
		{name: "lookalike path escape", o: ragSearchOptions{roots: []string{"drive/Home/../Other/"}, output: "json"}},
		{name: "candidate cap", o: ragSearchOptions{candidateK: 201, output: "json"}, set: map[string]string{"candidate-k": "201"}},
		{name: "top above candidates", o: ragSearchOptions{candidateK: 5, topK: 10, output: "json"}, set: map[string]string{"candidate-k": "5", "top-k": "10"}},
		{name: "rerank cap", o: ragSearchOptions{rerankTopN: 11, output: "json"}, set: map[string]string{"rerank-top-n": "11"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := buildRAGRequest(ragBuildCommand(t, tt.set), "query", &tt.o); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}

func TestExecuteRAGSearchUsesTrustedUserServiceEndpoint(t *testing.T) {
	doer := &ragFakeDoer{raw: `{"schema_version":"rag.retrieval.v1","status":"error","error":{"code":"invalid_request","message":"no authorized directory root is available"},"trace_id":"trace-1"}`}
	err := executeRAGSearch(context.Background(), doer, ragSearchRequest{Query: "query"}, "table")
	if err == nil || !strings.Contains(err.Error(), "no authorized directory root") {
		t.Fatalf("expected service error, got %v", err)
	}
	if doer.method != "POST" || doer.path != "/api/retrieval/query" {
		t.Fatalf("unexpected request %s %s", doer.method, doer.path)
	}
}

func TestSearchCommandIncludesRAG(t *testing.T) {
	cmd := NewSearchCommand(nil)
	found, _, err := cmd.Find([]string{"rag"})
	if err != nil {
		t.Fatal(err)
	}
	if found == nil || found.Name() != "rag" {
		t.Fatalf("rag command not registered: %#v", found)
	}
}
