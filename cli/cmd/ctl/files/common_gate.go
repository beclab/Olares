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
//     no cache, no --olares-version)   → rejected (suggest --olares-version)
//
// The --olares-version flag (cmdutil.FlagOlaresVersion) is the escape
// hatch for all three — it sets the version explicitly with no network
// round-trip. In the common case the version was cached eagerly at
// `profile login`, so this gate adds no extra request.
func requireCommonBackendVersion(ctx context.Context, f *cmdutil.Factory, touchesCommon bool) error {
	if !touchesCommon {
		return nil
	}
	ok, err := f.OlaresBackendAtLeast(ctx, commonNamespaceMinOlaresVersion)
	if err != nil {
		return fmt.Errorf(
			"drive/Common (the app common data area) requires Olares >= %s, but the backend "+
				"version could not be determined: %v; pass --%s <version> to set it manually "+
				"(e.g. --%s %s)",
			commonNamespaceMinOlaresVersion, err,
			cmdutil.FlagOlaresVersion, cmdutil.FlagOlaresVersion, commonNamespaceMinOlaresVersion)
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
