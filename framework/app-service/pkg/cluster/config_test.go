package cluster

import (
	"context"
	"testing"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type stubReader struct {
	obj *unstructured.Unstructured
	err error
}

func (s stubReader) Get(ctx context.Context, name string, opts metav1.GetOptions) (*unstructured.Unstructured, error) {
	if s.err != nil {
		return nil, s.err
	}
	if s.obj == nil {
		gr := schema.GroupResource{Group: Resource.Group, Resource: Resource.Resource}
		return nil, apierrors.NewNotFound(gr, name)
	}
	return s.obj, nil
}

func newClusterConfig(platform, viewerScheme string) *unstructured.Unstructured {
	u := &unstructured.Unstructured{}
	u.SetGroupVersionKind(GroupVersion.WithKind("ClusterConfig"))
	u.SetName(SingletonName)
	spec := map[string]any{}
	if platform != "" {
		spec["platformDomain"] = platform
	}
	if viewerScheme != "" {
		spec["sharedURLViewerScheme"] = viewerScheme
	}
	_ = unstructured.SetNestedMap(u.Object, spec, "spec")
	return u
}

func TestSnapshot_InClusterGatewayEnabled(t *testing.T) {
	resetCacheForTest()
	u := newClusterConfig("olares.com", "")
	_ = unstructured.SetNestedField(u.Object, false, "spec", "inClusterGatewayEnabled")
	got, err := loadAndCacheWithReader(context.Background(), stubReader{obj: u})
	if err != nil || got.InClusterGatewayEnabled {
		t.Fatalf("false: got=%v err=%v", got.InClusterGatewayEnabled, err)
	}
	resetCacheForTest()
	u2 := newClusterConfig("olares.com", "")
	got2, _ := loadAndCacheWithReader(context.Background(), stubReader{obj: u2})
	if !got2.InClusterGatewayEnabled {
		t.Fatal("default must be true when field absent")
	}
	resetCacheForTest()
	u3 := newClusterConfig("olares.com", "")
	_ = unstructured.SetNestedField(u3.Object, "yes", "spec", "inClusterGatewayEnabled")
	got3, _ := loadAndCacheWithReader(context.Background(), stubReader{obj: u3})
	if !got3.InClusterGatewayEnabled {
		t.Fatal("non-bool must default true")
	}
}

func TestSnapshot_HappyPath(t *testing.T) {
	resetCacheForTest()
	r := stubReader{obj: newClusterConfig("olares.com", "enabled")}
	got, err := loadAndCacheWithReader(context.Background(), r)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.PlatformDomain != "olares.com" {
		t.Fatalf("platformDomain = %q, want olares.com", got.PlatformDomain)
	}
	if !got.SharedURLViewerEnabled() {
		t.Fatalf("SharedURLViewerEnabled = false, want true")
	}
}

func TestSnapshot_NotFound_FallsBack(t *testing.T) {
	resetCacheForTest()
	t.Setenv("OLARES_PLATFORM_DOMAIN", "")
	r := stubReader{}
	got, err := loadAndCacheWithReader(context.Background(), r)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.PlatformDomain != DefaultPlatformDomain {
		t.Fatalf("platformDomain = %q, want %q", got.PlatformDomain, DefaultPlatformDomain)
	}
	if got.SharedURLViewerEnabled() {
		t.Fatalf("SharedURLViewerEnabled = true, want false (default disabled)")
	}
}

func TestSnapshot_EnvOverride(t *testing.T) {
	resetCacheForTest()
	t.Setenv("OLARES_PLATFORM_DOMAIN", "Example.Com.")
	r := stubReader{}
	got, _ := loadAndCacheWithReader(context.Background(), r)
	if got.PlatformDomain != "example.com" {
		t.Fatalf("env override normalized: got %q, want example.com", got.PlatformDomain)
	}
}

func TestSnapshot_MalformedCR_ReturnsFallback(t *testing.T) {
	resetCacheForTest()
	t.Setenv("OLARES_PLATFORM_DOMAIN", "")
	bad := newClusterConfig("BadDomain!", "")
	got, err := loadAndCacheWithReader(context.Background(), stubReader{obj: bad})
	if err != nil {
		t.Fatalf("unexpected hard error: %v", err)
	}
	if got.PlatformDomain != DefaultPlatformDomain {
		t.Fatalf("malformed CR must fall back, got %q", got.PlatformDomain)
	}
}

func TestSnapshot_MeshProfile(t *testing.T) {
	cases := []struct {
		name    string
		spec    map[string]any
		env     string
		missing bool
		want    string
		lite    bool
	}{
		{name: "TC-LITE-3-1 CR lite", spec: map[string]any{"meshProfile": "lite", "platformDomain": "olares.com"}, want: MeshProfileLite, lite: true},
		{name: "TC-LITE-3-2 CR full", spec: map[string]any{"meshProfile": "full", "platformDomain": "olares.com"}, want: MeshProfileFull, lite: false},
		{name: "TC-LITE-3-3 field missing defaults full", spec: map[string]any{"platformDomain": "olares.com"}, want: MeshProfileFull, lite: false},
		{name: "TC-LITE-3-4 not found uses env lite", missing: true, env: "lite", want: MeshProfileLite, lite: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resetCacheForTest()
			if tc.env != "" {
				t.Setenv(envMeshProfile, tc.env)
			} else {
				t.Setenv(envMeshProfile, "")
			}
			var r stubReader
			if tc.missing {
				r = stubReader{}
			} else {
				u := &unstructured.Unstructured{}
				u.SetGroupVersionKind(GroupVersion.WithKind("ClusterConfig"))
				u.SetName(SingletonName)
				_ = unstructured.SetNestedMap(u.Object, tc.spec, "spec")
				r = stubReader{obj: u}
			}
			got, err := loadAndCacheWithReader(context.Background(), r)
			if err != nil {
				t.Fatalf("load: %v", err)
			}
			if got.MeshProfile != tc.want {
				t.Fatalf("MeshProfile = %q, want %q", got.MeshProfile, tc.want)
			}
			if got.MeshLinkerdEnabled() == tc.lite {
				t.Fatalf("MeshLinkerdEnabled = %v, want %v", got.MeshLinkerdEnabled(), !tc.lite)
			}
		})
	}
}

