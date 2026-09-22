package release

import "testing"

func TestDetectChannel(t *testing.T) {
	cases := []struct {
		name    string
		path    string
		version string
		want    Channel
	}{{
		name:    "npm global install",
		path:    "/usr/local/lib/node_modules/@olares/cli/vendor/olares-cli",
		version: "1.12.7",
		want:    ChannelNPM,
	}, {
		name:    "npx cache",
		path:    "/home/u/.npm/_npx/2f0a1b/node_modules/@olares/cli/vendor/olares-cli",
		version: "1.12.7",
		want:    ChannelNPX,
	}, {
		// A developer who drops their own build into the npm layout still has
		// an npm-managed path, and `npm install -g` is how that machine is put
		// back. Location beats version here.
		name:    "own build dropped into the npm layout",
		path:    "/usr/local/lib/node_modules/@olares/cli/vendor/olares-cli",
		version: "1.12.7-20260915-5-gdeadbeef",
		want:    ChannelNPM,
	}, {
		name:    "OS bundle",
		path:    "/usr/local/bin/olares-cli",
		version: "1.12.7",
		want:    ChannelOS,
	}, {
		// The case the version check exists for: `make install` writes to the
		// same path as the OS bundle, and calling it OS-managed would offer an
		// upgrade that overwrites somebody's own build.
		name:    "make install over the OS path",
		path:    "/usr/local/bin/olares-cli",
		version: "1.12.7-20260913-5-g220d07f74",
		want:    ChannelSource,
	}, {
		name:    "make build in a dirty checkout",
		path:    "/home/u/Olares/cli/olares-cli",
		version: "1.12.7-20260915-dirty",
		want:    ChannelSource,
	}, {
		name:    "plain go build",
		path:    "/tmp/olares-cli",
		version: "0.0.0-development",
		want:    ChannelSource,
	}, {
		name:    "copied by hand",
		path:    "/opt/tools/olares-cli",
		version: "1.12.7",
		want:    ChannelUnknown,
	}, {
		name:    "windows npm path",
		path:    `C:\Users\u\AppData\Roaming\npm\node_modules\@olares\cli\vendor\olares-cli.exe`,
		version: "1.12.7",
		want:    ChannelNPM,
	}}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := detectChannel(tc.path, tc.version); got != tc.want {
				t.Fatalf("detectChannel(%q, %q) = %q, want %q", tc.path, tc.version, got, tc.want)
			}
		})
	}
}

func TestUpgradableOnlyOnNPM(t *testing.T) {
	for _, channel := range []Channel{ChannelNPX, ChannelOS, ChannelSource, ChannelUnknown} {
		if (Install{Channel: channel}).Upgradable() {
			t.Fatalf("%q must not be upgradable in place", channel)
		}
	}
	if !(Install{Channel: ChannelNPM}).Upgradable() {
		t.Fatal("the npm channel must be upgradable in place")
	}
}

// The binary version must never stand in for the npm version. 1.12.7 sorts
// below every 1.12.7-cli.N, so accepting it would report an available upgrade
// that is already installed, on every run, forever.
func TestComparableRejectsTheBinaryVersion(t *testing.T) {
	install := Install{BinaryVersion: "1.12.7"}
	if version, ok := install.Comparable(); ok {
		t.Fatalf("Comparable() = %q, true; want no comparable version", version)
	}
	install.NPMVersion = "1.12.7-cli.8"
	version, ok := install.Comparable()
	if !ok || version != "1.12.7-cli.8" {
		t.Fatalf("Comparable() = %q, %v; want 1.12.7-cli.8, true", version, ok)
	}
}

func TestComparableRejectsMalformedNPMVersions(t *testing.T) {
	for _, version := range []string{"", "1.12.7", "1.12.7-rc.1", "v1.12.7-cli.8", "next"} {
		if got, ok := (Install{NPMVersion: version}).Comparable(); ok {
			t.Fatalf("Comparable() accepted %q as %q", version, got)
		}
	}
}

func TestNewer(t *testing.T) {
	cases := []struct {
		candidate, current string
		want               bool
	}{
		{"1.12.7-cli.8", "1.12.6-cli.2", true},
		{"1.12.6-cli.2", "1.12.7-cli.8", false},
		// Numeric prerelease identifiers compare numerically, so cli.10 is
		// above cli.9 rather than below it as a string compare would have it.
		{"1.12.7-cli.10", "1.12.7-cli.9", true},
		{"1.12.7-cli.8", "1.12.7-cli.8", false},
		// A stable release outranks every prerelease of the same version.
		{"1.12.7", "1.12.7-cli.8", true},
		// Unparseable input is never an upgrade.
		{"next", "1.12.7-cli.8", false},
		{"1.12.7-cli.8", "development", false},
	}
	for _, tc := range cases {
		if got := Newer(tc.candidate, tc.current); got != tc.want {
			t.Errorf("Newer(%q, %q) = %v, want %v", tc.candidate, tc.current, got, tc.want)
		}
	}
}

// The placeholder every unreleased build carries has the shape of an npm
// version, which is the whole reason Detect filters it by name: without that,
// a development build would compare the newest release against 0.0.0-cli.0
// and report an available upgrade on every run.
func TestPlaceholderSuiteLooksLikeAnNPMVersion(t *testing.T) {
	if !npmVersion.MatchString(placeholderSuite) {
		t.Fatal("the placeholder no longer matches npmVersion, so Detect's explicit " +
			"check for it is now dead code — remove it or fix the pattern")
	}
	if Newer(placeholderSuite, "1.12.7-cli.8") {
		t.Fatal("the placeholder outranks a real release")
	}
	if !Newer("1.12.7-cli.8", placeholderSuite) {
		t.Fatal("a real release must outrank the placeholder, which is what makes " +
			"an unfiltered placeholder report a permanent false upgrade")
	}
}
