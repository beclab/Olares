package cp

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// newTestClient is the same kind of httptest harness rm_test.go uses:
// stand up a server, hand the caller a Client whose BaseURL points at
// it, and let the test inspect what landed on the wire.
func newTestClient(t *testing.T, h http.Handler) (*Client, *httptest.Server) {
	t.Helper()
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	return &Client{
		HTTPClient: srv.Client(),
		BaseURL:    srv.URL,
	}, srv
}

// TestPlan_DropIntoDir is the bread-and-butter case: `cp foo.pdf
// bar/Documents/`. The destination's basename comes from the source
// and the parent's trailing slash is preserved.
func TestPlan_DropIntoDir(t *testing.T) {
	srcs := []Source{
		{FileType: "drive", Extend: "Home", SubPath: "/Documents/foo.pdf"},
	}
	dst := Destination{FileType: "drive", Extend: "Home", SubPath: "/Backups/", IsDirIntent: true}

	ops, err := Plan(srcs, dst, ActionCopy, false, "node-a", "")
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}
	if len(ops) != 1 {
		t.Fatalf("want 1 op, got %d", len(ops))
	}
	got := ops[0]
	if got.Source != "/drive/Home/Documents/foo.pdf" {
		t.Errorf("Source: got %q", got.Source)
	}
	if got.Destination != "/drive/Home/Backups/foo.pdf" {
		t.Errorf("Destination: got %q", got.Destination)
	}
	if got.Action != ActionCopy {
		t.Errorf("Action: got %q", got.Action)
	}
	if got.Node != "node-a" {
		t.Errorf("Node: got %q (want default fallback)", got.Node)
	}
}

// TestPlan_DropIntoDir_RecursiveDir confirms that a dir source
// preserves its trailing slash on both Source and Destination, and
// that --recursive unblocks the dir-intent check.
func TestPlan_DropIntoDir_RecursiveDir(t *testing.T) {
	srcs := []Source{
		{FileType: "drive", Extend: "Home", SubPath: "/Documents/sub/", IsDirIntent: true},
	}
	dst := Destination{FileType: "drive", Extend: "Home", SubPath: "/Backups/", IsDirIntent: true}

	ops, err := Plan(srcs, dst, ActionCopy, true, "node-a", "")
	if err != nil {
		t.Fatalf("Plan with -r: %v", err)
	}
	if ops[0].Source != "/drive/Home/Documents/sub/" {
		t.Errorf("Source: got %q", ops[0].Source)
	}
	if ops[0].Destination != "/drive/Home/Backups/sub/" {
		t.Errorf("Destination: got %q", ops[0].Destination)
	}
	if !ops[0].IsDir {
		t.Errorf("IsDir: want true")
	}
}

// TestPlan_DirRequiresRecursive replicates Unix `cp` / `mv`'s refusal
// to operate on a directory without -r/-R; the error must name the
// offending path and mention the flag.
func TestPlan_DirRequiresRecursive(t *testing.T) {
	srcs := []Source{
		{FileType: "drive", Extend: "Home", SubPath: "/Documents/sub/", IsDirIntent: true},
	}
	dst := Destination{FileType: "drive", Extend: "Home", SubPath: "/Backups/", IsDirIntent: true}

	_, err := Plan(srcs, dst, ActionCopy, false, "node-a", "")
	if err == nil {
		t.Fatal("expected error without --recursive")
	}
	if !strings.Contains(err.Error(), "directory") || !strings.Contains(err.Error(), "Documents/sub") {
		t.Errorf("error should name the dir + flag, got: %v", err)
	}
}

// TestPlan_RenameMode covers exact-target / single-source mode where
// the destination has no trailing slash and is treated as the full
// target path (Unix `cp foo bar` style — bar is the new name).
func TestPlan_RenameMode(t *testing.T) {
	srcs := []Source{
		{FileType: "drive", Extend: "Home", SubPath: "/Documents/foo.pdf"},
	}
	dst := Destination{FileType: "drive", Extend: "Home", SubPath: "/Documents/foo-2025.pdf"}

	ops, err := Plan(srcs, dst, ActionMove, false, "node-a", "")
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}
	if ops[0].Destination != "/drive/Home/Documents/foo-2025.pdf" {
		t.Errorf("Destination: got %q", ops[0].Destination)
	}
	if ops[0].Action != ActionMove {
		t.Errorf("Action: got %q", ops[0].Action)
	}
}

// TestPlan_MultiSourceRequiresDirDst guards the "cp a b c" → "c must
// be a directory" Unix invariant. Without it, multi-source rename has
// no defined semantics.
func TestPlan_MultiSourceRequiresDirDst(t *testing.T) {
	srcs := []Source{
		{FileType: "drive", Extend: "Home", SubPath: "/a.pdf"},
		{FileType: "drive", Extend: "Home", SubPath: "/b.pdf"},
	}
	dst := Destination{FileType: "drive", Extend: "Home", SubPath: "/c.pdf"}

	_, err := Plan(srcs, dst, ActionCopy, false, "node-a", "")
	if err == nil {
		t.Fatal("expected error for multi-source + non-dir target")
	}
	if !strings.Contains(err.Error(), "directory") || !strings.Contains(err.Error(), "'/'") {
		t.Errorf("error should mention the trailing-slash requirement, got: %v", err)
	}
}

