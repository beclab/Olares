package files

import (
	"reflect"
	"strings"
	"testing"

	"github.com/beclab/Olares/cli/internal/files/share"
)

// TestShareFlavorsFor pins the share flavors each storage takes. It is
// the same table as common.ShareTypesOf in the files backend and the
// shareModes of the LarePass adapters; a change here is a change there.
func TestShareFlavorsFor(t *testing.T) {
	all := []share.Type{share.TypeInternal, share.TypeSMB, share.TypePublic}
	smb := []share.Type{share.TypeSMB}
	cases := []struct {
		fileType, extend string
		want             []share.Type
	}{
		{"drive", "Home", all},
		{"drive", "Data", nil},
		{"drive", "Common", nil},
		{"cache", "node-1", nil},
		{"external", "node-1", smb},
		{"sync", "repo", []share.Type{share.TypeInternal}},
		{"awss3", "acct", nil},
		{"google", "acct", nil},
		{"dropbox", "acct", nil},
		{"tencent", "acct", nil},
	}
	for _, c := range cases {
		got := shareFlavorsFor(c.fileType, c.extend)
		if !reflect.DeepEqual(got, c.want) {
			t.Errorf("shareFlavorsFor(%q, %q) = %v, want %v", c.fileType, c.extend, got, c.want)
		}
		for _, fl := range all {
			err := validateShareNamespace(fl, c.fileType, c.extend, "x/")
			wantOK := false
			for _, w := range c.want {
				wantOK = wantOK || w == fl
			}
			if (err == nil) != wantOK {
				t.Errorf("validateShareNamespace(%v, %q, %q) = %v, want allowed=%v",
					fl, c.fileType, c.extend, err, wantOK)
			}
		}
	}
}

// TestValidateShareNamespace_RejectsCloudForAllFlavors pins the cloud
// message: every flavor refuses every cloud namespace, citing the
// download / re-upload recovery path and the flavor name the user
// typed (not the wire value, where public is "external").
func TestValidateShareNamespace_RejectsCloudForAllFlavors(t *testing.T) {
	for _, fl := range []share.Type{share.TypeInternal, share.TypePublic, share.TypeSMB} {
		for _, ft := range []string{"awss3", "google", "dropbox", "tencent"} {
			display := ft + "/account-1/Photos/"
			err := validateShareNamespace(fl, ft, "account-1", display)
			if err == nil {
				t.Fatalf("validateShareNamespace(%v, %q): expected refusal", fl, ft)
			}
			for _, want := range []string{
				"refusing to create a " + shareFlavorFriendlyName(fl) + " share for " + display,
				"cloud namespaces",
				"`files download`",
				"drive/Home",
			} {
				if !strings.Contains(err.Error(), want) {
					t.Errorf("error must contain %q; got: %v", want, err)
				}
			}
		}
	}
}

// TestValidateShareNamespace_NamesTheFlavorsThatWork pins the message
// for a storage that takes other flavors: it names them, and the
// command to use instead.
func TestValidateShareNamespace_NamesTheFlavorsThatWork(t *testing.T) {
	cases := []struct {
		flavor           share.Type
		fileType, extend string
		display          string
		want             []string
		notWant          []string
	}{
		{share.TypeInternal, "external", "node-1", "external/node-1/hdd1/",
			[]string{"external only supports smb shares", "`files share smb`"}, nil},
		{share.TypeSMB, "sync", "abc-repo", "sync/abc-repo/",
			[]string{"sync only supports internal shares", "`files share internal`"},
			[]string{"`files share smb`", "cloud namespaces"}},
	}
	for _, c := range cases {
		err := validateShareNamespace(c.flavor, c.fileType, c.extend, c.display)
		if err == nil {
			t.Fatalf("validateShareNamespace(%v, %s/%s): expected refusal", c.flavor, c.fileType, c.extend)
		}
		for _, w := range c.want {
			if !strings.Contains(err.Error(), w) {
				t.Errorf("error must contain %q; got: %v", w, err)
			}
		}
		for _, w := range c.notWant {
			if strings.Contains(err.Error(), w) {
				t.Errorf("error must not contain %q; got: %v", w, err)
			}
		}
	}
}

