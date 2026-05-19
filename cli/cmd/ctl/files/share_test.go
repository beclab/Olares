package files

import (
	"strings"
	"testing"

	"github.com/beclab/Olares/cli/internal/files/share"
)

// TestValidateShareNamespace_RejectsCloudForAllFlavors pins the
// uniform cloud-rejection rule: every share-create flavor
// (internal / public / smb) must reject every cloud namespace
// (awss3 / google / dropbox / tencent) up front, citing
// cross-cloud-account semantics rather than the per-flavor allow-
// list. Mirrors the user requirement "所有和分享相关功能对 cloud
// 都不支持" (all share-related operations unsupported for cloud).
//
// The error message is checked for two anchors:
//
//   - the friendly flavor name ("internal" / "public" / "smb") —
//     NOT the wire value (where Public is the historically
//     confusing "external"), so the user sees the same word they
//     typed at the CLI;
//
//   - "cloud namespaces" + the recovery hint pointing at
//     `files download` followed by re-upload to drive.
//
// Anchoring on both halves stops a future refactor from quietly
// dropping the recovery suggestion or merging the message into the
// generic per-flavor branch.
func TestValidateShareNamespace_RejectsCloudForAllFlavors(t *testing.T) {
	flavors := []struct {
		flavor share.Type
		// expected friendly name in the rendered error message
		friendly string
	}{
		{share.TypeInternal, "internal"},
		{share.TypePublic, "public"},
		{share.TypeSMB, "smb"},
	}
	cloudFTs := []string{"awss3", "google", "dropbox", "tencent"}
	for _, fl := range flavors {
		for _, ft := range cloudFTs {
			t.Run(fl.friendly+"/"+ft, func(t *testing.T) {
				display := ft + "/account-1/Photos/"
				err := validateShareNamespace(fl.flavor, ft, display)
				if err == nil {
					t.Fatalf("validateShareNamespace(%v, %q, %q): expected refusal",
						fl.flavor, ft, display)
				}
				msg := err.Error()
				for _, want := range []string{
					"refusing to create a " + fl.friendly + " share for " + display,
					"cloud namespaces",
					"awss3 / google / dropbox / tencent",
					"`files download`",
					"drive/Home or drive/Data",
				} {
					if !strings.Contains(msg, want) {
						t.Errorf("error must contain %q; got: %v", want, err)
					}
				}
			})
		}
	}
}

// TestValidateShareNamespace_PublicAllowsOnlyDrive pins the Public
// flavor's tightening: only the `drive` namespace passes; sync,
// external, cache, and every cloud namespace are refused. The
// non-cloud branch is exercised here for sync / external / cache
// (cloud already covered by the cloud-specific test above).
//
// The recovery hint is dynamically built from the OTHER flavors'
// allow-lists — so it depends on which flavors currently accept
// each fileType:
//
//   - sync       → only Internal accepts it (SMB excluded sync to
//                  match the LarePass GUI), so the hint points at
//                  `files share internal` only.
//   - external   → both Internal and SMB accept it, hint mentions
//                  both.
//   - cache      → both Internal and SMB accept it, hint mentions
//                  both.
//
// mustNotContain pins the absences too: sync's hint must NOT
// mention `files share smb` (regression guard if SMB ever
// accidentally re-allows sync).
func TestValidateShareNamespace_PublicAllowsOnlyDrive(t *testing.T) {
	allowed := []string{"drive"}
	for _, ft := range allowed {
		t.Run("allow "+ft, func(t *testing.T) {
			if err := validateShareNamespace(share.TypePublic, ft, ft+"/Home/x/"); err != nil {
				t.Errorf("validateShareNamespace(public, %q, ...): unexpected error: %v", ft, err)
			}
		})
	}

	rejected := []struct {
		fileType       string
		mustContain    []string
		mustNotContain []string
	}{
		{
			fileType: "sync",
			mustContain: []string{
				"refusing to create a public share for sync/abc-repo/",
				"`files share public` only supports the {drive} namespace(s)",
				"`files share internal`",
			},
			mustNotContain: []string{
				// SMB excludes sync (matches the LarePass GUI),
				// so the auto-generated fallback hint must NOT
				// suggest it.
				"`files share smb`",
			},
		},
		{
			fileType: "external",
			mustContain: []string{
				"refusing to create a public share for external/node-1/hdd1/",
				"`files share public` only supports the {drive} namespace(s)",
				"`files share internal`",
				"`files share smb`",
			},
		},
		{
			fileType: "cache",
			mustContain: []string{
				"refusing to create a public share for cache/node-1/app1/",
				"`files share public` only supports the {drive} namespace(s)",
				"`files share internal`",
				"`files share smb`",
			},
		},
	}
	displays := map[string]string{
		"sync":     "sync/abc-repo/",
		"external": "external/node-1/hdd1/",
		"cache":    "cache/node-1/app1/",
	}
	for _, c := range rejected {
		t.Run("reject "+c.fileType, func(t *testing.T) {
			err := validateShareNamespace(share.TypePublic, c.fileType, displays[c.fileType])
			if err == nil {
				t.Fatalf("validateShareNamespace(public, %q, ...): expected refusal", c.fileType)
			}
			msg := err.Error()
			for _, want := range c.mustContain {
				if !strings.Contains(msg, want) {
					t.Errorf("error must contain %q; got: %v", want, err)
				}
			}
			for _, banned := range c.mustNotContain {
				if strings.Contains(msg, banned) {
					t.Errorf("error must NOT contain %q; got: %v", banned, err)
				}
			}
		})
	}
}