// TestPlan_RefusesRoot blocks `cp drive/Home/ ...` (and the same for
// any extend root). Operating on a whole volume root through this
// endpoint is not a meaningful UX and the cost of doing it
// accidentally is huge.
func TestPlan_RefusesRoot(t *testing.T) {
	srcs := []Source{
		{FileType: "drive", Extend: "Home", SubPath: "/", IsDirIntent: true},
	}
	dst := Destination{FileType: "drive", Extend: "Home", SubPath: "/Backups/", IsDirIntent: true}

	_, err := Plan(srcs, dst, ActionCopy, true, "node-a", "")
	if err == nil {
		t.Fatal("expected error for extend-root source")
	}
	if !strings.Contains(err.Error(), "root") {
		t.Errorf("error should mention 'root', got: %v", err)
	}
}

// TestPlan_RefusesSameSrcDst rejects `cp foo foo` outright — the
// frontend silently allows it (the backend then no-ops or errors),
// but on a CLI it's almost always a typo.
func TestPlan_RefusesSameSrcDst(t *testing.T) {
	srcs := []Source{
		{FileType: "drive", Extend: "Home", SubPath: "/Documents/foo.pdf"},
	}
	dst := Destination{FileType: "drive", Extend: "Home", SubPath: "/Documents/foo.pdf"}

	_, err := Plan(srcs, dst, ActionCopy, false, "node-a", "")
	if err == nil {
		t.Fatal("expected error for src==dst")
	}
	if !strings.Contains(err.Error(), "same") {
		t.Errorf("error should mention 'same', got: %v", err)
	}
}

// TestPlan_RefusesCycle catches the cp-into-itself trap: copying
// /a/ → /a/sub/ would create an infinitely-recursing tree on the
// server side.
func TestPlan_RefusesCycle(t *testing.T) {
	srcs := []Source{
		{FileType: "drive", Extend: "Home", SubPath: "/Backups/", IsDirIntent: true},
	}
	dst := Destination{FileType: "drive", Extend: "Home", SubPath: "/Backups/sub/", IsDirIntent: true}

	_, err := Plan(srcs, dst, ActionCopy, true, "node-a", "")
	if err == nil {
		t.Fatal("expected cycle error")
	}
	if !strings.Contains(err.Error(), "cycle") {
		t.Errorf("error should mention 'cycle', got: %v", err)
	}
}

// TestResolveNode_Cascade exercises the full
// flagNode > dst External/Cache > src External/Cache > defaultNode
// cascade. Keeping these in one table-driven test makes it cheap to
// add a new fileType to pasteMultiNodeFileTypes without re-checking
// every branch.
func TestResolveNode_Cascade(t *testing.T) {
	cases := []struct {
		name                                                              string
		srcType, srcExtend, dstType, dstExtend, defaultNode, flag, expect string
	}{
		{
			name: "flag overrides everything",
			srcType: "external", srcExtend: "node-x",
			dstType: "external", dstExtend: "node-y",
			defaultNode: "default", flag: "node-flag",
			expect: "node-flag",
		},
		{
			name: "dst external wins over src external",
			srcType: "external", srcExtend: "node-src",
			dstType: "external", dstExtend: "node-dst",
			defaultNode: "default",
			expect:      "node-dst",
		},
		{
			name: "dst cache wins over default",
			srcType: "drive", srcExtend: "Home",
			dstType: "cache", dstExtend: "node-cache",
			defaultNode: "default",
			expect:      "node-cache",
		},
		{
			name: "src external used when dst non-nodey",
			srcType: "external", srcExtend: "node-src",
			dstType: "drive", dstExtend: "Home",
			defaultNode: "default",
			expect:      "node-src",
		},
		{
			name: "fallback to default when neither side is nodey",
			srcType: "drive", srcExtend: "Home",
			dstType: "drive", dstExtend: "Home",
			defaultNode: "default",
			expect:      "default",
		},
		{
			name: "external with empty extend does NOT override (defensive)",
			srcType: "drive", srcExtend: "Home",
			dstType: "external", dstExtend: "",
			defaultNode: "default",
			expect:      "default",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := ResolveNode(tc.srcType, tc.srcExtend, tc.dstType, tc.dstExtend, tc.defaultNode, tc.flag)
			if got != tc.expect {
				t.Errorf("ResolveNode = %q, want %q", got, tc.expect)
			}
		})
	}
}

