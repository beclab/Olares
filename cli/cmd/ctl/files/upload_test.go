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
	"strings"
	"testing"
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