// TestValidateShareNamespace_InternalAllowsNonCloud locks in the
// Internal flavor's white-list: drive / sync / external / cache
// pass, every cloud namespace is rejected (covered separately by
// the cloud-specific test).
func TestValidateShareNamespace_InternalAllowsNonCloud(t *testing.T) {
	for _, ft := range []string{"drive", "sync", "external", "cache"} {
		t.Run("allow "+ft, func(t *testing.T) {
			if err := validateShareNamespace(share.TypeInternal, ft, ft+"/x/y/"); err != nil {
				t.Errorf("validateShareNamespace(internal, %q, ...): unexpected error: %v", ft, err)
			}
		})
	}
}

// TestValidateShareNamespace_SMB pins the SMB flavor's allow-list
// against the LarePass GUI exactly: drive / external / cache pass,
// sync is REJECTED.
//
// The sync rejection is the one that diverges between the three
// flavors — Internal allows it, SMB doesn't, Public doesn't — so
// the test asserts both the rejection AND the recovery hint
// pointing at `files share internal` (the only remaining flavor
// that accepts sync). If a future change accidentally re-allows
// sync for SMB, the rejection assertion fires; if it accidentally
// drops sync from Internal too, the hint assertion fires.
func TestValidateShareNamespace_SMB(t *testing.T) {
	for _, ft := range []string{"drive", "external", "cache"} {
		t.Run("allow "+ft, func(t *testing.T) {
			if err := validateShareNamespace(share.TypeSMB, ft, ft+"/x/y/"); err != nil {
				t.Errorf("validateShareNamespace(smb, %q, ...): unexpected error: %v", ft, err)
			}
		})
	}

	t.Run("reject sync", func(t *testing.T) {
		err := validateShareNamespace(share.TypeSMB, "sync", "sync/abc-repo/")
		if err == nil {
			t.Fatal("validateShareNamespace(smb, sync, ...): expected refusal")
		}
		msg := err.Error()
		for _, want := range []string{
			"refusing to create a smb share for sync/abc-repo/",
			"`files share smb` only supports the {cache, drive, external} namespace(s)",
			"`files share internal`",
		} {
			if !strings.Contains(msg, want) {
				t.Errorf("error must contain %q; got: %v", want, err)
			}
		}
		// Cloud framing must NOT leak into this branch — sync is
		// not cloud, and conflating the two would mis-suggest the
		// download/re-upload recovery flow.
		if strings.Contains(msg, "cloud namespaces") {
			t.Errorf("non-cloud rejection must not mention 'cloud namespaces'; got: %v", err)
		}
	})
}

