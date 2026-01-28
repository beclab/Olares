package pipelines

import (
	"fmt"
	"path"

	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/beclab/Olares/cli/pkg/phase"
	"github.com/beclab/Olares/cli/pkg/phase/system"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

func CliInstallStoragePipeline() error {
	var terminusVersion, _ = phase.GetOlaresVersion()
	if terminusVersion != "" {
		return errors.New("Olares is already installed, please uninstall it first.")
	}

	arg := common.NewArgument()
	arg.SetBaseDir(viper.GetString(common.FlagBaseDir))
	arg.SetOlaresVersion(viper.GetString(common.FlagVersion))
	arg.SetStorage(getStorageConfig())

	runtime, err := common.NewKubeRuntime(*arg)
	if err != nil {
		return fmt.Errorf("error creating runtime: %v", err)
	}

	manifest := path.Join(runtime.GetInstallerDir(), "installation.manifest")
	runtime.Arg.SetManifest(manifest)

	return system.InstallStoragePipeline(runtime).Start()
}

func getStorageConfig() *common.Storage {
	storageType := viper.GetString(common.FlagStorageType)
	if storageType == "" {
		storageType = common.ManagedMinIO
	}
	return &common.Storage{
		StorageType:         storageType,
		StorageBucket:       viper.GetString(common.FlagS3Bucket),
		StoragePrefix:       viper.GetString(common.FlagBackupKeyPrefix),
		StorageAccessKey:    viper.GetString(common.FlagAWSAccessKeyIDSetup),
		StorageSecretKey:    viper.GetString(common.FlagAWSSecretAccessKeySetup),
		StorageToken:        viper.GetString(common.FlagAWSSessionTokenSetup),
		StorageClusterId:    viper.GetString(common.FlagClusterID),
		StorageSyncSecret:   viper.GetString(common.FlagBackupSecret),
		StorageVendor:       viper.GetString(common.FlagIsCloudVersion),
		BackupClusterBucket: viper.GetString(common.FlagBackupClusterBucket),
	}
}
