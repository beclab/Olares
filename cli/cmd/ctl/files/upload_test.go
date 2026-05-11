// upload_test.go: focused unit tests for the namespace-dispatch
// helper that translates a parsed FrontendPath into the concrete
// upload-protocol parameters runUpload + BuildPlan need.
//
// Why a dedicated test:
//
//   - The dispatcher fans out into multiple shapes (drive/Home,
//     drive/Data, sync/<repo>, cache/<node>, external/<node>,
//     awss3/<account>, google/<account>, dropbox/<account>) — each
//     with its own apiRoot/chunkRoot/driveType/pathNode tuple. A
//     table-driven test makes "did we wire all of them correctly?"
//     a compile-time-cheap check; in particular, anything that
//     ships another namespace MUST add a row here, which forces
//     the author to decide chunkRoot (does Seafile-style inside-
//     repo path apply?) and pathNode (does the path supply the
//     upload node?) up-front instead of leaving it implicit.
//   - The HTTP-500 regression that motivated the original Sync split
//     (chunkRoot ≠ apiRoot for Seafile-backed namespaces) was caught
//     LATE — it didn't surface in unit tests and only blew up
//     end-to-end. Pinning the per-namespace tuple here lets future
//     changes assert the contract before any wire roundtrip happens.
//   - Negative tests pin the user-facing error messages: the moment
//     a typo migrates "drive/Home or sync/..." into something else,
//     this test will fail and the author has to update both the
//     dispatcher AND the help text in lockstep.
package files

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/beclab/Olares/cli/internal/files/upload"
)