// TestValidateShareNamespace_UnknownFlavorIsDefensive guards the
// defense-in-depth branch in validateShareNamespace: passing a
// share.Type the cobra layer doesn't construct must surface a
// typed error rather than silently allowing the share. If a future
// share-create verb forgets to thread its flavor through, this
// test fails immediately.
func TestValidateShareNamespace_UnknownFlavorIsDefensive(t *testing.T) {
	err := validateShareNamespace(share.Type("unknown-flavor"), "drive", "drive/Home/")
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
			name:   "internal: trailing slash",
			path:   "external/node-1/",
			flavor: share.TypeInternal,
		},
		{
			// Bare <fileType>/<extend> synthesizes SubPath="/", so
			// the rejection must trip the same way as the trailing-
			// slash form.
			name:   "internal: no trailing slash",
			path:   "external/node-1",
			flavor: share.TypeInternal,
		},
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

// TestFrontendPathToShareTarget_RejectsCacheNodeRoot pins the
// share-side fast-fail for `share internal|smb cache/<node>/`. As
// with the external counterpart, the Public flavor never reaches
// this branch — its namespace gate refuses the entire `cache`
// namespace earlier.
//
// Unlike external (which is fully virtual), cache HAS a real
// per-node filesystem on the wire — but the LarePass web app's
// /Files/Cache/ root view is rendered by formatAppDataNode as a
// node picker (apps/.../api/files/v2/cache/data.ts), so a "share
// this row" operation at that level points at a node selector,
// not a dataset. The CLI's share-create path mirrors that UX and
// refuses up front.
func TestFrontendPathToShareTarget_RejectsCacheNodeRoot(t *testing.T) {
	cases := []struct {
		name   string
		path   string
		flavor share.Type
	}{
		{
			name:   "internal: trailing slash",
			path:   "cache/node-1/",
			flavor: share.TypeInternal,
		},
		{
			name:   "internal: no trailing slash",
			path:   "cache/node-1",
			flavor: share.TypeInternal,
		},
		{
			name:   "smb: trailing slash",
			path:   "cache/node-1/",
			flavor: share.TypeSMB,
		},
		{
			name:   "smb: no trailing slash",
			path:   "cache/node-1",
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
			if !strings.Contains(msg, "node-picker layer") {
				t.Errorf("error should mention 'node-picker layer'; got: %v", err)
			}
			if !strings.Contains(msg, "cache/node-1/<sub>/") {
				t.Errorf("error should suggest corrected shape; got: %v", err)
			}
			if !strings.Contains(msg, "files ls cache/node-1/") {
				t.Errorf("error should hint at the discovery command; got: %v", err)
			}
		})
	}
}

// TestFrontendPathToShareTarget_PublicShortCircuitsBeforeRootCheck
// confirms ordering: for Public against external/<node>/ or
// cache/<node>/, the namespace-level rejection (the broader and
// more actionable error) must surface BEFORE the volume-listing /
// node-picker root rejection. Same final answer either way (the
// share isn't created), but the user sees the more useful "Public
// only supports drive" message instead of the narrower "this is
// the volume listing layer" one.
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
			if !strings.Contains(msg, "`files share public` only supports the {drive} namespace(s)") {
				t.Errorf("expected namespace-level rejection; got: %v", err)
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
// Cases use the most permissive flavor that accepts the path's
// fileType — Internal accepts every non-cloud namespace, so it's
// the right anchor here. Per-flavor coverage of allow / reject is
// already done by TestValidateShareNamespace_*.
func TestFrontendPathToShareTarget_AllowsRealPaths(t *testing.T) {
	cases := []struct {
		name   string
		path   string
		flavor share.Type
	}{
		{
			name:   "internal: external volume root",
			path:   "external/node-1/hdd1/",
			flavor: share.TypeInternal,
		},
		{
			name:   "internal: external nested dir",
			path:   "external/node-1/hdd1/Movies/",
			flavor: share.TypeInternal,
		},
		{
			name:   "internal: cache one segment past node",
			path:   "cache/node-1/app1/",
			flavor: share.TypeInternal,
		},
		{
			name:   "internal: cache nested file",
			path:   "cache/node-1/app1/data/cache.bin",
			flavor: share.TypeInternal,
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
