// Package update implements `olares-cli update`, which upgrades olares-cli
// itself.
//
// The name is deliberate and the distinction is the point. `olares-cli
// upgrade` already exists and it upgrades **Olares OS** — on a host, typing
// it to get a newer CLI starts a cluster upgrade. Nothing but a second verb
// fixes that, so this is it: `update` is the CLI, `upgrade` is the OS.
//
// Only the npm channel can be upgraded in place. The OS bundle is owned by
// the OS release, a local build by whoever built it, and an npx unpack is
// gone when the command ends. For those this verb does not act; it says what
// does move them, which is the question the user actually had.
package update

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/internal/release"
	"github.com/beclab/Olares/cli/pkg/cliutil"
)

// Actions a run can end in. A caller scripting this branches on these rather
// than on the exit code, which is 0 for every outcome that is not a failure —
// including "there is nothing I can do for this channel", which is a true
// answer rather than an error.
const (
	actionUpdated        = "updated"
	actionAlreadyCurrent = "already_current"
	actionAvailable      = "update_available"
	actionManual         = "manual_required"
)

type options struct {
	Check   bool
	Channel string
	// To names an exact version instead of a dist-tag. Deliberately not
	// spelled --version: on the root command that flag prints the version,
	// and one verb where it installs one instead is a trap.
	To      string
	Output  string
	Yes     bool
	Timeout time.Duration
}

func (o *options) isJSON() bool { return strings.EqualFold(strings.TrimSpace(o.Output), "json") }

type result struct {
	OK bool `json:"ok"`
	// Action is the outcome; see the constants above.
	Action string `json:"action"`
	// Channel is how this copy is installed, not the npm dist-tag.
	Channel string `json:"channel"`
	// Tag is the npm dist-tag the target was resolved from, empty when
	// --version named it outright.
	Tag             string            `json:"tag,omitempty"`
	CurrentVersion  string            `json:"current_version,omitempty"`
	TargetVersion   string            `json:"target_version,omitempty"`
	BinaryVersion   string            `json:"binary_version"`
	DistTags        map[string]string `json:"dist_tags,omitempty"`
	SkillsInstalled bool              `json:"skills_installed"`
	Message         string            `json:"message"`
	// Commands are the exact lines to run by hand, for the channels this
	// verb will not touch.
	Commands []string `json:"commands,omitempty"`
}

// NewUpdateCommand assembles `olares-cli update`.
func NewUpdateCommand() *cobra.Command {
	opts := &options{}
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update olares-cli itself (not Olares — that is `olares-cli upgrade`)",
		Long: `Update this olares-cli, and bring the agent skills along with it.

This is the CLI. ` + "`olares-cli upgrade`" + ` is Olares OS — a cluster upgrade on a
host, and not what you want if you are here for a newer binary.

Only an npm-installed copy can be updated in place. On an Olares OS bundle, a
local build, or an npx run, this reports what does move that copy instead of
doing something to it.

Two dist-tags, and the difference matters: ` + "`latest`" + ` is promoted by hand and
can sit well behind, ` + "`next`" + ` is every release as CI publishes it. Both are
printed so the gap is visible rather than something to find out later.`,
		Example: `  olares-cli update --check
  olares-cli update
  olares-cli update --channel next
  olares-cli update --to 1.12.7-cli.8 --yes
  olares-cli update --check -o json | jq -r .target_version`,
		Args:          cobra.NoArgs,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), opts)
		},
	}
	cmd.Flags().BoolVar(&opts.Check, "check", false, "report what an update would do, change nothing")
	cmd.Flags().StringVar(&opts.Channel, "channel", "latest", "npm dist-tag to update to: latest, next")
	cmd.Flags().StringVar(&opts.To, "to", "", "an exact version to install, instead of a dist-tag")
	cmd.Flags().StringVarP(&opts.Output, "output", "o", "table", "output format: table, json")
	cmd.Flags().BoolVarP(&opts.Yes, "yes", "y", false,
		"skip the confirmation prompt (required when stdin is not a terminal)")
	cmd.Flags().DurationVar(&opts.Timeout, "timeout", 10*time.Minute, "how long to allow the npm install to run")
	return cmd
}