// TestPlan_FlagNodeOverridesAll confirms that passing flagNode through
// Plan cascades into the Op even when External/Cache would otherwise
// pick a path-derived node.
func TestPlan_FlagNodeOverridesAll(t *testing.T) {
	srcs := []Source{
		{FileType: "external", Extend: "node-src", SubPath: "/foo.pdf"},
	}
	dst := Destination{FileType: "drive", Extend: "Home", SubPath: "/Backups/", IsDirIntent: true}

	ops, err := Plan(srcs, dst, ActionCopy, false, "default", "node-forced")
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}
	if ops[0].Node != "node-forced" {
		t.Errorf("Node: got %q (want flag override)", ops[0].Node)
	}
}

// TestPasteOne_WireShape inspects the actual PATCH that lands on the
// server: URL encoding of {node}, JSON body shape, action verb. This
// is the one that breaks loudly if either side of the protocol drifts.
func TestPasteOne_WireShape(t *testing.T) {
	var (
		gotMethod string
		gotPath   string
		gotCType  string
		gotBody   pasteRequestBody
	)
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		gotCType = r.Header.Get("Content-Type")
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &gotBody)
		// Match the shape `pasteAction` reads in the web app: the
		// JSON body is the response (no axios-like data wrapper).
		_, _ = io.WriteString(w, `{"task_id":"task-123"}`)
	}))
	op := Op{
		Action:      ActionCopy,
		Source:      "/drive/Home/Documents/foo.pdf",
		Destination: "/drive/Home/Backups/foo.pdf",
		Node:        "node-a",
	}
	taskID, err := client.PasteOne(context.Background(), op)
	if err != nil {
		t.Fatalf("PasteOne: %v", err)
	}
	if taskID != "task-123" {
		t.Errorf("taskID: got %q", taskID)
	}
	if gotMethod != http.MethodPatch {
		t.Errorf("Method: got %s", gotMethod)
	}
	if gotPath != "/api/paste/node-a/" {
		t.Errorf("Path: got %q", gotPath)
	}
	if !strings.HasPrefix(gotCType, "application/json") {
		t.Errorf("Content-Type: got %q", gotCType)
	}
	if gotBody.Action != ActionCopy ||
		gotBody.Source != "/drive/Home/Documents/foo.pdf" ||
		gotBody.Destination != "/drive/Home/Backups/foo.pdf" {
		t.Errorf("body: got %+v", gotBody)
	}
}

// TestPasteOne_CodeMinusOne mirrors the web app's `if (res.data.code
// === -1)` branch: a 2xx response with `code: -1` is a server-side
// rejection (typically a malformed path) and must surface as an
// error, not a silent success.
func TestPasteOne_CodeMinusOne(t *testing.T) {
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, `{"code":-1,"message":"bad path"}`)
	}))
	_, err := client.PasteOne(context.Background(), Op{
		Action: ActionCopy, Source: "/a", Destination: "/b", Node: "n",
	})
	if err == nil {
		t.Fatal("expected error for code:-1")
	}
	if !strings.Contains(err.Error(), "bad path") {
		t.Errorf("error should bubble up the server message, got: %v", err)
	}
}

// TestPasteOne_NoTaskID covers the "queued but no handle" failure mode:
// a 2xx response without task_id is still useless to the caller, so
// we error rather than returning "" and pretending it worked.
func TestPasteOne_NoTaskID(t *testing.T) {
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, `{}`)
	}))
	_, err := client.PasteOne(context.Background(), Op{
		Action: ActionCopy, Source: "/a", Destination: "/b", Node: "n",
	})
	if err == nil {
		t.Fatal("expected error for missing task_id")
	}
}

// TestPasteOne_HTTPError surfaces non-2xx responses as *HTTPError —
// same contract the cobra layer uses to reformat 401 / 403 / 404 with
// friendly CTAs.
func TestPasteOne_HTTPError(t *testing.T) {
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = io.WriteString(w, `{"error":"nope"}`)
	}))
	_, err := client.PasteOne(context.Background(), Op{
		Action: ActionCopy, Source: "/a", Destination: "/b", Node: "n",
	})
	if err == nil {
		t.Fatal("expected error")
	}
	var hErr *HTTPError
	if !errors.As(err, &hErr) {
		t.Fatalf("want *HTTPError, got %T", err)
	}
	if hErr.Status != http.StatusForbidden {
		t.Errorf("status: got %d", hErr.Status)
	}
}

// TestFetchNodes_Envelope confirms we read /api/nodes/ with the same
// {data: {nodes: [...]}} shape as the web app's fetchNodeList.
func TestFetchNodes_Envelope(t *testing.T) {
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, `{"data":{"nodes":[{"name":"node-a","master":true},{"name":"node-b"}]}}`)
	}))
	nodes, err := client.FetchNodes(context.Background())
	if err != nil {
		t.Fatalf("FetchNodes: %v", err)
	}
	if len(nodes) != 2 || nodes[0].Name != "node-a" || !nodes[0].Master || nodes[1].Name != "node-b" {
		t.Errorf("nodes: got %+v", nodes)
	}
}
