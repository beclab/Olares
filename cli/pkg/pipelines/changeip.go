package pipelines

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/beclab/Olares/cli/pkg/core/logger"
	"github.com/beclab/Olares/cli/pkg/phase"
	"github.com/beclab/Olares/cli/pkg/phase/cluster"
	"github.com/beclab/Olares/cli/pkg/utils"
	"github.com/beclab/Olares/cli/version"
	"github.com/spf13/viper"
)

func ChangeIPPipeline() error {
	var arg = common.NewArgument()
	// Normalize before querying Kubernetes or constructing any runtime state so
	// every subsequent step observes the canonical node name.
	if err := normalizeHostnameToLower(arg); err != nil {
		return err
	}

	kubeType := phase.GetKubeType()
	sysversion, _ := phase.GetOlaresVersion()
	if sysversion == "" {
		sysversion = version.VERSION
	}

	arg.SetOlaresVersion(sysversion)
	arg.SetConsoleLog("changeip.log", true)
	arg.SetKubeVersion(kubeType)
	arg.SetMinikubeProfile(viper.GetString(common.FlagMiniKubeProfile))
	arg.SetWSLDistribution(viper.GetString(common.FlagWSLDistribution))

	// Validate master host config only if it's a worker node with master host set
	if arg.MasterHost != "" {
		if err := arg.MasterHostConfig.Validate(); err != nil {
			return fmt.Errorf("invalid master host config: %w", err)
		}
	}

	runtime, err := common.NewKubeRuntime(*arg)
	if err != nil {
		return err
	}

	var p = cluster.ChangeIP(runtime)
	// TODO(ctx): plumb ctx in a follow-up; this entry point is not yet ctx-aware.
	if err := p.Start(context.Background()); err != nil {
		logger.Errorf("failed to run change ip pipeline: %v", err)
		return err
	}

	return nil
}

// normalizeHostnameToLower lowercases the OS hostname (and the in-memory
// SystemInfo used to build the runtime) when it contains uppercase characters.
// It is a no-op on macOS/Windows and when the hostname is already lowercase.
func normalizeHostnameToLower(arg *common.Argument) error {
	si := arg.SystemInfo
	if si.IsDarwin() || si.IsWindows() {
		return nil
	}
	hostname := si.GetHostname()
	if !utils.ContainsUppercase(hostname) {
		return nil
	}
	lower := strings.ToLower(hostname)
	// NOTE: this runs before the runtime is created, and the zap logger is only
	// initialized as part of that. So avoid both the logger package and
	// util.Exec (which logs) here; use fmt and exec directly.
	fmt.Printf("normalizing hostname %q to %q before change-ip\n", hostname, lower)
	// change-ip is expected to run as root (sudo), so hostnamectl works without
	// an explicit sudo prefix.
	cmd := exec.CommandContext(context.Background(), "hostnamectl", "set-hostname", lower)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to normalize hostname to %q: %w (output: %s)", lower, err, strings.TrimSpace(string(out)))
	}
	si.SetHostname(lower)
	return nil
}
