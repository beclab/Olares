package pipelines

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"path"
	"path/filepath"

	"github.com/pkg/errors"

	"github.com/beclab/Olares/cli/cmd/ctl/options"
	ctrl "github.com/beclab/Olares/cli/controllers"
	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/beclab/Olares/cli/pkg/core/logger"
	"github.com/beclab/Olares/cli/pkg/phase"
	"github.com/beclab/Olares/cli/pkg/phase/cluster"
)

func CliInstallTerminusPipeline(opts *options.CliTerminusInstallOptions) error {
	var terminusVersion, _ = phase.GetOlaresVersion()
	if terminusVersion != "" {
		return errors.New("Olares is already installed, please uninstall it first.")
	}

	arg := common.NewArgument()
	arg.SetBaseDir(opts.BaseDir)
	arg.SetKubeVersion(opts.KubeType)
	arg.SetOlaresVersion(opts.Version)
	arg.SetMinikubeProfile(opts.MiniKubeProfile)
	arg.SetStorage(getStorageValueFromEnv())
	arg.SetReverseProxy()
	arg.SetTokenMaxAge()
	arg.SetSwapConfig(opts.SwapConfig)
	if err := arg.SwapConfig.Validate(); err != nil {
		return err
	}
	if opts.WithJuiceFS {
		arg.WithJuiceFS = true
	}
	runtime, err := common.NewKubeRuntime(common.AllInOne, *arg)
	if err != nil {
		return fmt.Errorf("error creating runtime: %v", err)
	}

	manifest := path.Join(runtime.GetInstallerDir(), "installation.manifest")

	runtime.Arg.SetManifest(manifest)

	var p = cluster.InstallSystemPhase(runtime)
	logger.InfoInstallationProgress("Start to Install Olares ...")
	if err := p.Start(); err != nil {
		return err
	}

	if !runtime.GetSystemInfo().IsWindows() {
		if runtime.Arg.InCluster {
			if err := ctrl.UpdateStatus(runtime); err != nil {
				logger.Errorf("failed to update status: %v", err)
				return err
			}
			kkConfigPath := filepath.Join(runtime.GetWorkDir(), fmt.Sprintf("config-%s", runtime.ObjName))
			if config, err := ioutil.ReadFile(kkConfigPath); err != nil {
				logger.Errorf("failed to read kubeconfig: %v", err)
				return err
			} else {
				runtime.Kubeconfig = base64.StdEncoding.EncodeToString(config)
				if err := ctrl.UpdateKubeSphereCluster(runtime); err != nil {
					logger.Errorf("failed to update kubesphere cluster: %v", err)
					return err
				}
				if err := ctrl.SaveKubeConfig(runtime); err != nil {
					logger.Errorf("failed to save kubeconfig: %v", err)
					return err
				}
			}
		}
	}

	return nil
}
