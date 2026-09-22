// Package release answers two questions that every copy of olares-cli needs
// and none of them could previously answer about itself: which distribution
// channel put this binary here, and what version that channel calls it.
//
// The second is not a restatement of the first. version.VERSION is stamped
// with the Olares OS line (1.12.7), while the npm channel iterates inside it
// (1.12.7-cli.0 … 1.12.7-cli.8) — so an npm-installed binary reports 1.12.7
// no matter which of nine published packages it came from, and comparing
// that against an npm dist-tag is meaningless. The npm version is recovered
// from the Node shim, which knows it exactly, and from the embedded skills,
// which a release stamps with it.
package release

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/Masterminds/semver/v3"

	skillsuite "github.com/beclab/Olares/cli/skills"
	"github.com/beclab/Olares/cli/version"
)

// Package is the npm package this CLI is published as.
const Package = "@olares/cli"

// NPMVersionEnv carries the npm package version from the Node shim
// (cli/npm/bin/olares-cli.js) into the Go binary. The shim reads it from the
// package.json it was installed from, which is the only place the exact
// published version exists at runtime.
const NPMVersionEnv = "OLARES_CLI_NPM_VERSION"

// distTagsURL is the small endpoint: the full packument is ~47 KB of every
// version ever published, and all that is wanted here is two strings.
const distTagsURL = "https://registry.npmjs.org/-/package/@olares%2fcli/dist-tags"

// Channel names how this copy got onto the machine. It decides what an
// upgrade even means: three of the four are upgraded by something that is
// not olares-cli.
type Channel string

const (
	// ChannelNPM was installed by `npm install -g @olares/cli`. Upgradable
	// in place, and the only channel that is.
	ChannelNPM Channel = "npm"
	// ChannelNPX is a temporary unpack in the npx cache. There is nothing on
	// PATH to upgrade; the next `npx @olares/cli@<tag>` resolves the tag again.
	ChannelNPX Channel = "npx"
	// ChannelOS is the bundle an Olares OS release laid down at
	// /usr/local/bin/olares-cli. `olares-cli upgrade` (the OS upgrade) moves
	// it; npm must not.
	ChannelOS Channel = "os"
	// ChannelSource is a local build — `make install`, or `go build`.
	ChannelSource Channel = "source"
	// ChannelUnknown is a binary somewhere else, copied by hand or managed by
	// something this code does not know about. Treated like source: reported,
	// never written over.
	ChannelUnknown Channel = "unknown"
)

// gitDescribe matches what `git describe --tags --always --dirty` appends to
// a tag: a commit count and a g-prefixed hash, optionally -dirty. The
// Makefile stamps this, so its presence is proof of a local build even when
// the tag part looks like a release.
var gitDescribe = regexp.MustCompile(`(-\d+-g[0-9a-f]{7,})?-dirty$|-\d+-g[0-9a-f]{7,}$`)

// npmVersion matches the npm channel's own numbering, X.Y.Z-cli.N.
var npmVersion = regexp.MustCompile(`^\d+\.\d+\.\d+-cli\.\d+$`)

// placeholderSuite is what skills/stamp.py leaves in git. It has the shape of
// an npm version and is not one: every unreleased build carries it, so taking
// it for the published version would have `update` compare the newest release
// against 0.0.0 and offer an upgrade forever.
const placeholderSuite = "0.0.0-cli.0"

// Install describes the copy of olares-cli that is running.
type Install struct {
	Channel Channel `json:"channel"`
	// Path is the resolved executable. For the npm and npx channels this is
	// the vendored Go binary, not the `olares-cli` shim on PATH.
	Path string `json:"path"`
	// BinaryVersion is version.VERSION — the Olares OS line for a release
	// build, `git describe` output for a local one.
	BinaryVersion string `json:"binary_version"`
	// NPMVersion is the exact published npm version, when it can be known.
	// Empty on the OS and source channels, which are not npm packages.
	NPMVersion string `json:"npm_version,omitempty"`
	// NPMVersionSource records where NPMVersion came from, because the two
	// routes are not equally authoritative: "shim" is the package.json this
	// process was launched from, "skills" is the stamp a release left in the
	// embedded suite, which is exact on a published build and a placeholder
	// on any other.
	NPMVersionSource string `json:"npm_version_source,omitempty"`
	// RemoteOnly reports whether the host-management verbs are registered.
	RemoteOnly bool `json:"remote_only"`
}

