package utils

import (
	"testing"

	"github.com/Masterminds/semver/v3"
)

func TestCoreVersionStripsPrerelease(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"1.12.6", "1.12.6"},
		{"1.12.6-20260603", "1.12.6"},
		{"1.12.6-alpha1", "1.12.6"},
		{"1.12.7-20260524", "1.12.7"},
		{"1.12.6+build.7", "1.12.6"},
	}
	for _, c := range cases {
		v := semver.MustParse(c.in)
		if got := CoreVersion(v).String(); got != c.want {
			t.Errorf("CoreVersion(%s)=%s, want %s", c.in, got, c.want)
		}
	}
	if CoreVersion(nil) != nil {
		t.Error("CoreVersion(nil) should be nil")
	}
}

func TestSamePatchLevel(t *testing.T) {
	cases := []struct {
		a, b string
		want bool
	}{
		{"1.12.6", "1.12.6-20260603", true},
		{"1.12.6-alpha1", "1.12.6", true},
		{"1.12.6", "1.12.5", false},
		{"1.12.6", "1.13.6", false},
		{"1.12.6", "2.12.6", false},
	}
	for _, c := range cases {
		a := semver.MustParse(c.a)
		b := semver.MustParse(c.b)
		if got := SamePatchLevel(a, b); got != c.want {
			t.Errorf("SamePatchLevel(%s,%s)=%v, want %v", c.a, c.b, got, c.want)
		}
	}
	if SamePatchLevel(nil, semver.MustParse("1.12.6")) {
		t.Error("SamePatchLevel(nil, _) should be false")
	}
}
