package appcfg

import (
	"testing"

	"github.com/beclab/Olares/framework/app-service/pkg/constants"
	v1alpha1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TestApplicationConfigIsShared pins the (apiVersion, options.shared)
// truth table for the runtime config helper. The semantic intent —
// "shared cluster-wide app" — is true only for v3 + Shared:true. Every
// other combination, including v1 manifests that happen to set
// Shared:true (which is silently ignored to preserve the v1 install
// branch verbatim), must be false.
func TestApplicationConfigIsShared(t *testing.T) {
	cases := []struct {
		name       string
		apiVersion APIVersion
		shared     bool
		want       bool
	}{
		{name: "v1 + shared=false", apiVersion: V1, shared: false, want: false},
		{name: "v1 + shared=true (ignored)", apiVersion: V1, shared: true, want: false},
		{name: "v2 + shared=false", apiVersion: V2, shared: false, want: false},
		{name: "v2 + shared=true (ignored)", apiVersion: V2, shared: true, want: false},
		{name: "v3 + shared=false", apiVersion: V3, shared: false, want: false},
		{name: "v3 + shared=true", apiVersion: V3, shared: true, want: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := &ApplicationConfig{APIVersion: tc.apiVersion, Shared: tc.shared}
			if got := cfg.IsShared(); got != tc.want {
				t.Fatalf("IsShared()=%v, want %v", got, tc.want)
			}
			// Sanity: IsV3 stays a strict schema-version check regardless
			// of the shared flag, so v3+per-user still reports IsV3 true.
			if got := cfg.IsV3(); got != (tc.apiVersion == V3) {
				t.Fatalf("IsV3()=%v, want %v", got, tc.apiVersion == V3)
			}
		})
	}
}

// TestObjectIsShared covers the K8s-object variant of IsShared, the
// runtime discriminator used by handlers, controllers and webhooks. It
// reads the AppSharedLabel and is independent of the AppApiVersionLabel
// — that label is the schema marker (v3 = both shared and per-user) and
// must NOT be used to gate shared-app behavior on its own.
func TestObjectIsShared(t *testing.T) {
	cases := []struct {
		name   string
		labels map[string]string
		want   bool
	}{
		{name: "no labels", labels: nil, want: false},
		{name: "only api-version=v3 (per-user v3)", labels: map[string]string{constants.AppApiVersionLabel: constants.AppVersionV3}, want: false},
		{name: "shared=true alone", labels: map[string]string{constants.AppSharedLabel: constants.AppSharedTrue}, want: true},
		{name: "shared=true + api-version=v3", labels: map[string]string{constants.AppApiVersionLabel: constants.AppVersionV3, constants.AppSharedLabel: constants.AppSharedTrue}, want: true},
		{name: "shared=false (literal)", labels: map[string]string{constants.AppSharedLabel: "false"}, want: false},
		{name: "shared=garbage", labels: map[string]string{constants.AppSharedLabel: "yes"}, want: false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			am := &v1alpha1.ApplicationManager{
				ObjectMeta: metav1.ObjectMeta{Labels: tc.labels},
			}
			if got := IsShared(am); got != tc.want {
				t.Fatalf("IsShared(am)=%v, want %v", got, tc.want)
			}
			// IsShared(nil) must be safe and return false, so callers
			// that lift a maybe-nil object pointer through the helper
			// don't have to add their own guard.
		})
	}
	if IsShared(nil) {
		t.Fatalf("IsShared(nil) must return false")
	}
}
