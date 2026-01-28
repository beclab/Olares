package pipelines

import (
	"errors"
	"fmt"
	"path"

	"github.com/beclab/Olares/cli/pkg/core/logger"
	"github.com/beclab/Olares/cli/pkg/daemon"

	bootstrapos "github.com/beclab/Olares/cli/pkg/bootstrap/os"
	"github.com/beclab/Olares/cli/pkg/bootstrap/patch"
	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/beclab/Olares/cli/pkg/container"
	"github.com/beclab/Olares/cli/pkg/core/module"
	"github.com/beclab/Olares/cli/pkg/core/pipeline"
	"github.com/beclab/Olares/cli/pkg/images"
	"github.com/beclab/Olares/cli/pkg/manifest"
	"github.com/beclab/Olares/cli/pkg/phase"
	"github.com/beclab/Olares/cli/pkg/phase/system"
	"github.com/spf13/viper"
)

func PrepareSystemPipeline(components []string) error {
	var terminusVersion, _ = phase.GetOlaresVersion()
	if terminusVersion != "" && len(components) == 0 {
		return errors.New("Olares is already installed, please uninstall it first.")
	}

	var arg = common.NewArgument()
	arg.SetBaseDir(viper.GetString(common.FlagBaseDir))
	arg.SetKubeVersion(viper.GetString(common.FlagKubeType))
	arg.SetMinikubeProfile(viper.GetString(common.FlagMiniKubeProfile))
	arg.SetOlaresVersion(viper.GetString(common.FlagVersion))
	arg.SetStorage(getStorageConfig())

	runtime, err := common.NewKubeRuntime(*arg)
	if err != nil {
		return fmt.Errorf("error creating runtime: %w", err)
	}

	manifestPath := path.Join(runtime.GetInstallerDir(), "installation.manifest")
	runtime.Arg.SetManifest(manifestPath)

	manifestMap, err := manifest.ReadAll(manifestPath)
	if err != nil {
		return fmt.Errorf("error reading manifest: %w", err)
	}

	// if no components specified, run all
	if len(components) == 0 {
		var p = system.PrepareSystemPhase(runtime)
		if err := p.Start(); err != nil {
			return err
		}
		return nil
	}

	for _, component := range components {
		switch component {
		case "image", "images":
			p := &pipeline.Pipeline{
				Name: "Preload Container Images",
				Modules: []module.Module{
					&images.PreloadImagesModule{
						ManifestModule: manifest.ManifestModule{
							Manifest: manifestMap,
							BaseDir:  runtime.GetBaseDir(),
						},
					},
				},
				Runtime: runtime,
			}
			if err := p.Start(); err != nil {
				return fmt.Errorf("error preparing images: %w", err)
			}
		case "olaresd":
			p := &pipeline.Pipeline{
				Name: "Prepare Olaresd daemon",
				Modules: []module.Module{
					&daemon.ReplaceOlaresdBinaryModule{
						ManifestModule: manifest.ManifestModule{
							Manifest: manifestMap,
							BaseDir:  runtime.GetBaseDir(),
						},
					},
				},
				Runtime: runtime,
			}
			if err := p.Start(); err != nil {
				return fmt.Errorf("error preparing os environment: %w", err)
			}
		case "os":
			p := &pipeline.Pipeline{
				Name: "Prepare OS environment",
				Modules: []module.Module{
					&bootstrapos.PvePatchModule{Skip: !runtime.GetSystemInfo().IsPveOrPveLxc()},
					&patch.InstallDepsModule{
						ManifestModule: manifest.ManifestModule{
							Manifest: manifestMap,
							BaseDir:  runtime.GetBaseDir(),
						},
					},
					&bootstrapos.ConfigSystemModule{},
				},
				Runtime: runtime,
			}
			if err := p.Start(); err != nil {
				return fmt.Errorf("error preparing os environment: %w", err)
			}
		case "container":
			p := &pipeline.Pipeline{
				Name: "Install Container Runtime",
				Modules: []module.Module{
					&container.InstallContainerModule{
						ManifestModule: manifest.ManifestModule{
							Manifest: manifestMap,
							BaseDir:  runtime.GetBaseDir(),
						},
					},
				},
				Runtime: runtime,
			}
			if err := p.Start(); err != nil {
				return fmt.Errorf("error setting up container runtime: %w", err)
			}
		default:
			logger.Warnf("unkonwn component: %s", component)
		}
	}

	return nil
}
