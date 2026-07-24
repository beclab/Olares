package files

import (
	"context"
	"fmt"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// commonNamespaceMinOlaresVersion is the Olares OS release that
// introduced the drive/Common app common data area (the shared volume
// backed by JuiceFS /rootfs/Common — ollama / huggingface / comfyui
// caches).
//
// It mirrors TermiPass's isCommonEnable() gate, which surfaces the
// Common namespace ONLY when the backend is new enough:
//
//	// apps/.../src/api/files/index.ts
//	const isCommonEnable = () => userStore.current_user.isLargeVersion12_6;
//
//	// packages/sdk/src/core/user.ts
//	get isLargeVersion12_6() {
//	    return compareOlaresVersion(this.os_version, '1.12.6-0').compare >= 0;
//	}
const commonNamespaceMinOlaresVersion = "1.12.6"

// archiveMinOlaresVersion is the first Olares line that exposes the
// files archive wire surface (`/api/archive/<node>/{compress,extract,entries,entry}`).
// On 1.12.5 those routes do not exist (HTTP 404), so we fail fast with
// a version error before issuing the wire call.
const archiveMinOlaresVersion = "1.12.6"

// nfsMinOlaresVersion is the first Olares line that exposes the NFS
// mount flow (`external_type=nfs` on /api/mount + /api/unmount).
// Pre-1.12.6 backends don't support these semantics.
const nfsMinOlaresVersion = "1.12.6"

// isCommonFrontendPath reports whether a parsed (fileType, extend)
// pair targets the drive/Common namespace. Centralised so the version
// gate and any future Common-specific branching agree on the shape.
func isCommonFrontendPath(fileType, extend string) bool {
	return fileType == "drive" && extend == "Common"
}

// requireCommonBackendVersion is the client-side version preflight for
// drive/Common. When the operation touches the Common namespace
// (touchesCommon == true) it requires the target Olares backend to be
// >= 1.12.6 — the release that introduced the common data area. On an
// older (or undetectable) backend the namespace does not exist, so we
// reject up front with an actionable error instead of letting the user
// hit an opaque server 404 / 500.
//
// It is a no-op when touchesCommon is false, so callers wire it in
// unconditionally right after they parse their path(s).
//
// Fail-closed semantics mirror the LarePass GUI, which hides Common
// until it positively knows the backend is new enough:
//
//   - version detected and >= 1.12.6   → allowed
//   - version detected but < 1.12.6    → rejected (suggest upgrade)
//   - version undetectable (offline,
//     no cache)                       → rejected (suggest profile refresh)
//
// In the common case the version was cached eagerly at `profile login`, so
// this gate adds no extra request.
func requireCommonBackendVersion(ctx context.Context, f *cmdutil.Factory, touchesCommon bool) error {
	if !touchesCommon {
		return nil
	}
	ok, err := f.OlaresBackendAtLeast(ctx, commonNamespaceMinOlaresVersion)
	if err != nil {
		return fmt.Errorf(
			"drive/Common (the app common data area) requires Olares >= %s, but the backend "+
				"version could not be determined: %v",
			commonNamespaceMinOlaresVersion, err)
	}
	if !ok {
		got := "unknown"
		if v, verr := f.OlaresBackendVersion(ctx); verr == nil && v != nil {
			got = v.Original()
		}
		return fmt.Errorf(
			"drive/Common (the app common data area) requires Olares >= %s, but this backend is %s; "+
				"upgrade the Olares system, or operate on drive/Home or drive/Data instead",
			commonNamespaceMinOlaresVersion, got)
	}
	return nil
}

// requireArchiveBackendVersion is the feature gate for the files archive
// surface (compress / extract / archive entries / archive cat). The backend
// routes were introduced in 1.12.6, so on 1.12.5 we fail early with an
// actionable upgrade message instead of surfacing a confusing HTTP 404 from
// `/api/archive/...`.
func requireArchiveBackendVersion(ctx context.Context, f *cmdutil.Factory) error {
	ok, err := f.OlaresBackendAtLeast(ctx, archiveMinOlaresVersion)
	if err != nil {
		return fmt.Errorf(
			"`files compress` / `files extract` / `files archive` require Olares >= %s (archive APIs), but the backend "+
				"version could not be determined: %v",
			archiveMinOlaresVersion, err)
	}
	if !ok {
		got := "unknown"
		if v, verr := f.OlaresBackendVersion(ctx); verr == nil && v != nil {
			got = v.Original()
		}
		return fmt.Errorf(
			"`files compress` / `files extract` / `files archive` require Olares >= %s (archive APIs), but this backend is %s; "+
				"upgrade the Olares system before using archive commands",
			archiveMinOlaresVersion, got)
	}
	return nil
}

// requireNFSBackendVersion is the feature gate for `files nfs`.
// NFS mount/history semantics were introduced in 1.12.6.
func requireNFSBackendVersion(ctx context.Context, f *cmdutil.Factory) error {
	ok, err := f.OlaresBackendAtLeast(ctx, nfsMinOlaresVersion)
	if err != nil {
		return fmt.Errorf(
			"`files nfs` requires Olares >= %s, but the backend version could not be determined: %v",
			nfsMinOlaresVersion, err)
	}
	if !ok {
		got := "unknown"
		if v, verr := f.OlaresBackendVersion(ctx); verr == nil && v != nil {
			got = v.Original()
		}
		return fmt.Errorf(
			"`files nfs` requires Olares >= %s, but this backend is %s; upgrade the Olares system to use NFS mount commands",
			nfsMinOlaresVersion, got)
	}
	return nil
}