// TestValidateShareNamespace_NoShareStorages pins the message for the
// storages that take no share at all.
func TestValidateShareNamespace_NoShareStorages(t *testing.T) {
	for _, c := range []struct{ fileType, extend, storage string }{
		{"drive", "Data", "drive/Data"},
		{"drive", "Common", "drive/Common"},
		{"cache", "node-1", "cache"},
	} {
		for _, fl := range []share.Type{share.TypeInternal, share.TypeSMB, share.TypePublic} {
			err := validateShareNamespace(fl, c.fileType, c.extend, "x/")
			if err == nil || !strings.Contains(err.Error(), c.storage+" cannot be shared") {
				t.Errorf("validateShareNamespace(%v, %s/%s) = %v, want %q", fl, c.fileType, c.extend, err, c.storage+" cannot be shared")
			}
			if err != nil && strings.Contains(err.Error(), "cloud namespaces") {
				t.Errorf("non-cloud rejection must not mention cloud: %v", err)
			}
		}
	}
}

// TestValidateShareNamespace_UnknownFlavorIsDefensive guards the
// defense-in-depth branch in validateShareNamespace: passing a
// share.Type the cobra layer doesn't construct must surface a
// typed error rather than silently allowing the share.
func TestValidateShareNamespace_UnknownFlavorIsDefensive(t *testing.T) {
	err := validateShareNamespace(share.Type("unknown-flavor"), "drive", "Home", "drive/Home/")
	if err == nil {
		t.Fatal("expected error for unknown share flavor")
	}
	if !strings.Contains(err.Error(), "unknown share flavor") {
		t.Errorf("error must mention 'unknown share flavor'; got: %v", err)
	}
}

// TestFrontendPathToShareTarget_RejectsExternalNodeRoot pins the
// share-side fast-fail for `share internal|smb external/<node>/`
// (the volume-listing layer; see Server-side quirks #3 in the SKILL
// doc). The Public flavor never reaches this branch because its
// namespace gate refuses the entire `external` namespace earlier
// — see TestValidateShareNamespace_PublicAllowsOnlyDrive.
//
// The error message must point at the corrected shape
// (`external/<node>/<volume>/<sub>/`) AND at `files ls
// external/<node>/` as the discovery step, so the user can recover
// without round-tripping through the docs.
func TestFrontendPathToShareTarget_RejectsExternalNodeRoot(t *testing.T) {
	cases := []struct {
		name   string
		path   string
		flavor share.Type
	}{
		{
			name:   "smb: trailing slash",
			path:   "external/node-1/",
			flavor: share.TypeSMB,
		},
		{
			name:   "smb: no trailing slash",
			path:   "external/node-1",
			flavor: share.TypeSMB,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := frontendPathToShareTarget(c.path, c.flavor)
			if err == nil {
				t.Fatalf("frontendPathToShareTarget(%q, %v): expected refusal", c.path, c.flavor)
			}
			msg := err.Error()
			if !strings.Contains(msg, "volume listing layer") {
				t.Errorf("error should mention 'volume listing layer'; got: %v", err)
			}
			if !strings.Contains(msg, "external/node-1/<volume>/<sub>/") {
				t.Errorf("error should suggest corrected shape; got: %v", err)
			}
			if !strings.Contains(msg, "files ls external/node-1/") {
				t.Errorf("error should hint at the discovery command; got: %v", err)
			}
		})
	}
}

