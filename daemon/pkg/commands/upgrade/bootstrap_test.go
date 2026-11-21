package upgrade

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"testing"

	"github.com/Masterminds/semver/v3"
	"github.com/beclab/Olares/daemon/pkg/cluster/clusterop"
	"github.com/beclab/Olares/daemon/pkg/cluster/state"
	"github.com/beclab/Olares/daemon/pkg/commands"
)

// The three things a released olaresd relies on when it bootstraps a later
// version. Changing any of them breaks upgrades from every release that has
// this code, so they are spelled out literally here.
func TestBootstrapContractIsFrozen(t *testing.T) {
	got := BootstrapNodeArgs(BootstrapRequest{Version: "1.12.10", CDN: "https://cdn.olares.com", WizardURL: "https://mirror/wizard.tar.gz"}, "/olares")
	want := []string{"upgrade", "bootstrap-node", "--version", "1.12.10", "--base-dir", "/olares",
		"--cdn-service", "https://cdn.olares.com", "--url-override", "https://mirror/wizard.tar.gz"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("bootstrap-node args = %q, want %q", got, want)
	}
	got = BootstrapNodeArgs(BootstrapRequest{Version: "1.12.10"}, "/olares")
	want = []string{"upgrade", "bootstrap-node", "--version", "1.12.10", "--base-dir", "/olares"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("bootstrap-node args = %q, want %q", got, want)
	}

	if got := CLIDownloadURL("https://cdn.olares.com/", "1.12.10", "arm64"); got != "https://cdn.olares.com/olares-cli-v1.12.10_linux_arm64.tar.gz" {
		t.Errorf("CLI archive URL = %s", got)
	}

	var req clusterop.UpgradeStageRequest
	wire := `{"operationId":"op","stage":"bootstrap","version":"1.12.10","clusterId":"c",` +
		`"cliUrl":"https://mirror/cli.tar.gz","wizardUrl":"https://mirror/wizard.tar.gz","cdn":"https://cdn.olares.cn"}`
	if err := json.Unmarshal([]byte(wire), &req); err != nil {
		t.Fatal(err)
	}
	if got := bootstrapRequestOf(req); got != (BootstrapRequest{Version: "1.12.10", CliURL: "https://mirror/cli.tar.gz",
		WizardURL: "https://mirror/wizard.tar.gz", CDN: "https://cdn.olares.cn"}) {
		t.Errorf("bootstrap request read from the wire = %+v", got)
	}
}

type bootstrapFakes struct {
	daemon, cli string
	phases      []string
	cliArgs     []string
}

func fakeBootstrap(t *testing.T, daemon, cli string) *bootstrapFakes {
	t.Helper()
	f := &bootstrapFakes{daemon: daemon, cli: cli}
	prevDaemon, prevCLI, prevPhase, prevRun := runningDaemonVersion, installedCLI, bootstrapPhase, runBootstrapCLI
	t.Cleanup(func() {
		runningDaemonVersion, installedCLI, bootstrapPhase, runBootstrapCLI = prevDaemon, prevCLI, prevPhase, prevRun
	})
	parse := func(v string) (*semver.Version, error) {
		if v == "" {
			return nil, errors.New("not installed")
		}
		return semver.NewVersion(v)
	}
	runningDaemonVersion = func() (*semver.Version, error) { return parse(f.daemon) }
	installedCLI = func() (*semver.Version, error) { return parse(f.cli) }
	bootstrapPhase = func(_ context.Context, cmd commands.Interface, _ state.UpgradeTarget) error {
		f.phases = append(f.phases, string(cmd.OperationName()))
		return nil
	}
	runBootstrapCLI = func(_ context.Context, args []string) error {
		f.cliArgs = args
		return nil
	}
	return f
}