func run(ctx context.Context, opts *options) error {
	if ctx == nil {
		ctx = context.Background()
	}
	install := release.Detect()

	tags, err := release.DistTags(ctx)
	if err != nil {
		return err
	}

	tag := strings.TrimSpace(opts.Channel)
	target := strings.TrimSpace(opts.To)
	if target == "" {
		resolved, ok := tags[tag]
		if !ok {
			return fmt.Errorf("npm has no dist-tag %q for %s (it has: %s)",
				tag, release.Package, strings.Join(tagNames(tags), ", "))
		}
		target = resolved
	} else {
		tag = ""
	}

	current, comparable := install.Comparable()
	res := result{
		Channel:        string(install.Channel),
		Tag:            tag,
		CurrentVersion: current,
		TargetVersion:  target,
		BinaryVersion:  install.BinaryVersion,
		DistTags:       tags,
	}

	// A copy this verb cannot move is reported, never acted on. Doing an npm
	// install over an OS bundle or somebody's `make install` is precisely the
	// damage the channel check exists to prevent.
	if !install.Upgradable() {
		res.OK = true
		res.Action = actionManual
		res.Message, res.Commands = manualRoute(install, target)
		return report(opts, res)
	}

	if comparable && !release.Newer(target, current) && opts.To == "" {
		res.OK = true
		res.Action = actionAlreadyCurrent
		if release.Newer(current, target) {
			// Not a rounding error and not rare: `latest` is promoted by hand
			// and has sat months behind what CI publishes, so anyone who ever
			// installed from `next` lands here. Saying "you are up to date"
			// would be the wrong half of the truth.
			res.Message = fmt.Sprintf("olares-cli %s is ahead of the %s tag (%s); nothing to install",
				current, tag, target)
		} else {
			res.Message = fmt.Sprintf("olares-cli %s is the newest on the %s tag", current, tag)
		}
		if opts.Check {
			return report(opts, res)
		}
		// Nothing to download, but the skills on disk can still be from a
		// different build — that is the drift this CLI already complains
		// about on every command, and an `update` that left it in place
		// would be the one command that noticed and did nothing.
		res.SkillsInstalled = installSkills(ctx, install, opts)
		return report(opts, res)
	}

	if opts.Check {
		res.OK = true
		res.Action = actionAvailable
		if comparable {
			res.Message = fmt.Sprintf("olares-cli %s → %s available on the %s tag", current, target, tag)
		} else {
			res.Message = fmt.Sprintf("%s@%s is what the %s tag points at; this copy does not report an npm version to compare",
				release.Package, target, tag)
		}
		return report(opts, res)
	}

	spec := release.Package + "@" + target
	if !opts.isJSON() {
		fmt.Fprintln(os.Stderr, tagAdvice(tags, tag))
	}
	if err := cliutil.ConfirmDestructive(os.Stderr, os.Stdin,
		fmt.Sprintf("Install %s globally with npm, replacing %s", spec, install.Path), opts.Yes); err != nil {
		return err
	}

	restore, err := prepareSelfReplace(install.Path)
	if err != nil {
		return err
	}
	if err := npmInstall(ctx, spec, opts.Timeout); err != nil {
		restore()
		return err
	}

	// The binary at install.Path is now the new one. Read the version from
	// the package.json npm just wrote rather than from our own environment:
	// OLARES_CLI_NPM_VERSION in this process still names the version being
	// replaced.
	if installed, ok := installedNPMVersion(install.Path); ok {
		res.TargetVersion = installed
	}
	res.SkillsInstalled = installSkills(ctx, install, opts)
	res.OK = true
	res.Action = actionUpdated
	if comparable {
		res.Message = fmt.Sprintf("olares-cli updated from %s to %s", current, res.TargetVersion)
	} else {
		res.Message = fmt.Sprintf("olares-cli updated to %s", res.TargetVersion)
	}
	return report(opts, res)
}