// TestUploadRootAndDriveType exercises every supported upload
// namespace plus the rejection arms (tencent's octet-only protocol,
// drive/<unknown>, missing extend). The fields are duplicated
// explicitly per row rather than computed because the whole point
// of the test is to catch a future "I'll just clean up the
// redundancy" change that flips the wrong knob — having the literal
// expected values in the table makes diff review trivial.
func TestUploadRootAndDriveType(t *testing.T) {
	cases := []struct {
		name           string
		path           string
		wantAPIRoot    string
		wantChunkRoot  string
		wantDriveType  string
		wantPathNode   string
		wantErrSubstr  string
	}{
		{
			name:          "drive Home",
			path:          "drive/Home/Documents/",
			wantAPIRoot:   "/drive/Home",
			wantChunkRoot: "/drive/Home",
			wantDriveType: "Drive",
			wantPathNode:  "",
		},
		{
			name:          "drive Data",
			path:          "drive/Data/Backups/",
			wantAPIRoot:   "/drive/Data",
			wantChunkRoot: "/drive/Data",
			wantDriveType: "Data",
			wantPathNode:  "",
		},
		{
			name: "sync repo (chunkRoot empty: Seafile reads parent_dir " +
				"as inside-repo path because the upload token already pins " +
				"the repo)",
			path:          "sync/repo-abc/Documents/",
			wantAPIRoot:   "/sync/repo-abc",
			wantChunkRoot: "",
			wantDriveType: "Sync",
			wantPathNode:  "",
		},
		{
			name: "cache (path's <node> IS the upload node, " +
				"so pathNode short-circuits the /api/nodes/ round-trip)",
			path:          "cache/node-1/AppName/data/",
			wantAPIRoot:   "/cache/node-1",
			wantChunkRoot: "/cache/node-1",
			wantDriveType: "Cache",
			wantPathNode:  "node-1",
		},
		{
			name:          "external (same pathNode short-circuit as cache)",
			path:          "external/node-1/hdd1/Movies/",
			wantAPIRoot:   "/external/node-1",
			wantChunkRoot: "/external/node-1",
			wantDriveType: "External",
			wantPathNode:  "node-1",
		},
		// Cloud drives whose v2 dataAPIs inherit the regular Drive
		// chunk pipeline without overrides (Awss3DataAPI /
		// GoogleDataAPI / DropboxDataAPI). chunkRoot == apiRoot,
		// pathNode is empty (these all use masterNode default via
		// DriveDataAPI.getUploadNode()).
		{
			name:          "awss3 (cloud drive, regular multipart pipeline)",
			path:          "awss3/account-x/bucket/Backups/",
			wantAPIRoot:   "/awss3/account-x",
			wantChunkRoot: "/awss3/account-x",
			wantDriveType: "Awss3",
			wantPathNode:  "",
		},
		{
			name:          "google (cloud drive, regular multipart pipeline)",
			path:          "google/account-x/Documents/",
			wantAPIRoot:   "/google/account-x",
			wantChunkRoot: "/google/account-x",
			wantDriveType: "Google",
			wantPathNode:  "",
		},
		{
			name:          "dropbox (cloud drive, regular multipart pipeline)",
			path:          "dropbox/account-x/Notes/",
			wantAPIRoot:   "/dropbox/account-x",
			wantChunkRoot: "/dropbox/account-x",
			wantDriveType: "Dropbox",
			wantPathNode:  "",
		},
		// Negative cases: every supported-namespace error path goes
		// through the same human-readable list, so anchoring on a
		// shared substring guards both the dispatcher and the help
		// text against drift.
		//
		// Note: the path parser already pre-rejects bare drive/<bad>
		// (driveExtends only accepts Home/Data), so the dispatcher's
		// drive-default arm is defense-in-depth — exercised via the
		// direct-construction sibling test below, not here.
		{
			name: "tencent rejected (octet /drive/direct_upload_file " +
				"protocol the CLI does not implement)",
			path: "tencent/account-x/folder/",
			// Anchor on the protocol name so both the diagnostic
			// "what's wrong" and the verb's "we know we can't do
			// this yet" are pinned simultaneously.
			wantErrSubstr: "/drive/direct_upload_file/<task_id>",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fp, err := ParseFrontendPath(tc.path)
			if err != nil {
				t.Fatalf("ParseFrontendPath(%q): %v", tc.path, err)
			}
			gotAPIRoot, gotChunkRoot, gotDriveType, gotPathNode, gotErr :=
				uploadRootAndDriveType(fp)
			if tc.wantErrSubstr != "" {
				if gotErr == nil {
					t.Fatalf("uploadRootAndDriveType(%q) err = nil, want substring %q",
						tc.path, tc.wantErrSubstr)
				}
				if !strings.Contains(gotErr.Error(), tc.wantErrSubstr) {
					t.Fatalf("uploadRootAndDriveType(%q) err = %q, want substring %q",
						tc.path, gotErr.Error(), tc.wantErrSubstr)
				}
				return
			}
			if gotErr != nil {
				t.Fatalf("uploadRootAndDriveType(%q): %v", tc.path, gotErr)
			}
			if gotAPIRoot != tc.wantAPIRoot {
				t.Errorf("apiRoot = %q, want %q", gotAPIRoot, tc.wantAPIRoot)
			}
			if gotChunkRoot != tc.wantChunkRoot {
				t.Errorf("chunkRoot = %q, want %q", gotChunkRoot, tc.wantChunkRoot)
			}
			if gotDriveType != tc.wantDriveType {
				t.Errorf("driveType = %q, want %q", gotDriveType, tc.wantDriveType)
			}
			if gotPathNode != tc.wantPathNode {
				t.Errorf("pathNode = %q, want %q", gotPathNode, tc.wantPathNode)
			}
		})
	}
}

