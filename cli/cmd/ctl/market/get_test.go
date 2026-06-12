package market

import (
	"testing"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// TestMarketGetDoesNotExposeNoHeaders pins the contract that
// `olares-cli market get` does NOT accept --no-headers. The flag was
// briefly registered here (and documented in --help) but the runner
// always called printAppDetail unconditionally — every "Name: …" /
// "Title: …" / ... line was hard-coded, so the flag silently did
// nothing in scripts. `get` renders a key:value detail layout (one
// record, fields prefixed by their label) closer to `kubectl
// describe` than a row-oriented table; there are no "headers"
// separable from values. Use -o json for machine-readable output.
//
// This test guards against the flag ever being re-added without a
// real implementation — if a future change wires addNoHeadersFlag
// back onto `get`, the test fails and the author has to either drop
// the registration or actually teach printAppDetail to honor it.
func TestMarketGetDoesNotExposeNoHeaders(t *testing.T) {
	f := &cmdutil.Factory{}
	cmd := NewCmdMarketGet(f)
	if flag := cmd.Flags().Lookup("no-headers"); flag != nil {
		t.Fatalf("`market get` must NOT register --no-headers (registered: %+v); see TestMarketGetDoesNotExposeNoHeaders doc for why", flag)
	}
}

// TestResolveI18nFieldDoesNotMutateCaller locks down the slice-aliasing fix in
// resolveI18nField: append(path[:len(path)-1], "i18n") used to write "i18n"
// into path[len(path)-1] whenever the caller's variadic slice had spare
// capacity, silently corrupting the caller's data.
func TestResolveI18nFieldDoesNotMutateCaller(t *testing.T) {
	t.Run("happy path returns localized value", func(t *testing.T) {
		// Mirror the call shape used in printAppDetail: callers pass
		// "i18n" as the second-to-last segment, so resolveI18nField
		// strips the trailing field name and re-appends "i18n",
		// landing on m.app_info.app_entry.i18n.i18n.<locale>.metadata.<field>.
		m := map[string]interface{}{
			"app_info": map[string]interface{}{
				"app_entry": map[string]interface{}{
					"i18n": map[string]interface{}{
						"i18n": map[string]interface{}{
							"en-US": map[string]interface{}{
								"metadata": map[string]interface{}{
									"title": "Firefox",
								},
							},
						},
					},
				},
			},
		}
		got := resolveI18nField(m, "app_info", "app_entry", "i18n", "title")
		if got != "Firefox" {
			t.Fatalf("resolveI18nField = %q, want %q", got, "Firefox")
		}
	})

	t.Run("does not mutate caller's variadic slice", func(t *testing.T) {
		// Build a slice with explicit spare capacity. Without the fix,
		// append(path[:len(path)-1], "i18n") would land in the
		// final slot and overwrite "title" with "i18n".
		path := make([]string, 4, 8)
		path[0], path[1], path[2], path[3] = "app_info", "app_entry", "i18n", "title"

		m := map[string]interface{}{} // i18n map missing → resolveI18nField returns ""
		_ = resolveI18nField(m, path...)

		want := []string{"app_info", "app_entry", "i18n", "title"}
		for i, w := range want {
			if path[i] != w {
				t.Fatalf("path[%d] = %q after call, want %q (caller's slice was mutated)",
					i, path[i], w)
			}
		}
	})
}
