// Package version implements `olares-cli version`, the machine-readable
// counterpart to the `--version` flag.
//
// The flag's three lines of free text are load-bearing for existing parsers
// (the npm install wizard reads them to decide whether to replace a binary),
// so they are left exactly as they are. Everything an agent actually needs to
// answer "which olares-cli is this, and is it current" — the npm version the
// OS-line number hides, the channel, and whether the skills on disk came from
// this build — is reported here instead.
package version

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/cmd/ctl/skills"
	"github.com/beclab/Olares/cli/internal/release"
	skillsuite "github.com/beclab/Olares/cli/skills"
	"github.com/beclab/Olares/cli/version"
)

type report struct {
	Version          string      `json:"version"`
	NPMVersion       string      `json:"npm_version,omitempty"`
	NPMVersionSource string      `json:"npm_version_source,omitempty"`
	Channel          string      `json:"channel"`
	Upgradable       bool        `json:"upgradable"`
	RemoteOnly       bool        `json:"remote_only"`
	Path             string      `json:"path,omitempty"`
	GitCommit        string      `json:"git_commit"`
	BuildTime        string      `json:"build_time"`
	Vendor           string      `json:"vendor"`
	Go               goReport    `json:"go"`
	Skills           skillReport `json:"skills"`
}

type goReport struct {
	Version string `json:"version"`
	OS      string `json:"os"`
	Arch    string `json:"arch"`
}

// skillReport pairs what this binary carries with what is on disk. Carried
// and Installed disagreeing is the condition the staleness notice fires on;
// reporting both is what lets a caller check it without parsing that line.
type skillReport struct {
	Version          string `json:"version"`
	Digest           string `json:"digest,omitempty"`
	Store            string `json:"store,omitempty"`
	InstalledVersion string `json:"installed_version,omitempty"`
	InstalledDigest  string `json:"installed_digest,omitempty"`
	// InSync is nil when there is nothing on disk to compare, which is not
	// the same as being out of sync and must not be reported as false.
	InSync *bool `json:"in_sync"`
}

// NewVersionCommand assembles `olares-cli version`.
func NewVersionCommand() *cobra.Command {
	var output string
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Report this binary's version, channel, and skill state",
		Long: `Report which olares-cli this is, in a form a script can read.

` + "`--version`" + ` prints the Olares OS line (1.12.7). On the npm channel that
number is the same across every release inside it (1.12.7-cli.0 through
-cli.8), so it cannot answer "which published version do I have" — this can,
and also names the channel that upgrades it and whether the agent skills on
disk came from this build.`,
		Example: `  olares-cli version
  olares-cli version -o json
  olares-cli version -o json | jq -r '.npm_version // .version'`,
		Args:          cobra.NoArgs,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(output)
		},
	}
	cmd.Flags().StringVarP(&output, "output", "o", "table", "output format: table, json")
	return cmd
}

func run(output string) error {
	data := collect()
	if strings.EqualFold(strings.TrimSpace(output), "json") {
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(data)
	}
	printTable(data)
	return nil
}

func collect() report {
	install := release.Detect()
	data := report{
		Version:          install.BinaryVersion,
		NPMVersion:       install.NPMVersion,
		NPMVersionSource: install.NPMVersionSource,
		Channel:          string(install.Channel),
		Upgradable:       install.Upgradable(),
		RemoteOnly:       install.RemoteOnly,
		Path:             install.Path,
		GitCommit:        version.GitCommit,
		BuildTime:        version.BuildTime,
		Vendor:           version.VENDOR,
		Go: goReport{
			Version: runtime.Version(),
			OS:      runtime.GOOS,
			Arch:    runtime.GOARCH,
		},
		Skills: collectSkills(),
	}
	return data
}

func collectSkills() skillReport {
	data := skillReport{}
	if suite, err := skillsuite.SuiteVersion(); err == nil {
		data.Version = suite
	}
	if digest, err := skillsuite.Digest(); err == nil {
		data.Digest = digest
	}
	store := skills.StorePath()
	if store == "" {
		return data
	}
	data.Store = store
	identity, ok := skillsuite.ReadIdentity(store)
	if !ok {
		// Either nothing is installed or it was written by an olares-cli from
		// before the marker existed. Both are "cannot tell", and saying false
		// would send a caller to reinstall skills that may be correct.
		return data
	}
	data.InstalledVersion = identity.Version
	data.InstalledDigest = identity.Digest
	inSync := data.Digest != "" && identity.Digest == data.Digest
	data.InSync = &inSync
	return data
}

func printTable(data report) {
	line := func(label, value string) {
		if value == "" {
			return
		}
		fmt.Printf("%-16s %s\n", label+":", value)
	}
	line("olares-cli", data.Version)
	if data.NPMVersion != "" {
		line("npm package", fmt.Sprintf("%s  (%s)", data.NPMVersion, data.NPMVersionSource))
	}
	line("channel", channelDescription(data))
	line("path", data.Path)
	line("git commit", data.GitCommit)
	line("build time", data.BuildTime)
	line("vendor", data.Vendor)
	line("platform", fmt.Sprintf("%s/%s  %s", data.Go.OS, data.Go.Arch, data.Go.Version))
	line("skills", skillsDescription(data.Skills))
	if data.Skills.Store != "" {
		line("skills store", data.Skills.Store)
	}
}

// channelDescription says what upgrades this copy, because that is the only
// thing the channel name is for.
func channelDescription(data report) string {
	switch release.Channel(data.Channel) {
	case release.ChannelNPM:
		return "npm — `olares-cli update` moves it"
	case release.ChannelNPX:
		return "npx — temporary; the next `npx @olares/cli@<tag>` resolves the tag again"
	case release.ChannelOS:
		return "Olares OS bundle — `olares-cli upgrade` (the OS upgrade) moves it"
	case release.ChannelSource:
		return "local build — `git pull && make install` moves it"
	default:
		return "unknown — nothing here manages this copy"
	}
}

func skillsDescription(data skillReport) string {
	if data.Version == "" {
		return ""
	}
	switch {
	case data.InSync == nil:
		return data.Version + "  (nothing installed to compare)"
	case *data.InSync:
		return data.Version + "  (installed copy matches)"
	case data.InstalledVersion != "" && data.InstalledVersion != data.Version:
		return fmt.Sprintf("%s  (installed copy declares %s — run `olares-cli skills install`)",
			data.Version, data.InstalledVersion)
	default:
		return data.Version + "  (installed copy is a different build — run `olares-cli skills install`)"
	}
}