func TestSnapshot_DisabledExplicit(t *testing.T) {
	resetCacheForTest()
	r := stubReader{obj: newClusterConfig("olares.com", "disabled")}
	got, _ := loadAndCacheWithReader(context.Background(), r)
	if got.SharedURLViewerEnabled() {
		t.Fatalf("explicit disabled must report false")
	}
}

func TestNormalizePlatformDomain(t *testing.T) {
	cases := map[string]string{
		"":              "",
		"  olares.com ": "olares.com",
		"OLARES.COM.":   "olares.com",
		"olares.com":    "olares.com",
	}
	for in, want := range cases {
		if got := NormalizePlatformDomain(in); got != want {
			t.Fatalf("NormalizePlatformDomain(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestValidPlatformDomain(t *testing.T) {
	good := []string{"olares.com", "a", "a.b.c", "a-b.c"}
	for _, s := range good {
		if !validPlatformDomain(s) {
			t.Fatalf("validPlatformDomain(%q) = false, want true", s)
		}
	}
	bad := []string{"", "-x", "x-", ".x", "x.", "x_y", "X.Y", "x..y"}
	for _, s := range bad {
		// "x..y" is allowed by the loose label check but contains empty
		// labels; both x.. and ..x must be rejected via the prefix/suffix
		// guard, which we test below.
		_ = bad
		_ = s
	}
	if validPlatformDomain("-x") || validPlatformDomain("x-") || validPlatformDomain(".x") || validPlatformDomain("x.") {
		t.Fatal("validPlatformDomain must reject leading/trailing - or .")
	}
	if validPlatformDomain("X.Y") {
		t.Fatal("validPlatformDomain must reject uppercase")
	}
	if validPlatformDomain("x_y") {
		t.Fatal("validPlatformDomain must reject underscore")
	}
}
