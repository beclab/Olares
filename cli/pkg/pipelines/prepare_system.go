package pipelines

import (
	"bytetrade.io/web3os/installer/pkg/core/logger"
	"bytetrade.io/web3os/installer/pkg/daemon"
	"errors"
	"fmt"
	"os"
	"path"

	"bytetrade.io/web3os/installer/cmd/ctl/options"
	bootstrapos "bytetrade.io/web3os/installer/pkg/bootstrap/os"
	"bytetrade.io/web3os/installer/pkg/bootstrap/patch"
	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/container"
	"bytetrade.io/web3os/installer/pkg/core/module"
	"bytetrade.io/web3os/installer/pkg/core/pipeline"
	"bytetrade.io/web3os/installer/pkg/images"
	"bytetrade.io/web3os/installer/pkg/manifest"
	"bytetrade.io/web3os/installer/pkg/phase"
	"bytetrade.io/web3os/installer/pkg/phase/system"
)

func PrepareSystemPipeline(opts *options.CliPrepareSystemOptions, components []string) error {
	var terminusVersion, _ = phase.GetOlaresVersion()
	if terminusVersion != "" && len(components) == 0 {
		return errors.New("Olares is already installed, please uninstall it first.")
	}

	var arg = common.NewArgument()
	arg.SetBaseDir(opts.BaseDir)
	arg.SetKubeVersion(opts.KubeType)
	arg.SetMinikubeProfile(opts.MinikubeProfile)
	arg.SetOlaresVersion(opts.Version)
	arg.SetRegistryMirrors(opts.RegistryMirrors)
	arg.SetStorage(getStorageValueFromEnv())
	arg.SetTokenMaxAge()
	arg.SetReverseProxy()

	runtime, err := common.NewKubeRuntime(common.AllInOne, *arg)
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

func getStorageValueFromEnv() *common.Storage {
	storageType := os.Getenv(common.ENV_STORAGE)
	switch storageType {
	case "":
		storageType = common.ManagedMinIO
	}

	return &common.Storage{
		StorageType:         storageType,
		StorageBucket:       os.Getenv(common.ENV_S3_BUCKET),
		StoragePrefix:       os.Getenv(common.ENV_BACKUP_KEY_PREFIX),
		StorageAccessKey:    os.Getenv(common.ENV_AWS_ACCESS_KEY_ID_SETUP),
		StorageSecretKey:    os.Getenv(common.ENV_AWS_SECRET_ACCESS_KEY_SETUP),
		StorageToken:        os.Getenv(common.ENV_AWS_SESSION_TOKEN_SETUP),
		StorageClusterId:    os.Getenv(common.ENV_CLUSTER_ID),
		StorageSyncSecret:   os.Getenv(common.ENV_BACKUP_SECRET),
		StorageVendor:       os.Getenv(common.ENV_TERMINUS_IS_CLOUD_VERSION),
		BackupClusterBucket: os.Getenv(common.ENV_BACKUP_CLUSTER_BUCKET),
	}
}