// TestUploadRootAndDriveType_DirectConstruction exercises arms that
// are unreachable through ParseFrontendPath (because the parser
// already rejects them) but are kept in the dispatcher for
// defense-in-depth. Constructing FrontendPath directly is the only
// way to exercise these arms; without this test, a future refactor
// that loosens the parser could silently regress the dispatcher's
// validation contract.
func TestUploadRootAndDriveType_DirectConstruction(t *testing.T) {
	cases := []struct {
		name    string
		fp      FrontendPath
		wantSub string
	}{
		{
			name:    "drive unknown extend",
			fp:      FrontendPath{FileType: "drive", Extend: "Other", SubPath: "/"},
			wantSub: "drive extend must be Home or Data",
		},
		{
			name:    "sync empty extend",
			fp:      FrontendPath{FileType: "sync", Extend: "", SubPath: "/"},
			wantSub: "sync extend(repo_id) must be non-empty",
		},
		{
			name:    "cache empty extend",
			fp:      FrontendPath{FileType: "cache", Extend: "", SubPath: "/"},
			wantSub: "cache extend(node) must be non-empty",
		},
		{
			name:    "external empty extend",
			fp:      FrontendPath{FileType: "external", Extend: "", SubPath: "/"},
			wantSub: "external extend(node) must be non-empty",
		},
		{
			name:    "unknown file type",
			fp:      FrontendPath{FileType: "internal", Extend: "x", SubPath: "/"},
			wantSub: "upload destination must be under",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, _, _, _, err := uploadRootAndDriveType(tc.fp)
			if err == nil {
				t.Fatalf("err = nil, want substring %q", tc.wantSub)
			}
			if !strings.Contains(err.Error(), tc.wantSub) {
				t.Errorf("err = %q, want substring %q", err.Error(), tc.wantSub)
			}
		})
	}
}

// TestSubPathForBuildPlan pins the FrontendPath.SubPath →
// upload.BuildPlan-input conversion so the namespace-root regression
// (`upload ./mydir/ drive/Home/` failing with `remote "" must end
// with '/'`) cannot return without breaking this test.
//
// The conversion has to satisfy two invariants simultaneously:
//
//  1. The trailing-slash directory hint MUST survive when the
//     destination is the namespace root (SubPath="/"). Stripping the
//     leading '/' verbatim collapses the only slash, leaving "" which
//     BuildPlan reads as "no directory hint" and rejects directory
//     uploads.
//  2. Non-root paths MUST come through in the relative form
//     BuildPlan expects ("Documents/Backups/"), not the absolute
//     form ParseFrontendPath emits ("/Documents/Backups/"). Otherwise
//     ParentDir computation in parentDirFor would double the leading
//     slash on the way out.
//
// Anchoring this here (rather than implicitly through a runUpload
// integration test) keeps the contract obvious for future readers:
// the conversion is a self-contained function with a tiny test, so
// any change has to update both at once.
func TestSubPathForBuildPlan(t *testing.T) {
	cases := []struct {
		name     string
		path     string
		wantSub  string
	}{
		// The bug: bare extend (with or without trailing slash) used
		// to produce "" → BuildPlan rejected directory uploads. Both
		// forms must now return "/" so directory uploads to the
		// namespace root succeed across every fileType.
		{
			name:    "drive Home root, trailing slash (regression: was '', now '/')",
			path:    "drive/Home/",
			wantSub: "/",
		},
		{
			name:    "drive Home bare extend, no trailing slash (parser synthesizes SubPath='/')",
			path:    "drive/Home",
			wantSub: "/",
		},
		{
			name:    "drive Data root",
			path:    "drive/Data/",
			wantSub: "/",
		},
		{
			name:    "sync repo root",
			path:    "sync/repo-abc/",
			wantSub: "/",
		},
		{
			name:    "cache node root",
			path:    "cache/node-1/",
			wantSub: "/",
		},
		{
			name:    "external node root",
			path:    "external/node-1/",
			wantSub: "/",
		},
		{
			name:    "awss3 account root",
			path:    "awss3/account-x/",
			wantSub: "/",
		},
		{
			name:    "google account root",
			path:    "google/account-x/",
			wantSub: "/",
		},
		{
			name:    "dropbox account root",
			path:    "dropbox/account-x/",
			wantSub: "/",
		},

		// Non-root paths: leading '/' is stripped, trailing '/' is
		// preserved so BuildPlan can still see the directory hint.
		{
			name:    "drive Home subdir with trailing slash",
			path:    "drive/Home/Documents/",
			wantSub: "Documents/",
		},
		{
			name:    "drive Home subdir without trailing slash (file rename target)",
			path:    "drive/Home/Documents/2026.pdf",
			wantSub: "Documents/2026.pdf",
		},
		{
			name:    "deep subdir with trailing slash",
			path:    "drive/Home/Documents/Backups/",
			wantSub: "Documents/Backups/",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fp, err := ParseFrontendPath(tc.path)
			if err != nil {
				t.Fatalf("ParseFrontendPath(%q): %v", tc.path, err)
			}
			got := subPathForBuildPlan(fp)
			if got != tc.wantSub {
				t.Errorf("subPathForBuildPlan(%q) = %q, want %q",
					tc.path, got, tc.wantSub)
			}
		})
	}
}