// manualRoute names what upgrades a copy this verb will not touch.
func manualRoute(install release.Install, target string) (string, []string) {
	switch install.Channel {
	case release.ChannelOS:
		return "this is the Olares OS bundle at /usr/local/bin/olares-cli, which the OS release owns",
			[]string{
				"olares-cli upgrade                                          # upgrades Olares OS, and this binary with it",
				"# or install the npm copy beside it — never over it:",
				fmt.Sprintf("npm install -g %s@%s --prefix ~/.olares-cli-npm", release.Package, target),
				`export PATH="$HOME/.olares-cli-npm/bin:$PATH"`,
				"olares-cli skills install",
			}
	case release.ChannelNPX:
		return "this is a temporary npx unpack; there is nothing on PATH to update",
			[]string{
				fmt.Sprintf("npx %s@%s install    # install it properly, CLI and skills together", release.Package, target),
				fmt.Sprintf("npx %s@%s <verb>     # or keep running it per-command", release.Package, target),
			}
	case release.ChannelSource:
		return "this is a local build; git is what moves it",
			[]string{
				"git pull && make install    # in your Olares checkout, under cli/",
				"olares-cli skills install",
			}
	default:
		return fmt.Sprintf("this copy at %s was not installed by npm, so npm cannot replace it", install.Path),
			[]string{
				fmt.Sprintf("npm install -g %s@%s", release.Package, target),
				"olares-cli skills install",
			}
	}
}

// tagAdvice says out loud when the two dist-tags disagree. `latest` is
// promoted by hand and has sat months behind `next`; a user who asked for
// latest and got a release from the summer should be told, at the moment
// they are choosing, not afterwards.
func tagAdvice(tags map[string]string, tag string) string {
	line := fmt.Sprintf("npm dist-tags: %s", strings.Join(tagPairs(tags), "  "))
	if tag != "latest" {
		return line
	}
	if hint := staleTagHint(tags); hint != "" {
		return line + "\n" + hint
	}
	return line
}

// staleTagHint is the one line that says `latest` is behind. It is not a
// judgement about which tag to use: promotion to `latest` is a manual step,
// so the gap is a fact about the release process that a user choosing a tag
// has no other way to see.
func staleTagHint(tags map[string]string) string {
	latest, hasLatest := tags["latest"]
	next, hasNext := tags["next"]
	if !hasLatest || !hasNext || !release.Newer(next, latest) {
		return ""
	}
	return fmt.Sprintf("note: `next` (%s) is ahead of `latest` (%s); `latest` is promoted by hand.\n"+
		"      Pass --channel next for the current release.", next, latest)
}

func prepareSelfReplace(path string) (func(), error) {
	// Unix replaces a running executable by unlinking it; the inode stays
	// alive for this process and npm writes a new file. Windows refuses, so
	// the running file is moved aside first and moved back if the install
	// fails.
	if runtime.GOOS != "windows" || path == "" {
		return func() {}, nil
	}
	old := path + ".old"
	_ = os.Remove(old)
	if err := os.Rename(path, old); err != nil {
		return nil, fmt.Errorf("move the running binary aside so npm can replace it: %w", err)
	}
	return func() {
		if _, err := os.Stat(path); err == nil {
			_ = os.Remove(old)
			return
		}
		_ = os.Rename(old, path)
	}, nil
}