// TestFrontendPathToShareTarget_PublicShortCircuitsBeforeRootCheck
// confirms ordering: for Public against external/<node>/ or
// cache/<node>/, the storage-level rejection (the broader and more
// actionable error) must surface BEFORE the volume-listing /
// node-picker root rejection. Same final answer either way (the
// share isn't created), but the user sees the more useful "only
// supports smb shares" message instead of the narrower "this is the
// volume listing layer" one.
func TestFrontendPathToShareTarget_PublicShortCircuitsBeforeRootCheck(t *testing.T) {
	cases := []struct {
		name string
		path string
	}{
		{name: "external root", path: "external/node-1/"},
		{name: "cache root", path: "cache/node-1/"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := frontendPathToShareTarget(c.path, share.TypePublic)
			if err == nil {
				t.Fatalf("frontendPathToShareTarget(%q, public): expected refusal", c.path)
			}
			msg := err.Error()
			if !strings.Contains(msg, "only supports smb shares") && !strings.Contains(msg, "cannot be shared") {
				t.Errorf("expected storage-level rejection; got: %v", err)
			}
			if strings.Contains(msg, "volume listing layer") || strings.Contains(msg, "node-picker layer") {
				t.Errorf("namespace gate should short-circuit before the root-layer message; got: %v", err)
			}
		})
	}
}

// TestFrontendPathToShareTarget_AllowsRealPaths confirms the share
// guards stay narrow: deeper paths under external / cache (which
// resolve to real filesystems on the wire) and unrelated namespaces
// pass through cleanly for the flavors that white-list them.
// Otherwise users would lose legitimate share workflows when those
// workflows are exactly what we want to keep.
//
// Each case uses a flavor its storage accepts. Per-storage coverage
// of allow / reject is done by TestShareFlavorsFor.
func TestFrontendPathToShareTarget_AllowsRealPaths(t *testing.T) {
	cases := []struct {
		name   string
		path   string
		flavor share.Type
	}{
		{
			name:   "smb: external volume root",
			path:   "external/node-1/hdd1/",
			flavor: share.TypeSMB,
		},
		{
			name:   "smb: external nested dir",
			path:   "external/node-1/hdd1/Movies/",
			flavor: share.TypeSMB,
		},
		{
			name:   "public: drive Home root",
			path:   "drive/Home/",
			flavor: share.TypePublic,
		},
		{
			name:   "public: drive Home subdir",
			path:   "drive/Home/Documents/",
			flavor: share.TypePublic,
		},
		{
			// Sync is allowed for Internal only — SMB excludes it
			// to match the LarePass GUI, Public is locked to
			// drive. Use Internal as the anchor here; SMB+sync
			// rejection has its own dedicated test.
			name:   "internal: sync repo root",
			path:   "sync/abc-repo/",
			flavor: share.TypeInternal,
		},
		{
			name:   "internal: sync repo subdir",
			path:   "sync/abc-repo/sub/",
			flavor: share.TypeInternal,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			tgt, err := frontendPathToShareTarget(c.path, c.flavor)
			if err != nil {
				t.Fatalf("frontendPathToShareTarget(%q, %v): unexpected error: %v",
					c.path, c.flavor, err)
			}
			if tgt.FileType == "" || tgt.Extend == "" {
				t.Errorf("Target produced for %q has empty fields: %+v", c.path, tgt)
			}
		})
	}
}

// TestFrontendPathToShareTarget_CommonCannotBeShared pins that
// drive/Common takes no share, as the LarePass GUI and the files backend
// allow: every flavor is refused before any request.
func TestFrontendPathToShareTarget_CommonCannotBeShared(t *testing.T) {
	for _, fl := range []share.Type{share.TypeInternal, share.TypeSMB, share.TypePublic} {
		_, err := frontendPathToShareTarget("drive/Common/ollama/", fl)
		if err == nil || !strings.Contains(err.Error(), "drive/Common cannot be shared") {
			t.Errorf("frontendPathToShareTarget(drive/Common, %v) = %v, want a refusal", fl, err)
		}
	}
}

