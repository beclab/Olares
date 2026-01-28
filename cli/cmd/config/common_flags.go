package config

import (
	"github.com/beclab/Olares/cli/pkg/common"
	cc "github.com/beclab/Olares/cli/pkg/core/common"
	"github.com/beclab/Olares/cli/version"
)

func AddStorageFlagsBy(flagSetter CommandFlagSetter) {
	// Storage backend type: managed-minio, s3, oss, cos, minio
	flagSetter.Add(common.FlagStorageType,
		"",
		"",
		"Set storage backend type: managed-minio, s3, oss, cos, minio",
	).WithAlias(common.FlagLegacyStorageType)

	flagSetter.Add(common.FlagS3Bucket, "", "", "Object storage bucket name")
	flagSetter.Add(common.FlagBackupKeyPrefix, "", "", "Object storage key prefix for backups")
	flagSetter.Add(common.FlagAWSAccessKeyIDSetup, "", "", "Access key ID for object storage")
	flagSetter.Add(common.FlagAWSSecretAccessKeySetup, "", "", "Secret access key for object storage")
	flagSetter.Add(common.FlagAWSSessionTokenSetup, "", "", "Session token for temporary credentials")
	flagSetter.Add(common.FlagClusterID, "", "", "Cluster ID used as JuiceFS filesystem name in cloud environments")
	flagSetter.Add(common.FlagBackupSecret, "", "", "Backup sync secret for Terminus cloud")
	flagSetter.Add(common.FlagBackupClusterBucket, "", "", "Backup cluster bucket name")
	flagSetter.Add(common.FlagIsCloudVersion, "", "", "Set to true when running in Terminus cloud environment")
}

func AddVersionFlagBy(flagSetter CommandFlagSetter) {
	flagSetter.Add(common.FlagVersion,
		"v",
		version.VERSION,
		"Set Olares version, e.g., 1.10.0, 1.10.0-20241109",
	)
}

func AddBaseDirFlagBy(flagSetter CommandFlagSetter) {
	flagSetter.Add(common.FlagBaseDir,
		"b",
		"",
		"Set Olares package base dir, defaults to $HOME/"+cc.DefaultBaseDir,
	)
}

func AddMiniKubeProfileFlagBy(flagSetter CommandFlagSetter) {
	flagSetter.Add(common.FlagMiniKubeProfile,
		"p",
		"",
		"Set Minikube profile name, only in MacOS platform, defaults to "+common.MinikubeDefaultProfile,
	).WithAlias(common.FlagLegacyMiniKubeProfile)
}

func AddKubeTypeFlagBy(flagSetter CommandFlagSetter) {
	flagSetter.Add(common.FlagKubeType,
		"",
		common.K3s,
		"Set kube type, e.g., k3s or k8s",
	).WithAlias(common.FlagLegacyKubeType)
}

func AddCDNServiceFlagBy(flagSetter CommandFlagSetter) {
	flagSetter.Add(common.FlagCDNService,
		"",
		cc.DefaultOlaresCDNService,
		"Set the CDN accelerated download address in the format https://cdn.olares.cn. If not set, the default download address will be used",
	).WithEnv(common.ENV_OLARES_CDN_SERVICE)
}

func AddManifestFlagBy(flagSetter CommandFlagSetter) {
	flagSetter.Add(common.FlagManifest,
		"",
		"",
		"Set package manifest file, defaults to ${base-dir}/versions/v{version}/installation.manifest",
	)
}

func AddMasterHostFlagsBy(flagSetter CommandFlagSetter) {
	flagSetter.Add(common.FlagMasterHost, "", "", "IP address of the master node")
	flagSetter.Add(common.FlagMasterNodeName, "", "", "Name of the master node")
	flagSetter.Add(common.FlagMasterSSHUser, "", "", "Username of the master node, defaults to root")
	flagSetter.Add(common.FlagMasterSSHPassword, "", "", "Password of the master node")
	flagSetter.Add(common.FlagMasterSSHPrivateKeyPath, "", "", "Path to the SSH key to access the master node, defaults to ~/.ssh/id_rsa")
	flagSetter.Add(common.FlagMasterSSHPort, "", 0, "SSH Port of the master node, defaults to 22")
}