func npmInstall(ctx context.Context, spec string, timeout time.Duration) error {
	if _, err := exec.LookPath("npm"); err != nil {
		return fmt.Errorf("npm is not on PATH, so this copy cannot be updated from here; install %s by hand", spec)
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	fmt.Fprintf(os.Stderr, "Running: npm install -g %s\n", spec)
	cmd := exec.CommandContext(ctx, "npm", "install", "-g", spec)
	var combined bytes.Buffer
	cmd.Stdout = &combined
	cmd.Stderr = &combined
	if err := cmd.Run(); err != nil {
		output := combined.String()
		fmt.Fprint(os.Stderr, output)
		if strings.Contains(output, "EACCES") {
			return fmt.Errorf("npm could not write to its global prefix: %w\n"+
				"  Move npm to a prefix you own so global installs stop needing root:\n"+
				"    npm config set prefix ~/.npm-global\n"+
				"    export PATH=\"$HOME/.npm-global/bin:$PATH\"", err)
		}
		if strings.Contains(output, "EEXIST") {
			return fmt.Errorf("npm refused to write over an existing olares-cli: %w\n"+
				"  That guard protects an Olares OS bundle. Install beside it instead:\n"+
				"    npm install -g %s --prefix ~/.olares-cli-npm", err, spec)
		}
		return fmt.Errorf("npm install -g %s failed: %w", spec, err)
	}
	return nil
}

// installSkills runs `skills install` from the binary that is now on disk.
// It must not be done in-process: this process carries the suite it was
// compiled with, which after an update is the old one, and writing that out
// would leave the machine one release behind in exactly the files an agent
// reads.
func installSkills(ctx context.Context, install release.Install, opts *options) bool {
	if install.Path == "" {
		return false
	}
	if !opts.isJSON() {
		fmt.Fprintln(os.Stderr, "Installing the agent skills this build carries ...")
	}
	cmd := exec.CommandContext(ctx, install.Path, "skills", "install")
	// Unset the inherited version: the new binary is reached directly rather
	// than through the Node shim, so the shim's value would misreport it.
	cmd.Env = append(os.Environ(), release.NPMVersionEnv+"=")
	var combined bytes.Buffer
	cmd.Stdout = &combined
	cmd.Stderr = &combined
	if err := cmd.Run(); err != nil {
		fmt.Fprint(os.Stderr, combined.String())
		fmt.Fprintf(os.Stderr, "warning: the binary was updated but the skills were not: %v\n"+
			"  Run `olares-cli skills install` yourself.\n", err)
		return false
	}
	return true
}

// installedNPMVersion reads the package.json npm just wrote. The vendored
// binary sits at <package>/vendor/olares-cli, so its grandparent is the
// package directory.
func installedNPMVersion(binary string) (string, bool) {
	if binary == "" {
		return "", false
	}
	manifest := filepath.Join(filepath.Dir(filepath.Dir(binary)), "package.json")
	source, err := os.ReadFile(manifest)
	if err != nil {
		return "", false
	}
	var parsed struct {
		Name    string `json:"name"`
		Version string `json:"version"`
	}
	if err := json.Unmarshal(source, &parsed); err != nil || parsed.Name != release.Package {
		return "", false
	}
	return parsed.Version, parsed.Version != ""
}

func report(opts *options, res result) error {
	if opts.isJSON() {
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(res)
	}
	fmt.Fprintln(os.Stderr, res.Message)
	if len(res.Commands) > 0 {
		fmt.Fprintln(os.Stderr)
		for _, line := range res.Commands {
			fmt.Fprintf(os.Stderr, "  %s\n", line)
		}
	}
	// The tags are the context for every answer this verb gives, not only the
	// ones that end in an install: "up to date" means nothing without knowing
	// which tag it was measured against, and that the other one is ahead.
	if len(res.DistTags) > 0 && res.Action != actionUpdated {
		fmt.Fprintf(os.Stderr, "\nnpm dist-tags: %s\n", strings.Join(tagPairs(res.DistTags), "  "))
		// Not repeated to someone who already asked for `next`: they have
		// made the choice the hint exists to inform.
		if res.Tag == "latest" {
			if hint := staleTagHint(res.DistTags); hint != "" {
				fmt.Fprintln(os.Stderr, hint)
			}
		}
	}
	return nil
}

func tagNames(tags map[string]string) []string {
	names := make([]string, 0, len(tags))
	for name := range tags {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func tagPairs(tags map[string]string) []string {
	pairs := make([]string, 0, len(tags))
	for _, name := range tagNames(tags) {
		pairs = append(pairs, fmt.Sprintf("%s=%s", name, tags[name]))
	}
	return pairs
}