// TestShareFlavorFriendlyName covers the wire ↔ CLI-verb name
// translation. The Public flavor's wire value is the historically
// confusing `"external"` discriminator, and getting that wrong in
// an error message would be very misleading — pinning the mapping
// here ensures every error built by validateShareNamespace uses
// the verb the user actually typed.
func TestShareFlavorFriendlyName(t *testing.T) {
	cases := []struct {
		flavor share.Type
		want   string
	}{
		{share.TypeInternal, "internal"},
		{share.TypePublic, "public"},
		{share.TypeSMB, "smb"},
		// Defense-in-depth fallback: an unknown flavor stringifies
		// to its wire value rather than panicking.
		{share.Type("future"), "future"},
	}
	for _, c := range cases {
		t.Run(string(c.flavor), func(t *testing.T) {
			got := shareFlavorFriendlyName(c.flavor)
			if got != c.want {
				t.Errorf("shareFlavorFriendlyName(%q) = %q, want %q",
					c.flavor, got, c.want)
			}
		})
	}
}

// TestShareNameFromPath_SyncRepoNameSubstitution pins the share-
// label derivation rules with a special case for sync repo roots:
//
//   - Non-empty subpath: last segment, regardless of namespace.
//   - Empty subpath + sync namespace + a resolved repo name:
//     prefer the repo name over the bare repo_id (UUID), so the
//     `name` written to the wire is human-readable.
//   - Empty subpath + non-sync OR no resolved name: fall through
//     to the legacy "extend as name" behavior.
//
// Without this special case, sharing `sync/<repo-uuid>/` would
// label the share with the UUID — the user would see that bare
// UUID in `share list`, in the LarePass app's share view, and in
// recipients' Shared-with-me lists.
func TestShareNameFromPath_SyncRepoNameSubstitution(t *testing.T) {
	cases := []struct {
		name         string
		target       share.Target
		syncRepoName string
		want         string
	}{
		{
			// Sync repo root WITH a resolved name: the human label
			// wins over the UUID.
			name:         "sync repo root with name",
			target:       share.Target{FileType: "sync", Extend: "abc-123-uuid", SubPath: "/"},
			syncRepoName: "Project Alpha",
			want:         "Project Alpha",
		},
		{
			// Sync repo root WITHOUT a resolved name: fall back to
			// the UUID (better than empty; the share record on
			// the wire requires a non-empty `name`).
			name:         "sync repo root without name",
			target:       share.Target{FileType: "sync", Extend: "abc-123-uuid", SubPath: "/"},
			syncRepoName: "",
			want:         "abc-123-uuid",
		},
		{
			// Non-empty subpath under sync: last segment wins,
			// the repo name is irrelevant. (The sub-folder's name
			// is more descriptive than the library's name when
			// the share targets a specific folder inside.)
			name:         "sync subdir ignores repo name",
			target:       share.Target{FileType: "sync", Extend: "abc-123-uuid", SubPath: "/Reports/Q1/"},
			syncRepoName: "Project Alpha",
			want:         "Q1",
		},
		{
			// Non-sync namespaces are unaffected by syncRepoName,
			// even if a (irrelevant) name happens to be passed in.
			name:         "drive root with stray sync name",
			target:       share.Target{FileType: "drive", Extend: "Home", SubPath: "/"},
			syncRepoName: "should-not-leak",
			want:         "Home",
		},
		{
			name:         "drive subdir",
			target:       share.Target{FileType: "drive", Extend: "Home", SubPath: "/Documents/Q1.pdf"},
			syncRepoName: "",
			want:         "Q1.pdf",
		},
		{
			// Single segment without trailing slash — file intent
			// in the trailing-slash convention, but the label
			// derivation is identical (last segment wins).
			name:         "drive single segment",
			target:       share.Target{FileType: "drive", Extend: "Home", SubPath: "/file.txt"},
			syncRepoName: "",
			want:         "file.txt",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := shareNameFromPath(c.target, c.syncRepoName)
			if got != c.want {
				t.Errorf("shareNameFromPath(%+v, %q) = %q, want %q",
					c.target, c.syncRepoName, got, c.want)
			}
		})
	}
}

