package upgrade

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/beclab/Olares/daemon/pkg/cli"
	"github.com/beclab/Olares/daemon/pkg/cluster/clusterop"
	"github.com/beclab/Olares/daemon/pkg/cluster/state"
	"github.com/beclab/Olares/daemon/pkg/commands"
	"k8s.io/klog/v2"
)

// The bootstrap stage is the one piece of an upgrade that runs on the source
// version: it is what puts the target olares-cli and olaresd on a machine, so
// nothing of the target exists there yet to run it. Everything in this file is
// therefore a contract with releases that do not exist yet — a later target
// is bootstrapped by this code as it is written today. What it may rely on is
// limited to three things, each fixed by a test:
//
//   - the request fields it reads: version, cliUrl, wizardUrl, cdn;
//   - the CDN name of the CLI archive, CLIDownloadURL;
//   - the target CLI's command line, BootstrapNodeArgs.
//
// The legacy Job patch (clusterop.legacyBootstrapScript) does the same three
// steps in shell for an olaresd too old to have this code.

// BootstrapRequest is what the bootstrap stage is told.
type BootstrapRequest struct {
	Version   string
	CliURL    string
	WizardURL string
	CDN       string
}

func bootstrapRequestOf(req clusterop.UpgradeStageRequest) BootstrapRequest {
	return BootstrapRequest{Version: req.Version, CliURL: req.CliURL, WizardURL: req.WizardURL, CDN: req.CDN}
}

// CLIDownloadURL is where a release's olares-cli archive is on a CDN.
func CLIDownloadURL(cdn, version, arch string) string {
	return fmt.Sprintf("%s/olares-cli-v%s_linux_%s.tar.gz", strings.TrimRight(cdn, "/"), version, arch)
}

// BootstrapNodeArgs is the target olares-cli's command that replaces olaresd.
func BootstrapNodeArgs(req BootstrapRequest, baseDir string) []string {
	args := []string{"upgrade", "bootstrap-node", "--version", req.Version, "--base-dir", baseDir}
	if req.CDN != "" {
		args = append(args, "--cdn-service", req.CDN)
	}
	if req.WizardURL != "" {
		args = append(args, "--url-override", req.WizardURL)
	}
	return args
}

// Seams for tests: a test must neither install binaries nor restart olaresd.
var (
	runningDaemonVersion = getCurrentDaemonVersion
	installedCLI         = installedCLIVersion
	bootstrapPhase       = runPhaseToCompletion
	runBootstrapCLI      = func(ctx context.Context, args []string) error {
		cmd := commands.NewCommandContext(ctx, cli.TERMINUS_CLI, args...)
		if _, err := cmd.Output(); err != nil {
			var exitErr *exec.ExitError
			if errors.As(err, &exitErr) && len(exitErr.Stderr) > 0 {
				return commands.NewCommandExecutionError(err, string(exitErr.Stderr))
			}
			return fmt.Errorf("olares-cli %s: %w", strings.Join(args, " "), err)
		}
		return nil
	}
)

// Bootstrap brings this machine's olares-cli and olaresd to req.Version.
//
// It very often does not return: bootstrap-node restarts olaresd, and this runs
// inside olaresd. The orchestrator then dispatches the stage again, to the
// olaresd that replaced this one, and that run finds both binaries already at
// the target and returns at once. Without that first check the stage would
// install and restart forever.
func Bootstrap(ctx context.Context, req BootstrapRequest) error {
	v, err := semver.NewVersion(req.Version)
	if err != nil {
		return fmt.Errorf("invalid target version %q: %v", req.Version, err)
	}
	if bootstrapped(v) {
		klog.Infof("olaresd and olares-cli are already %s, nothing to bootstrap", v.Original())
		return nil
	}

	cdn := req.CDN
	if cdn == "" {
		cdn = commands.OLARES_CDN_SERVICE
	}
	target := state.UpgradeTarget{Version: *v, CliURL: req.CliURL, WizardURL: req.WizardURL}
	download := &downloadCLI{
		Operation:        commands.Operation{Name: commands.DownloadCLI},
		installedVersion: installedCLI,
		cdn:              cdn,
	}
	for _, phase := range []commands.Interface{download, NewInstallCLI()} {
		if err := bootstrapPhase(ctx, phase, target); err != nil {
			return fmt.Errorf("%s: %w", phase.OperationName(), err)
		}
	}

	args := BootstrapNodeArgs(BootstrapRequest{Version: v.Original(), WizardURL: req.WizardURL, CDN: cdn},
		commands.TERMINUS_BASE_DIR)
	klog.Infof("bootstrapping olaresd to %s", v.Original())
	if err := runBootstrapCLI(ctx, args); err != nil {
		return fmt.Errorf("%s: %w", commands.BootstrapNode, err)
	}
	return nil
}

func bootstrapped(target *semver.Version) bool {
	daemon, err := runningDaemonVersion()
	if err != nil || !daemon.Equal(target) {
		return false
	}
	installed, err := installedCLI()
	return err == nil && installed.Equal(target)
}
