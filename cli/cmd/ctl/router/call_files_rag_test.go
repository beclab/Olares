package router

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

func validFilesRAGRequest() filesRAGSearchRequest {
	return filesRAGSearchRequest{
		SchemaVersion:     filesRAGSchemaVersion,
		DirectoryRoots:    []string{"files_v2:alice:drive/Home/Work/"},
		OriginalQuery:     "港口安全标识",
		EffectiveQuery:    "港口安全标识",
		RetrievalMode:     "hybrid",
		CandidateK:        40,
		TopK:              10,
		Rerank:            filesRAGRerankRequest{Enabled: false, TopN: 5},
		DeadlineMS:        5_000,
		IncludeHighlights: true,
	}
}

func TestFilesRAGSearchUsesDelegatedRouteAndExactPublicContract(t *testing.T) {
	want := validFilesRAGRequest()
	var got filesRAGSearchRequest
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != epFilesRAGSearch {
			t.Errorf("request = %s %s, want POST %s", r.Method, r.URL.Path, epFilesRAGSearch)
		}
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Errorf("decode request: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"schema_version":"rag.retrieval.v1","status":"ok","hits":[],"abstention":{"applied":true,"reason":"no_match"},"degraded":[],"timing_ms":{},"trace_id":"t-1","future_field":{"kept":true}}`)
	}))
	defer srv.Close()

	raw, err := callFilesRAGSearch(t.Context(), newRouterClient(srv.Client(), srv.URL, "alice@olares"), want)
	if err != nil {
		t.Fatalf("call: %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("request:\n got %#v\nwant %#v", got, want)
	}
	if !bytes.Contains(raw, []byte(`"future_field"`)) {
		t.Errorf("unknown response field was lost: %s", raw)
	}
	var out bytes.Buffer
	if err := printJSON(&out, raw); err != nil {
		t.Fatalf("print JSON: %v", err)
	}
	if !strings.Contains(out.String(), `"future_field"`) || !strings.Contains(out.String(), `"kept": true`) {
		t.Errorf("JSON output lost the future field: %s", out.String())
	}
}

func TestFilesRAGSearchRequiresHostDelegationBeforeNetworkWork(t *testing.T) {
	t.Setenv(dataPlaneURLEnv, "")
	t.Setenv(dataPlaneKeyEnv, "")
	cmd := newCallSearchCommand(&cmdutil.Factory{})
	cmd.SetArgs([]string{"--source", "files", "query"})
	err := cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), dataPlaneURLEnv) {
		t.Fatalf("platform identity should be rejected with host guidance, got %v", err)
	}

	t.Setenv(dataPlaneURLEnv, "http://lares.example/v1")
	t.Setenv(dataPlaneKeyEnv, "sk-wrong-principal")
	cmd = newCallSearchCommand(&cmdutil.Factory{})
	cmd.SetArgs([]string{"--source", "files", "query"})
	err = cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "host-delegated") {
		t.Fatalf("API-key identity should be rejected, got %v", err)
	}
}

func TestSearchSourceFlagsCannotSilentlyChangeTheOtherSearch(t *testing.T) {
	cmd := newCallSearchCommand(&cmdutil.Factory{})
	cmd.SetArgs([]string{"--mode", "lexical", "web query"})
	if err := cmd.Execute(); err == nil || !strings.Contains(err.Error(), "only to --source files") {
		t.Fatalf("Files-only flag on web search should fail before networking, got %v", err)
	}

	t.Setenv(dataPlaneURLEnv, "http://lares.example/v1")
	t.Setenv(dataPlaneKeyEnv, "")
	cmd = newCallSearchCommand(&cmdutil.Factory{})
	cmd.SetArgs([]string{"--source", "files", "--model", "default-search", "file query"})
	if err := cmd.Execute(); err == nil || !strings.Contains(err.Error(), "not a model route") {
		t.Fatalf("model flag on Files search should fail before networking, got %v", err)
	}
}

func TestFilesRAGSearchBoundsAreCheckedLocally(t *testing.T) {
	for _, tc := range []struct {
		name string
		edit func(*filesRAGSearchRequest)
		want string
	}{
		{"mode", func(r *filesRAGSearchRequest) { r.RetrievalMode = "magic" }, "--mode"},
		{"candidate", func(r *filesRAGSearchRequest) { r.CandidateK = 201 }, "--candidate-k"},
		{"top greater candidate", func(r *filesRAGSearchRequest) { r.CandidateK, r.TopK = 5, 6 }, "--limit"},
		{"rerank", func(r *filesRAGSearchRequest) { r.Rerank.TopN = 11 }, "--rerank-top-n"},
		{"deadline", func(r *filesRAGSearchRequest) { r.DeadlineMS = (30*time.Second + time.Millisecond).Milliseconds() }, "--deadline"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			req := validFilesRAGRequest()
			tc.edit(&req)
			err := validateFilesRAGSearchRequest(req)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("got %v, want error naming %s", err, tc.want)
			}
		})
	}
}

func TestRenderFilesRAGSearchMakesEvidenceVisible(t *testing.T) {
	var resp filesRAGSearchResponse
	if err := json.Unmarshal([]byte(`{
		"hits":[{"rank":1,"resource_uri":"files_v2:alice:drive/Home/guide.md","content":"The E1042 recovery steps","scores":{"fused":0.91}}],
		"abstention":{"applied":false,"reason":""},"degraded":[]
	}`), &resp); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	if err := renderFilesRAGSearch(&out, &resp); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"SOURCE", "guide.md", "E1042", "0.91"} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("table missing %q:\n%s", want, out.String())
		}
	}
}