// TestFormatSharePathLine covers the list / get / update display
// path renderer. The interesting cases are the sync ones: server-
// supplied SyncRepoName, caller-override (used by get / update
// after a one-shot lookupSyncRepoName fallback), and the bare-
// UUID fallback when nothing resolved.
//
// Anchors that MUST hold for every sync record with a resolved
// name:
//
//   - The path leads with `<fileType>/<repo-name><Path>` so the
//     human label appears where users scan first.
//   - The repo_id is preserved in `(repo <id>)` so cross-references
//     to `repos rename` / `repos rm` still work.
//
// Non-sync namespaces must render identically to the legacy
// `<fileType>/<extend><Path>` shape (no parens, no name swap) —
// otherwise we'd break every drive / sync (other?) caller's
// output format.
func TestFormatSharePathLine(t *testing.T) {
	cases := []struct {
		name     string
		result   *share.Result
		override string
		want     string
	}{
		{
			name: "sync with server SyncRepoName",
			result: &share.Result{
				FileType: "sync", Extend: "abc-123", Path: "/",
				SyncRepoName: "Project Alpha",
			},
			want: "sync/Project Alpha/  (repo abc-123)",
		},
		{
			name: "sync with override (server field empty)",
			result: &share.Result{
				FileType: "sync", Extend: "abc-123", Path: "/Reports/",
			},
			override: "Project Alpha",
			want:     "sync/Project Alpha/Reports/  (repo abc-123)",
		},
		{
			name: "override beats server field",
			result: &share.Result{
				FileType: "sync", Extend: "abc-123", Path: "/",
				SyncRepoName: "Old Name",
			},
			override: "New Name",
			want:     "sync/New Name/  (repo abc-123)",
		},
		{
			name: "sync with neither server nor override",
			result: &share.Result{
				FileType: "sync", Extend: "abc-123", Path: "/",
			},
			want: "sync/abc-123/",
		},
		{
			name: "drive: sync renderer is a no-op",
			result: &share.Result{
				FileType: "drive", Extend: "Home", Path: "/Documents/",
				// Even if the server wrongly populated this for a
				// drive record, the renderer must NOT use it.
				SyncRepoName: "should-not-leak",
			},
			override: "should-not-leak-either",
			want:     "drive/Home/Documents/",
		},
		{
			name:   "nil record",
			result: nil,
			want:   "",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := formatSharePathLine(c.result, c.override)
			if got != c.want {
				t.Errorf("formatSharePathLine = %q, want %q", got, c.want)
			}
		})
	}
}

// TestShareTargetDisplay mirrors the result-side renderer for the
// Target shape (what the create-flavor commands hold BEFORE the
// share record exists). Keeping these in sync is important — the
// "created share" output should read identically to the matching
// `share get` / list row, otherwise users get confused about
// whether the path they typed and the path the server stored are
// the same thing.
func TestShareTargetDisplay(t *testing.T) {
	cases := []struct {
		name         string
		target       share.Target
		syncRepoName string
		want         string
	}{
		{
			name:         "sync root with name",
			target:       share.Target{FileType: "sync", Extend: "abc-123", SubPath: "/"},
			syncRepoName: "Project Alpha",
			want:         "sync/Project Alpha/  (repo abc-123)",
		},
		{
			name:         "sync subdir with name",
			target:       share.Target{FileType: "sync", Extend: "abc-123", SubPath: "/Reports/Q1/"},
			syncRepoName: "Project Alpha",
			want:         "sync/Project Alpha/Reports/Q1/  (repo abc-123)",
		},
		{
			name:         "sync root without name",
			target:       share.Target{FileType: "sync", Extend: "abc-123", SubPath: "/"},
			syncRepoName: "",
			want:         "sync/abc-123/",
		},
		{
			name:         "drive: sync helper is no-op",
			target:       share.Target{FileType: "drive", Extend: "Home", SubPath: "/Documents/"},
			syncRepoName: "should-not-leak",
			want:         "drive/Home/Documents/",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := shareTargetDisplay(c.target, c.syncRepoName)
			if got != c.want {
				t.Errorf("shareTargetDisplay = %q, want %q", got, c.want)
			}
		})
	}
}
