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
// no defined semantics. The error message has evolved to omit the
// literal word "directory" and instead phrases the rule as
// "must end with '/' when more than one source is given (got N
// sources)", so the assertion below pins those two substrings (the
// trailing-slash requirement marker and the source-count fragment)
// which together uniquely identify this guard.
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
	msg := err.Error()
	if !strings.Contains(msg, "must end with '/'") {
		t.Errorf("error should mention the trailing-slash requirement, got: %v", err)
	}
	if !strings.Contains(msg, "more than one source") {
		t.Errorf("error should mention the multi-source condition, got: %v", err)
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

// TestPlan_RefusesExternalNodeRootDestination guards the
// volume-listing-layer destination (`external/<node>/`). The
// per-user files-backend exposes attached volumes (hdd1 / usb1 /
// smb-...) as virtual children of this layer, so a paste landing
// here either fails server-side or auto-renames against a
// non-existent target. The error must point the user at the
// `external/<node>/<volume>/<sub>` shape so the next invocation
// works without trial and error.
func TestPlan_RefusesExternalNodeRootDestination(t *testing.T) {
	srcs := []Source{
		{FileType: "drive", Extend: "Home", SubPath: "/foo.pdf"},
	}
	dst := Destination{FileType: "external", Extend: "node-1", SubPath: "/", IsDirIntent: true}

	_, err := Plan(srcs, dst, ActionCopy, false, "node-default", "")
	if err == nil {
		t.Fatal("expected error for external/<node>/ destination")
	}
	if !strings.Contains(err.Error(), "volume listing layer") {
		t.Errorf("error should mention 'volume listing layer', got: %v", err)
	}
	if !strings.Contains(err.Error(), "external/node-1/<volume>/<sub>/") {
		t.Errorf("error should suggest the corrected shape, got: %v", err)
	}

	// Same rule applies to mv (action="move"). Anchoring both
	// actions in one helper test keeps the constraint visible
	// for both verbs without doubling the test surface.
	_, errMv := Plan(srcs, dst, ActionMove, false, "node-default", "")
	if errMv == nil {
		t.Fatal("expected error for external/<node>/ destination on mv")
	}
}

// TestPlan_RefusesProtectedDriveHomeChildOnMove pins the
// LarePass-aligned policy that the system-managed first-level
// children directly under drive/Home/ (Pictures / Music / Movies /
// Downloads / Documents / Code / Cache / Data / Home / Ollama /
// Huggingface) refuse `mv` as the source — the GUI's
// `disableMenuItem` array in
// apps/packages/app/src/stores/operation.ts greys out cut/move when
// the user is at /Files/Home/, and a CLI mv would silently produce a
// state the GUI cannot reach (and would unlink bootstrap dirs that
// user apps assume exist).
//
// The matrix below covers all 11 LarePass-protected names plus a
// trailing-slash variant, and ensures the SAME sources stay valid
// for `cp` (copy) — the policy is intentionally narrower than the
// GUI for actions that don't unlink the source.
func TestPlan_RefusesProtectedDriveHomeChildOnMove(t *testing.T) {
	rejectCases := []struct {
		name string
		sub  string
		dir  bool
	}{
		{"Pictures", "/Pictures", true},
		{"Pictures with trailing slash", "/Pictures/", true},
		{"Music", "/Music", true},
		{"Movies", "/Movies", true},
		{"Downloads", "/Downloads", true},
		{"Documents", "/Documents", true},
		{"Code", "/Code", true},
		{"Cache", "/Cache", true},
		{"Data", "/Data", true},
		{"Home nested", "/Home", true},
		{"Ollama", "/Ollama", true},
		{"Huggingface one-word", "/Huggingface", true},
	}
	dst := Destination{FileType: "drive", Extend: "Home", SubPath: "/Backups/", IsDirIntent: true}

	for _, c := range rejectCases {
		t.Run("reject mv "+c.name, func(t *testing.T) {
			srcs := []Source{
				{FileType: "drive", Extend: "Home", SubPath: c.sub, IsDirIntent: c.dir},
			}
			_, err := Plan(srcs, dst, ActionMove, true, "node-a", "")
			if err == nil {
				t.Fatalf("Plan: expected mv refusal for drive/Home%s", c.sub)
			}
			msg := err.Error()
			if !strings.Contains(msg, "system-managed Home folder") {
				t.Errorf("error should mention 'system-managed Home folder'; got: %v", err)
			}
			if !strings.Contains(msg, "Files GUI") {
				t.Errorf("error should reference the Files GUI for context; got: %v", err)
			}
			if !strings.Contains(msg, "Pictures") || !strings.Contains(msg, "Huggingface") {
				t.Errorf("error should enumerate protected names; got: %v", err)
			}

			// Mirror operation: cp (copy) MUST go through. Copy
			// preserves the source unchanged, so the data-loss
			// risk that justifies the move guard does not apply.
			// Anchoring the cp-allowed branch in the same test
			// keeps the cp-vs-mv asymmetry visible at the test
			// site (rather than requiring a reader to check two
			// tests).
			if _, err := Plan(srcs, dst, ActionCopy, true, "node-a", ""); err != nil {
				t.Errorf("Plan: cp must NOT be gated by the protected-children policy, got: %v", err)
			}
		})
	}

	// User-content paths and other namespaces / extends MUST stay
	// movable — the policy must not over-extend.
	allowMoveCases := []struct {
		name string
		src  Source
	}{
		{
			// Album/sub-folder under Pictures: pure user content.
			name: "deeper path under Pictures",
			src: Source{
				FileType: "drive", Extend: "Home",
				SubPath: "/Pictures/Trip2024/", IsDirIntent: true,
			},
		},
		{
			// File inside Documents.
			name: "file inside Documents",
			src: Source{
				FileType: "drive", Extend: "Home",
				SubPath: "/Documents/notes.md",
			},
		},
		{
			// drive/Data/<same-name>: different volume root, the
			// policy is Home-only.
			name: "drive Data same name",
			src: Source{
				FileType: "drive", Extend: "Data",
				SubPath: "/Pictures", IsDirIntent: true,
			},
		},
		{
			// Other namespace: out of scope.
			name: "sync repo same name",
			src: Source{
				FileType: "sync", Extend: "abc-repo",
				SubPath: "/Pictures", IsDirIntent: true,
			},
		},
		{
			// User-created folder at drive/Home/<name> not in the
			// protected list.
			name: "drive Home user folder",
			src: Source{
				FileType: "drive", Extend: "Home",
				SubPath: "/MyProjects", IsDirIntent: true,
			},
		},
	}
	for _, c := range allowMoveCases {
		t.Run("allow mv "+c.name, func(t *testing.T) {
			if _, err := Plan([]Source{c.src}, dst, ActionMove, true, "node-a", ""); err != nil {
				t.Errorf("Plan: unexpected mv refusal for %s/%s%s: %v",
					c.src.FileType, c.src.Extend, c.src.SubPath, err)
			}
		})
	}
}

// TestPlan_RefusesProtectedExternalChildOnMove pins the LarePass-
// aligned policy that the system-managed AI mountpoint folders
// under `external/<node>/...` refuse `mv` as the source: the GUI
// greys out cut / move via `externalFolderWhiteList` (depth-1: ai)
// and `externalAiFolderWhiteList` (depth-2: output / model /
// comfyui / ollama) in apps/packages/app/src/stores/operation.ts.
// Moving the source would unlink the directory from the mountpoint
// and silently break the Ollama / ComfyUI / Huggingface readers
// that look it up by name.
//
// Path-shape ground truth: the GUI's `<X>` in
// `/Files/External/<X>/` is the LarePass `masterNode`
// (apps/.../external/data.ts:77), so `<X>` maps to
// FrontendPath.Extend on the CLI — the depth-1 `ai/` entry lives
// at `external/<node>/ai/` (SubPath="/ai/"), NOT under any nested
// volume segment.
//
// Symmetric with TestPlan_RefusesProtectedDriveHomeChildOnMove:
// the SAME sources stay valid for `cp` (copy preserves the source,
// so the data-loss risk that justifies the move guard does not
// apply). The cp branch is anchored here so the cp-vs-mv asymmetry
// stays visible at the test site.
func TestPlan_RefusesProtectedExternalChildOnMove(t *testing.T) {
	rejectCases := []struct {
		name string
		src  Source
	}{
		// ----- depth-1: external/<node>/ai -----
		{
			name: "depth-1 ai under olares node",
			src: Source{
				FileType: "external", Extend: "olares",
				SubPath: "/ai", IsDirIntent: true,
			},
		},
		{
			name: "depth-1 ai with trailing slash",
			src: Source{
				FileType: "external", Extend: "olares",
				SubPath: "/ai/", IsDirIntent: true,
			},
		},
		{
			name: "depth-1 ai on arbitrary node name (node is opaque)",
			src: Source{
				FileType: "external", Extend: "node-1",
				SubPath: "/ai/", IsDirIntent: true,
			},
		},
		// ----- depth-2: external/<node>/ai/<name> -----
		{
			name: "depth-2 ai/output",
			src: Source{
				FileType: "external", Extend: "olares",
				SubPath: "/ai/output", IsDirIntent: true,
			},
		},
		{
			name: "depth-2 ai/model",
			src: Source{
				FileType: "external", Extend: "olares",
				SubPath: "/ai/model/", IsDirIntent: true,
			},
		},
		{
			name: "depth-2 ai/comfyui",
			src: Source{
				FileType: "external", Extend: "olares",
				SubPath: "/ai/comfyui", IsDirIntent: true,
			},
		},
		{
			name: "depth-2 ai/ollama",
			src: Source{
				FileType: "external", Extend: "olares",
				SubPath: "/ai/ollama/", IsDirIntent: true,
			},
		},
	}
	// Destination is a regular drive/Home subdir so it doesn't
	// preempt the source-side guard with its own validation.
	dst := Destination{FileType: "drive", Extend: "Home", SubPath: "/Backups/", IsDirIntent: true}

	for _, c := range rejectCases {
		t.Run("reject mv "+c.name, func(t *testing.T) {
			srcs := []Source{c.src}
			_, err := Plan(srcs, dst, ActionMove, true, "node-a", "")
			if err == nil {
				t.Fatalf("Plan: expected mv refusal for %s/%s%s",
					c.src.FileType, c.src.Extend, c.src.SubPath)
			}
			msg := err.Error()
			if !strings.Contains(msg, "system-managed AI mountpoint folder") {
				t.Errorf("error should mention 'system-managed AI mountpoint folder'; got: %v", err)
			}
			if !strings.Contains(msg, "LarePass") {
				t.Errorf("error should reference LarePass for context; got: %v", err)
			}
			// Error must echo the offending path so the user
			// can match it against their command line.
			displayHint := c.src.FileType + "/" + c.src.Extend
			if !strings.Contains(msg, displayHint) {
				t.Errorf("error should echo path prefix %q; got: %v", displayHint, err)
			}
			// Sanity-check that both whitelists are enumerated.
			if !strings.Contains(msg, "comfyui") || !strings.Contains(msg, "output") {
				t.Errorf("error should enumerate depth-2 whitelist (comfyui / output); got: %v", err)
			}

			// Mirror operation: cp (copy) MUST go through, same
			// as the drive/Home test above. Anchors the cp-vs-mv
			// asymmetry at the test site so it stays observable.
			if _, err := Plan(srcs, dst, ActionCopy, true, "node-a", ""); err != nil {
				t.Errorf("Plan: cp must NOT be gated by the external-protected-children policy, got: %v", err)
			}
		})
	}

	// User-content paths and other namespaces MUST stay movable —
	// the policy must not over-extend.
	allowMoveCases := []struct {
		name string
		src  Source
	}{
		{
			// Depth-2 under a non-ai depth-1 parent — name
			// happens to match the ai-whitelist but the parent
			// isn't "ai", so this is just a regular dir inside
			// some volume.
			name: "depth-2 under non-ai parent",
			src: Source{
				FileType: "external", Extend: "olares",
				SubPath: "/USB-0/output", IsDirIntent: true,
			},
		},
		{
			// Depth-2 ai/<other> not in the whitelist — user
			// content under ai/.
			name: "depth-2 ai/<other> not whitelisted",
			src: Source{
				FileType: "external", Extend: "olares",
				SubPath: "/ai/my-experiments", IsDirIntent: true,
			},
		},
		{
			// Depth-3 under ai/output — per-run cleanups
			// remain movable (only the dirs themselves are
			// pinned).
			name: "depth-3 under ai/output",
			src: Source{
				FileType: "external", Extend: "olares",
				SubPath: "/ai/output/run-2026-05", IsDirIntent: true,
			},
		},
		{
			// Case-sensitive mismatch — the GUI compares the
			// lowercase string values; CLI must too.
			name: "case mismatch on depth-1 name",
			src: Source{
				FileType: "external", Extend: "olares",
				SubPath: "/AI", IsDirIntent: true,
			},
		},
		{
			name: "case mismatch on depth-2 name",
			src: Source{
				FileType: "external", Extend: "olares",
				SubPath: "/ai/Output", IsDirIntent: true,
			},
		},
		{
			// drive/Home/ai is NOT external — different
			// namespace.
			name: "drive/Home/ai is not external",
			src: Source{
				FileType: "drive", Extend: "Home",
				SubPath: "/ai", IsDirIntent: true,
			},
		},
	}
	for _, c := range allowMoveCases {
		t.Run("allow mv "+c.name, func(t *testing.T) {
			if _, err := Plan([]Source{c.src}, dst, ActionMove, true, "node-a", ""); err != nil {
				t.Errorf("Plan: unexpected mv refusal for %s/%s%s: %v",
					c.src.FileType, c.src.Extend, c.src.SubPath, err)
			}
		})
	}
}

// TestPlan_AllowsExternalVolumeRootDestination confirms the writer
// guard is narrow: pointing at `external/<node>/<volume>/` (one
// segment past the node) DOES go through. The strict rule the user
// agreed on is `SubPath != "/"` for external destinations, so a
// volume root is a valid drop target — whether <volume> is a real
// mount is the server's call.
func TestPlan_AllowsExternalVolumeRootDestination(t *testing.T) {
	srcs := []Source{
		{FileType: "drive", Extend: "Home", SubPath: "/foo.pdf"},
	}
	dst := Destination{FileType: "external", Extend: "node-1", SubPath: "/hdd1/", IsDirIntent: true}

	ops, err := Plan(srcs, dst, ActionCopy, false, "node-default", "")
	if err != nil {
		t.Fatalf("Plan: %v (volume root should be allowed)", err)
	}
	if len(ops) != 1 || ops[0].Destination != "/external/node-1/hdd1/foo.pdf" {
		t.Errorf("Destination shape unexpected: %+v", ops)
	}
}

// TestPlan_AllowsSameFileSrcDst confirms `cp foo foo` (same wire
// path on both sides, file target) is NOT rejected client-side —
// the LarePass web app doesn't enforce this and the backend's paste
// endpoint auto-renames into `foo (1)` (same POST-mkdir quirk users
// already work with). The cycle check below is dir-only (gated by
// IsDirIntent), so a same-path FILE pair must reach Plan's tail and
// produce a normal Op.
func TestPlan_AllowsSameFileSrcDst(t *testing.T) {
	srcs := []Source{
		{FileType: "drive", Extend: "Home", SubPath: "/Documents/foo.pdf"},
	}
	dst := Destination{FileType: "drive", Extend: "Home", SubPath: "/Documents/foo.pdf"}

	ops, err := Plan(srcs, dst, ActionCopy, false, "node-a", "")
	if err != nil {
		t.Fatalf("Plan: unexpected error for same-path file copy: %v", err)
	}
	if len(ops) != 1 {
		t.Fatalf("Plan: want 1 op, got %d", len(ops))
	}
	if ops[0].Source != "/drive/Home/Documents/foo.pdf" || ops[0].Destination != "/drive/Home/Documents/foo.pdf" {
		t.Errorf("Plan: source/destination shape unexpected: %+v", ops[0])
	}
}

// TestPlan_RefusesSameDirSrcDst keeps the dir-to-same-dir case
// rejected — but via the cycle check (`destination ... is inside
// source`), not the removed `source and destination are the same`
// guard. A dir copy onto itself would create an infinitely-recursing
// tree, which the cycle check catches because dstWire == srcWire ⇒
// HasPrefix(srcWire+"/", srcWire+"/") is true.
func TestPlan_RefusesSameDirSrcDst(t *testing.T) {
	srcs := []Source{
		{FileType: "drive", Extend: "Home", SubPath: "/Documents/", IsDirIntent: true},
	}
	dst := Destination{FileType: "drive", Extend: "Home", SubPath: "/Documents/"}

	_, err := Plan(srcs, dst, ActionCopy, true, "node-a", "")
	if err == nil {
		t.Fatal("expected cycle error for dir src == dst")
	}
	if !strings.Contains(err.Error(), "cycle") {
		t.Errorf("error should mention 'cycle', got: %v", err)
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