// Upgradable reports whether `olares-cli update` can move this copy itself.
// Only the npm channel can: everything else is owned by something that would
// undo or be broken by an npm install over it.
func (i Install) Upgradable() bool { return i.Channel == ChannelNPM }

// Comparable returns the version to compare against an npm dist-tag, and
// whether there is one. The binary version is deliberately not a fallback:
// 1.12.7 is below every 1.12.7-cli.N in semver, so using it would report an
// upgrade that is already installed, every time.
func (i Install) Comparable() (string, bool) {
	if npmVersion.MatchString(i.NPMVersion) {
		return i.NPMVersion, true
	}
	return "", false
}

// Detect resolves the running binary's channel and version.
func Detect() Install {
	install := Install{
		BinaryVersion: version.VERSION,
		RemoteOnly:    os.Getenv("OLARES_CLI_REMOTE_ONLY") == "1",
	}
	if path, err := os.Executable(); err == nil {
		if resolved, err := filepath.EvalSymlinks(path); err == nil {
			path = resolved
		}
		install.Path = path
	}

	if shim := strings.TrimSpace(os.Getenv(NPMVersionEnv)); shim != "" {
		install.NPMVersion = shim
		install.NPMVersionSource = "shim"
	} else if stamped, err := skillsuite.SuiteVersion(); err == nil &&
		npmVersion.MatchString(stamped) && stamped != placeholderSuite {
		// A release stamps the suite with the version it publishes, so on the
		// npm channel this is exact. On the OS line the stamp is X.Y.Z-cli.0,
		// which also matches — harmless, since the channel check below is
		// what decides whether anything npm-shaped is done with it.
		install.NPMVersion = stamped
		install.NPMVersionSource = "skills"
	}

	install.Channel = detectChannel(install.Path, install.BinaryVersion)
	return install
}

func detectChannel(path, binaryVersion string) Channel {
	// Location first, and only for the two npm shapes. A binary inside
	// node_modules/@olares/cli/vendor got there by npm, and stays npm's to
	// replace even when a developer has dropped their own build on top of
	// it — `npm install -g` is exactly how that machine is put back.
	if path != "" {
		// Both separators are normalised explicitly rather than through
		// filepath.ToSlash, which only converts on Windows: the check would
		// then be one that can only be exercised on the platform it is for.
		slashed := strings.ReplaceAll(path, `\`, "/")
		switch {
		case strings.Contains(slashed, "/_npx/"):
			return ChannelNPX
		case strings.Contains(slashed, "/node_modules/@olares/cli/"):
			return ChannelNPM
		}
	}
	// Everywhere else a local build is recognised by its version rather than
	// its location, because `make install` puts one at /usr/local/bin/olares-cli
	// — exactly where the OS bundle lives. Calling that OS-managed would
	// hand somebody's own build to an upgrade path that overwrites it.
	if strings.Contains(binaryVersion, "development") || gitDescribe.MatchString(binaryVersion) {
		return ChannelSource
	}
	if path == "/usr/local/bin/olares-cli" {
		return ChannelOS
	}
	return ChannelUnknown
}

// DistTags reads the npm dist-tags for @olares/cli.
func DistTags(ctx context.Context) (map[string]string, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, distTagsURL, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Accept", "application/json")
	request.Header.Set("User-Agent", "olares-cli update")

	client := &http.Client{Timeout: 20 * time.Second}
	response, err := client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("reach the npm registry: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("npm registry answered HTTP %d for %s", response.StatusCode, Package)
	}
	tags := map[string]string{}
	if err := json.NewDecoder(response.Body).Decode(&tags); err != nil {
		return nil, fmt.Errorf("read the npm registry answer: %w", err)
	}
	if len(tags) == 0 {
		return nil, fmt.Errorf("npm registry returned no dist-tags for %s", Package)
	}
	return tags, nil
}

// Newer reports whether candidate is a higher version than current. Both must
// parse as semver; anything else is reported as "no", because an upgrade
// offered on the strength of a version nobody can read is worse than none.
func Newer(candidate, current string) bool {
	a, err := semver.NewVersion(strings.TrimPrefix(candidate, "v"))
	if err != nil {
		return false
	}
	b, err := semver.NewVersion(strings.TrimPrefix(current, "v"))
	if err != nil {
		return false
	}
	return a.GreaterThan(b)
}