// After bootstrap-node restarts olaresd the stage is dispatched again; the new
// olaresd must answer at once rather than install and restart itself again.
func TestBootstrapDoesNothingWhenBothBinariesAreTheTarget(t *testing.T) {
	f := fakeBootstrap(t, "1.12.10", "1.12.10")
	if err := Bootstrap(context.Background(), BootstrapRequest{Version: "1.12.10"}); err != nil {
		t.Fatal(err)
	}
	if len(f.phases) != 0 || f.cliArgs != nil {
		t.Errorf("bootstrapped a node already at the target: phases %v, cli %v", f.phases, f.cliArgs)
	}
}

// A new olares-cli next to an old olaresd is not bootstrapped: olaresd still
// has to be replaced.
func TestBootstrapReplacesOlaresdWhenOnlyTheCLIIsTheTarget(t *testing.T) {
	f := fakeBootstrap(t, "1.12.9", "1.12.10")
	if err := Bootstrap(context.Background(), BootstrapRequest{Version: "1.12.10", CDN: "https://cdn.olares.cn"}); err != nil {
		t.Fatal(err)
	}
	if want := []string{string(commands.DownloadCLI), string(commands.InstallCLI)}; !reflect.DeepEqual(f.phases, want) {
		t.Errorf("phases = %v, want %v", f.phases, want)
	}
	want := BootstrapNodeArgs(BootstrapRequest{Version: "1.12.10", CDN: "https://cdn.olares.cn"}, commands.TERMINUS_BASE_DIR)
	if !reflect.DeepEqual(f.cliArgs, want) {
		t.Errorf("ran olares-cli %q, want %q", f.cliArgs, want)
	}
}

func TestBootstrapUsesThisMachinesCDNWhenTheRequestNamesNone(t *testing.T) {
	f := fakeBootstrap(t, "1.12.9", "1.12.9")
	if err := Bootstrap(context.Background(), BootstrapRequest{Version: "1.12.10"}); err != nil {
		t.Fatal(err)
	}
	want := BootstrapNodeArgs(BootstrapRequest{Version: "1.12.10", CDN: commands.OLARES_CDN_SERVICE}, commands.TERMINUS_BASE_DIR)
	if !reflect.DeepEqual(f.cliArgs, want) {
		t.Errorf("ran olares-cli %q, want %q", f.cliArgs, want)
	}
}

func TestOlaresdRunsOnlyTheStagesBeforeNotifyStart(t *testing.T) {
	prev := stagePhase
	t.Cleanup(func() { stagePhase = prev })
	var ran []string
	stagePhase = func(_ context.Context, cmd commands.Interface, _ state.UpgradeTarget) error {
		ran = append(ran, string(cmd.OperationName()))
		return nil
	}

	for stage, want := range map[string][]commands.Operations{
		clusterop.StageDownload:        {commands.DownloadWizard, commands.DownloadSpaceCheck, commands.DownloadComponent},
		clusterop.StagePrecheckNode:    {commands.ImageImportSpaceCheck},
		clusterop.StagePrecheckCluster: {commands.ClusterPreCheck},
		clusterop.StageImportImages:    {commands.ImportImages},
	} {
		ran = nil
		handled, err := RunOlaresdStage(context.Background(), clusterop.UpgradeStageRequest{Stage: stage, Version: "1.12.10"})
		if !handled || err != nil {
			t.Fatalf("%s: handled=%v err=%v", stage, handled, err)
		}
		var names []string
		for _, w := range want {
			names = append(names, string(w))
		}
		if !reflect.DeepEqual(ran, names) {
			t.Errorf("%s ran %v, want %v", stage, ran, names)
		}
	}

	for _, stage := range []string{clusterop.StageNotifyStart, "upgrade-cluster", "reboot-nodes"} {
		if handled, _ := RunOlaresdStage(context.Background(), clusterop.UpgradeStageRequest{Stage: stage, Version: "1.12.10"}); handled {
			t.Errorf("olaresd ran olares-cli stage %s", stage)
		}
	}
	if _, ok := olaresdStages[clusterop.StageBootstrap]; !ok {
		t.Error("olaresd does not run the bootstrap stage")
	}
}