// TestSubPathForBuildPlan_DirectoryUploadToRoot is the end-to-end
// regression test for the original bug report: uploading a directory
// to a namespace root failed with `local "./mydir/" is a directory;
// remote "" must end with '/'`. We exercise the full chain
// ParseFrontendPath → subPathForBuildPlan → upload.BuildPlan with
// a real directory on disk to make sure the fix is wired correctly
// (i.e. that the helper's output is what runUpload actually feeds to
// BuildPlan). Without the fix, BuildPlan would return the
// "must end with '/'" error for every namespace tested here.
func TestSubPathForBuildPlan_DirectoryUploadToRoot(t *testing.T) {
	cases := []struct {
		name      string
		path      string
		apiRoot   string
		chunkRoot string
		wantPD    string
	}{
		{
			name:      "drive Home root",
			path:      "drive/Home/",
			apiRoot:   "/drive/Home",
			chunkRoot: "/drive/Home",
			wantPD:    "/drive/Home/",
		},
		{
			name:      "drive Data root",
			path:      "drive/Data/",
			apiRoot:   "/drive/Data",
			chunkRoot: "/drive/Data",
			wantPD:    "/drive/Data/",
		},
		// Sync's chunkRoot is empty (chunks go to seafhttp/upload-aj
		// which expects an inside-repo parent_dir); the API parent_dir
		// still anchors at /sync/<repo>/.
		{
			name:      "sync repo root (chunkRoot empty by design)",
			path:      "sync/repo-abc/",
			apiRoot:   "/sync/repo-abc",
			chunkRoot: "",
			wantPD:    "/sync/repo-abc/",
		},
		{
			name:      "cache node root",
			path:      "cache/node-1/",
			apiRoot:   "/cache/node-1",
			chunkRoot: "/cache/node-1",
			wantPD:    "/cache/node-1/",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			src := filepath.Join(dir, "mydir")
			if err := os.MkdirAll(src, 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(src, "a.txt"), []byte("a"), 0o644); err != nil {
				t.Fatal(err)
			}
			fp, err := ParseFrontendPath(tc.path)
			if err != nil {
				t.Fatalf("ParseFrontendPath(%q): %v", tc.path, err)
			}
			remoteSub := subPathForBuildPlan(fp)
			plan, err := upload.BuildPlan(src, remoteSub, tc.apiRoot, tc.chunkRoot)
			if err != nil {
				t.Fatalf("BuildPlan: directory upload to %q rejected: %v",
					tc.path, err)
			}
			if plan.ParentDir != tc.wantPD {
				t.Errorf("ParentDir = %q, want %q", plan.ParentDir, tc.wantPD)
			}
			if len(plan.Files) != 1 {
				t.Fatalf("Files = %d, want 1", len(plan.Files))
			}
			// The source folder name should appear as the first
			// path component (mydir/a.txt), matching the LarePass
			// folder-upload UX where the picked folder's name is
			// preserved on the server.
			if plan.Files[0].RelativePath != "mydir/a.txt" {
				t.Errorf("Files[0].RelativePath = %q, want %q",
					plan.Files[0].RelativePath, "mydir/a.txt")
			}
		})
	}
}
